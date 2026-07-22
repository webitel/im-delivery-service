package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	threadv1 "github.com/webitel/im-delivery-service/gen/go/thread/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/store"
)

const (
	// [REF_TTL] How long an envelope-id -> message-context mapping lives.
	// Matches the scheduler event retention: a client that has not ACKed
	// within a day will not produce a meaningful receipt anyway.
	messageRefTTL = 24 * time.Hour

	// [BATCHING] Receipts are flushed to im-thread-service either when the
	// buffer fills up or on the ticker — whichever comes first.
	statusFlushBatch    = 100
	statusFlushInterval = 300 * time.Millisecond

	// [RETRY] A failed flush is retried a few times with linear backoff so a
	// transient im-thread blip does not drop a whole batch of receipts.
	maxFlushAttempts  = 3
	flushRetryBackoff = 200 * time.Millisecond

	// [VIA] Confirmation sources reported to im-thread-service.
	viaWebSocket = "ws"
	viaPush      = "push"
)

// [THREAD_CLIENT] The im-thread MessageStatus surface the reporter needs.
type ThreadStatusClient interface {
	MarkDelivered(ctx context.Context, req *threadv1.MarkDeliveredRequest) (*threadv1.MarkStatusResponse, error)
	MarkRead(ctx context.Context, req *threadv1.MarkReadRequest) (*threadv1.MarkStatusResponse, error)
}

// [DELIVERY_CONFIRMER] Resolves an ACKed envelope into a delivery report.
type DeliveryConfirmer interface {
	// ConfirmDelivered reports that the envelope's message reached the
	// recipient via the given source (ws|push). Unknown envelopes
	// (non-message events, expired refs) are silently skipped.
	ConfirmDelivered(ctx context.Context, eid uuid.UUID, via string)
}

// [INTERFACE_GUARDS]
var (
	_ EventHandler      = (*MessageStatusReporter)(nil)
	_ DismissHandler    = (*MessageStatusReporter)(nil)
	_ DeliveryConfirmer = (*MessageStatusReporter)(nil)
)

// MessageStatusReporter ties client ACKs, successful pushes, and client
// read receipts back to per-recipient message delivery statuses in
// im-thread-service:
//
//  1. As an EventHandler it observes every fan-out envelope and remembers
//     the message context (message/thread/recipient) per envelope id.
//  2. As a DeliveryConfirmer it resolves ACKed envelope ids back to that
//     context and reports batched MarkDelivered RPCs.
//  3. As a DismissHandler it turns a client "read" frame into a batched
//     MarkRead (read-up-to) RPC.
type MessageStatusReporter struct {
	log    *slog.Logger
	refs   store.MessageRefTracker
	thread ThreadStatusClient

	delivered chan *threadv1.DeliveryReceipt
	read      chan *threadv1.ReadReceipt
	wg        sync.WaitGroup

	// [SHUTDOWN_GUARD] closed protects the queues from a send-on-closed-channel
	// panic: producers (WS pumps, push goroutines running on context.Background)
	// may outlive Close, so every enqueue checks closed under the read lock
	// while Close flips it under the write lock.
	mu     sync.RWMutex
	closed bool
	once   sync.Once
}

func NewMessageStatusReporter(
	log *slog.Logger,
	refs store.MessageRefTracker,
	thread ThreadStatusClient,
) *MessageStatusReporter {
	r := &MessageStatusReporter{
		log:       log.With("component", "message_status_reporter"),
		refs:      refs,
		thread:    thread,
		delivered: make(chan *threadv1.DeliveryReceipt, 4096),
		read:      make(chan *threadv1.ReadReceipt, 4096),
	}

	r.wg.Add(1)
	go r.flushLoop()

	return r
}

// [HANDLE] Observes fan-out events: message envelopes addressed to a
// recipient (not the sender echo) are remembered so a later ACK or read
// frame can be resolved into a status report.
func (r *MessageStatusReporter) Handle(ctx context.Context, ev event.Eventer) {
	if ev == nil || ev.IsEcho() || ev.GetKind() != event.MessageCreated {
		return
	}

	msg, ok := ev.GetPayload().(*model.Message)
	if !ok || msg == nil || msg.ID == uuid.Nil {
		return
	}

	eid, err := uuid.Parse(ev.GetID())
	if err != nil {
		return
	}

	ref := &model.EventMessageRef{
		MessageID: msg.ID,
		ThreadID:  msg.ThreadID,
		MemberID:  ev.GetUserID(),
		DomainID:  msg.DomainID,
	}

	if err := r.refs.SaveRef(ctx, eid, ref, messageRefTTL); err != nil {
		r.log.Error("REF_SAVE_FAILED", slog.String("eid", ev.GetID()), slog.Any("err", err))
	}
}

// [HANDLE_DISMISS] A client "read" frame arrives as a MessageRead dismiss.
// Resolve it back to the message context and report a read receipt.
func (r *MessageStatusReporter) HandleDismiss(ctx context.Context, ev event.Eventer) {
	if ev == nil || ev.GetKind() != event.MessageRead {
		return
	}

	eid, err := uuid.Parse(ev.GetID())
	if err != nil {
		return
	}

	r.ConfirmRead(ctx, eid, viaWebSocket)
}

// [CONFIRM_DELIVERED] Resolves the envelope back to its message context and
// enqueues a delivery receipt for the batched MarkDelivered report.
func (r *MessageStatusReporter) ConfirmDelivered(ctx context.Context, eid uuid.UUID, via string) {
	ref := r.lookup(ctx, eid)
	if ref == nil {
		return
	}

	r.enqueueDelivered(eid, &threadv1.DeliveryReceipt{
		ThreadId:    ref.ThreadID.String(),
		MessageId:   ref.MessageID.String(),
		MemberId:    ref.MemberID.String(),
		DeliveredAt: time.Now().UnixMilli(),
		Via:         via,
		DomainId:    int32(ref.DomainID),
	})
}

// [CONFIRM_READ] Resolves the envelope back to its message context and
// enqueues a read receipt for the batched MarkRead report. The resolved
// message id is the read-up-to boundary: im-thread marks it and every
// earlier unread message of the recipient in the thread as read.
func (r *MessageStatusReporter) ConfirmRead(ctx context.Context, eid uuid.UUID, via string) {
	ref := r.lookup(ctx, eid)
	if ref == nil {
		return
	}

	r.enqueueRead(eid, &threadv1.ReadReceipt{
		ThreadId:      ref.ThreadID.String(),
		MemberId:      ref.MemberID.String(),
		UpToMessageId: ref.MessageID.String(),
		ReadAt:        time.Now().UnixMilli(),
		Via:           via,
		DomainId:      int32(ref.DomainID),
	})
}

// [LOOKUP] Shared envelope -> message-context resolution. Unknown envelopes
// (system events, expired/evicted refs) resolve to nil and are skipped.
func (r *MessageStatusReporter) lookup(ctx context.Context, eid uuid.UUID) *model.EventMessageRef {
	ref, err := r.refs.GetRef(ctx, eid)
	if err != nil {
		r.log.Error("REF_LOOKUP_FAILED", slog.String("eid", eid.String()), slog.Any("err", err))

		return nil
	}

	return ref
}

func (r *MessageStatusReporter) enqueueDelivered(eid uuid.UUID, receipt *threadv1.DeliveryReceipt) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return
	}

	select {
	case r.delivered <- receipt:
	default:
		r.log.Warn("DELIVERED_QUEUE_CLOGGED", slog.String("eid", eid.String()))
	}
}

func (r *MessageStatusReporter) enqueueRead(eid uuid.UUID, receipt *threadv1.ReadReceipt) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return
	}

	select {
	case r.read <- receipt:
	default:
		r.log.Warn("READ_QUEUE_CLOGGED", slog.String("eid", eid.String()))
	}
}

// [FLUSH_LOOP] Batches delivered/read receipts and reports them to
// im-thread-service. Local channel views are set to nil once closed so the
// loop drains both queues and exits only after both are done.
func (r *MessageStatusReporter) flushLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(statusFlushInterval)
	defer ticker.Stop()

	dq := r.delivered
	rq := r.read

	dbatch := make([]*threadv1.DeliveryReceipt, 0, statusFlushBatch)
	rbatch := make([]*threadv1.ReadReceipt, 0, statusFlushBatch)

	for dq != nil || rq != nil {
		select {
		case receipt, ok := <-dq:
			if !ok {
				r.flushDelivered(dbatch)
				dbatch = dbatch[:0]
				dq = nil

				continue
			}

			dbatch = append(dbatch, receipt)
			if len(dbatch) >= statusFlushBatch {
				r.flushDelivered(dbatch)
				dbatch = dbatch[:0]
			}

		case receipt, ok := <-rq:
			if !ok {
				r.flushRead(rbatch)
				rbatch = rbatch[:0]
				rq = nil

				continue
			}

			rbatch = append(rbatch, receipt)
			if len(rbatch) >= statusFlushBatch {
				r.flushRead(rbatch)
				rbatch = rbatch[:0]
			}

		case <-ticker.C:
			if len(dbatch) > 0 {
				r.flushDelivered(dbatch)
				dbatch = dbatch[:0]
			}

			if len(rbatch) > 0 {
				r.flushRead(rbatch)
				rbatch = rbatch[:0]
			}
		}
	}
}

func (r *MessageStatusReporter) flushDelivered(batch []*threadv1.DeliveryReceipt) {
	if len(batch) == 0 {
		return
	}

	receipts := make([]*threadv1.DeliveryReceipt, len(batch))
	copy(receipts, batch)

	r.sendWithRetry("MARK_DELIVERED", len(receipts), func(ctx context.Context) error {
		_, err := r.thread.MarkDelivered(ctx, &threadv1.MarkDeliveredRequest{Receipts: receipts})

		return err
	})
}

func (r *MessageStatusReporter) flushRead(batch []*threadv1.ReadReceipt) {
	if len(batch) == 0 {
		return
	}

	receipts := make([]*threadv1.ReadReceipt, len(batch))
	copy(receipts, batch)

	r.sendWithRetry("MARK_READ", len(receipts), func(ctx context.Context) error {
		_, err := r.thread.MarkRead(ctx, &threadv1.MarkReadRequest{Receipts: receipts})

		return err
	})
}

// [SEND_WITH_RETRY] Reports a batch with bounded linear-backoff retries.
// Receipts are idempotent (im-thread enforces monotonic status transitions),
// so a resend after a transient failure is safe. A batch is dropped only
// after all attempts fail — logged loudly as data loss.
func (r *MessageStatusReporter) sendWithRetry(label string, n int, fn func(ctx context.Context) error) {
	for attempt := 1; attempt <= maxFlushAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
		err := fn(ctx)
		cancel()

		if err == nil {
			return
		}

		if attempt < maxFlushAttempts {
			r.log.Warn(label+"_RETRY",
				slog.Int("attempt", attempt),
				slog.Int("receipts", n),
				slog.Any("err", err))
			time.Sleep(time.Duration(attempt) * flushRetryBackoff)

			continue
		}

		r.log.Error(label+"_DROPPED", slog.Int("receipts", n), slog.Any("err", err))
	}
}

// [CLOSE] Stops accepting new receipts, drains the queues, and stops the
// flush loop gracefully. Idempotent.
func (r *MessageStatusReporter) Close() error {
	r.once.Do(func() {
		r.mu.Lock()
		r.closed = true
		close(r.delivered)
		close(r.read)
		r.mu.Unlock()
	})
	r.wg.Wait()

	return nil
}

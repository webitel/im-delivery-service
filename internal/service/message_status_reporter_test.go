package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	threadv1 "github.com/webitel/im-delivery-service/gen/go/thread/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type fakeRefTracker struct {
	mu   sync.Mutex
	refs map[uuid.UUID]*model.EventMessageRef
}

func newFakeRefTracker() *fakeRefTracker {
	return &fakeRefTracker{refs: map[uuid.UUID]*model.EventMessageRef{}}
}

func (f *fakeRefTracker) SaveRef(_ context.Context, eid uuid.UUID, ref *model.EventMessageRef, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refs[eid] = ref

	return nil
}

func (f *fakeRefTracker) GetRef(_ context.Context, eid uuid.UUID) (*model.EventMessageRef, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.refs[eid], nil
}

type fakeThreadStatusClient struct {
	calls     chan *threadv1.MarkDeliveredRequest
	readCalls chan *threadv1.MarkReadRequest

	mu           sync.Mutex
	deliverFails int // leading MarkDelivered calls that return a transient error
}

func newFakeThreadStatusClient() *fakeThreadStatusClient {
	return &fakeThreadStatusClient{
		calls:     make(chan *threadv1.MarkDeliveredRequest, 16),
		readCalls: make(chan *threadv1.MarkReadRequest, 16),
	}
}

func (f *fakeThreadStatusClient) MarkDelivered(_ context.Context, req *threadv1.MarkDeliveredRequest) (*threadv1.MarkStatusResponse, error) {
	f.mu.Lock()
	if f.deliverFails > 0 {
		f.deliverFails--
		f.mu.Unlock()

		return nil, errors.New("transient")
	}
	f.mu.Unlock()

	f.calls <- req

	return &threadv1.MarkStatusResponse{}, nil
}

func (f *fakeThreadStatusClient) MarkRead(_ context.Context, req *threadv1.MarkReadRequest) (*threadv1.MarkStatusResponse, error) {
	f.readCalls <- req

	return &threadv1.MarkStatusResponse{}, nil
}

func (f *fakeThreadStatusClient) wait(t *testing.T, timeout time.Duration) *threadv1.MarkDeliveredRequest {
	t.Helper()

	select {
	case req := <-f.calls:
		return req
	case <-time.After(timeout):
		t.Fatalf("expected a MarkDelivered call within %v", timeout)

		return nil
	}
}

func (f *fakeThreadStatusClient) waitRead(t *testing.T, timeout time.Duration) *threadv1.MarkReadRequest {
	t.Helper()

	select {
	case req := <-f.readCalls:
		return req
	case <-time.After(timeout):
		t.Fatalf("expected a MarkRead call within %v", timeout)

		return nil
	}
}

func newTestReporter(t *testing.T) (*MessageStatusReporter, *fakeRefTracker, *fakeThreadStatusClient) {
	t.Helper()

	refs := newFakeRefTracker()
	thread := newFakeThreadStatusClient()
	r := NewMessageStatusReporter(slog.New(slog.NewTextHandler(io.Discard, nil)), refs, thread)

	t.Cleanup(func() { _ = r.Close() })

	return r, refs, thread
}

func testMessageEvent(member uuid.UUID) event.Eventer {
	msg := &model.Message{ID: uuid.New(), ThreadID: uuid.New(), DomainID: 5}

	return event.NewMessageEvent(msg, member)
}

func TestHandle_RemembersMessageEnvelopeContext(t *testing.T) {
	r, refs, _ := newTestReporter(t)
	member := uuid.New()
	ev := testMessageEvent(member)

	r.Handle(context.Background(), ev)

	eid := uuid.MustParse(ev.GetID())
	ref, _ := refs.GetRef(context.Background(), eid)

	if ref == nil {
		t.Fatal("expected the envelope ref to be saved")
	}

	msg := ev.GetPayload().(*model.Message)

	if ref.MessageID != msg.ID || ref.ThreadID != msg.ThreadID || ref.MemberID != member || ref.DomainID != msg.DomainID {
		t.Errorf("ref context mismatch: %+v", ref)
	}
}

func TestHandle_SkipsNonTrackableEvents(t *testing.T) {
	r, refs, _ := newTestReporter(t)

	echo := &event.Envelope[*model.Message]{
		ID:      uuid.New(),
		Payload: &model.Message{ID: uuid.New()},
		Kind:    event.MessageCreated,
		Echo:    true,
	}

	wrongKind := &event.Envelope[*model.Thread]{
		ID:      uuid.New(),
		Payload: &model.Thread{},
		Kind:    event.ThreadCreated,
	}

	nilPayload := &event.Envelope[*model.Message]{
		ID:   uuid.New(),
		Kind: event.MessageCreated,
	}

	r.Handle(context.Background(), echo)
	r.Handle(context.Background(), wrongKind)
	r.Handle(context.Background(), nilPayload)
	r.Handle(context.Background(), nil)

	refs.mu.Lock()
	defer refs.mu.Unlock()

	if len(refs.refs) != 0 {
		t.Fatalf("expected no refs saved, got %d", len(refs.refs))
	}
}

func TestConfirmDelivered_UnknownEnvelopeSkipped(t *testing.T) {
	r, _, thread := newTestReporter(t)

	r.ConfirmDelivered(context.Background(), uuid.New(), viaWebSocket)

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case req := <-thread.calls:
		t.Fatalf("expected no MarkDelivered, got %d receipts", len(req.Receipts))
	default:
	}
}

func TestConfirmDelivered_FlushesFullBatchImmediately(t *testing.T) {
	r, _, thread := newTestReporter(t)

	for range statusFlushBatch {
		ev := testMessageEvent(uuid.New())
		r.Handle(context.Background(), ev)
		r.ConfirmDelivered(context.Background(), uuid.MustParse(ev.GetID()), viaWebSocket)
	}

	// Well under the 300ms ticker: the flush must be size-triggered.
	req := thread.wait(t, 200*time.Millisecond)

	if len(req.Receipts) != statusFlushBatch {
		t.Fatalf("expected %d receipts in the batch, got %d", statusFlushBatch, len(req.Receipts))
	}

	got := req.Receipts[0]

	if got.Via != viaWebSocket || got.DeliveredAt <= 0 {
		t.Errorf("receipt attrs mismatch: %+v", got)
	}
}

func TestConfirmDelivered_TickerFlushesPartialBatch(t *testing.T) {
	r, _, thread := newTestReporter(t)

	ev := testMessageEvent(uuid.New())
	r.Handle(context.Background(), ev)
	r.ConfirmDelivered(context.Background(), uuid.MustParse(ev.GetID()), viaPush)

	req := thread.wait(t, 3*statusFlushInterval)

	if len(req.Receipts) != 1 || req.Receipts[0].Via != viaPush {
		t.Fatalf("expected one via=push receipt from the ticker flush, got %+v", req.Receipts)
	}
}

func TestHandleDismiss_ReadFrameReportsMarkRead(t *testing.T) {
	r, _, thread := newTestReporter(t)
	member := uuid.New()

	// A delivered message envelope must be remembered first.
	ev := testMessageEvent(member)
	r.Handle(context.Background(), ev)
	eid := uuid.MustParse(ev.GetID())
	msg := ev.GetPayload().(*model.Message)

	// The client "read" frame arrives as a MessageRead dismiss.
	r.HandleDismiss(context.Background(), event.NewReadEvent(eid, member))

	req := thread.waitRead(t, 3*statusFlushInterval)

	if len(req.Receipts) != 1 {
		t.Fatalf("expected one read receipt, got %d", len(req.Receipts))
	}

	got := req.Receipts[0]
	if got.UpToMessageId != msg.ID.String() || got.MemberId != member.String() || got.Via != viaWebSocket || got.ReadAt <= 0 {
		t.Errorf("read receipt mismatch: %+v", got)
	}
}

func TestConfirmRead_UnknownEnvelopeSkipped(t *testing.T) {
	r, _, thread := newTestReporter(t)

	r.ConfirmRead(context.Background(), uuid.New(), viaWebSocket)

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case req := <-thread.readCalls:
		t.Fatalf("expected no MarkRead, got %d receipts", len(req.Receipts))
	default:
	}
}

func TestFlushDelivered_RetriesOnTransientFailure(t *testing.T) {
	refs := newFakeRefTracker()
	thread := newFakeThreadStatusClient()
	thread.deliverFails = 1 // first MarkDelivered fails, retry must succeed
	r := NewMessageStatusReporter(slog.New(slog.NewTextHandler(io.Discard, nil)), refs, thread)
	t.Cleanup(func() { _ = r.Close() })

	ev := testMessageEvent(uuid.New())
	r.Handle(context.Background(), ev)
	r.ConfirmDelivered(context.Background(), uuid.MustParse(ev.GetID()), viaWebSocket)

	// Ticker flush (~300ms) + one backoff (~200ms); allow generous slack.
	req := thread.wait(t, 2*time.Second)

	if len(req.Receipts) != 1 {
		t.Fatalf("expected the retried batch to carry 1 receipt, got %d", len(req.Receipts))
	}
}

func TestClose_DrainsPendingReceipts(t *testing.T) {
	refs := newFakeRefTracker()
	thread := newFakeThreadStatusClient()
	r := NewMessageStatusReporter(slog.New(slog.NewTextHandler(io.Discard, nil)), refs, thread)

	for range 3 {
		ev := testMessageEvent(uuid.New())
		r.Handle(context.Background(), ev)
		r.ConfirmDelivered(context.Background(), uuid.MustParse(ev.GetID()), viaWebSocket)
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	total := 0

	for {
		select {
		case req := <-thread.calls:
			total += len(req.Receipts)

			continue
		default:
		}

		break
	}

	if total != 3 {
		t.Fatalf("expected all 3 pending receipts flushed on close, got %d", total)
	}
}

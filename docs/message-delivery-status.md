# Message Delivery Status

Per-recipient delivery state of a message — `sent → delivered → read`, or
`failed` — and how it is reported, stored, evented, and delivered to clients.

Ownership is split:

- **im-thread-service** owns the state: it stores one status row per
  (message, recipient), applies confirmations, and publishes a status-change
  event. This is where the transition rules live.
- **im-delivery-service** (this repo) is the edge: it turns client WebSocket
  frames (`ack` / `read`) into confirmations sent to im-thread, and relays the
  status-change events back out to connected clients.
- **im-providers-service** reports confirmations for external channels
  (provider webhooks).

## The four statuses

Status is per-recipient: a message to a 3-member thread has an independent
status for each of the two recipients (the sender is not tracked).

| # | Status | Meaning |
|:-:|--------|---------|
| 1 | `sent` | Accepted by the system and stored. Initial state for every recipient. |
| 2 | `delivered` | Reached the recipient's device or channel. |
| 3 | `read` | Opened by the recipient. |
| 4 | `failed` | Delivery failed; `error` holds the provider reason. |

(`0` / `unspecified` exists in the enum but is never stored — the column is
constrained to `1..4`.)

Progress is **monotonic** by rank: `sent(1) < delivered(2) < read(3)`. A status
never moves to a lower rank. `failed` is the one off-ramp: a message can go
`sent → failed`, and later recover `failed → delivered/read` on a retry.

```
        ┌───────────────────────────┐
        ▼                           │ (retry)
sent ──▶ delivered ──▶ read         │
  │                     ▲           │
  └──▶ failed ──────────┴───────────┘
```

## State machine

Every confirmation is a monotonic upsert. The guard decides whether it actually
changes the row; if it doesn't, nothing happens (and no event is emitted).

| From | To | Reported as | Guard / notes |
|------|----|-------------|---------------|
| — | `sent` | (message creation) | Row inserted for every thread member except the sender. |
| — | `delivered` | delivery receipt | Late receipt for a message that predates status tracking. |
| `sent` | `delivered` | delivery receipt | Normal path. |
| `failed` | `delivered` | delivery receipt | Retry succeeded; clears `error` / `failed_at`. |
| —, `sent`, `delivered` | `read` | read receipt | Read implies the message arrived. |
| `failed` | `read` | read receipt | Retry read; clears `error` / `failed_at`. |
| `sent` | `failed` | failure receipt | **Only** from `sent`. A failure for an already-delivered/read message is ignored. |

What deliberately does **not** happen:

- `delivered` on a row that is already `delivered` or `read` → no-op.
- `read` on a row that is already `read` → no-op.
- `failed` on anything but `sent` → ignored (you never lose a `read` to a late
  failure webhook).
- Any receipt that would move the status backwards → no-op.

Duplicate and out-of-order receipts are therefore safe: they collapse to
no-ops and produce no event.

### Timestamps

`delivered_at`, `read_at`, `failed_at` are separate columns, each set when the
row first reaches that state. `delivered_at` / `read_at` are kept once set (a
later transition never overwrites them). `failed_at` / `error` are cleared when
a `failed` row recovers to `delivered` / `read`. Confirmation time comes from
the receipt; a zero time means "now".

## How each status is triggered

**sent** — written inside the message-save transaction: one row per thread
member except the sender. There is no status event for `sent`; the client
learns the message exists from the message itself.

**delivered** — a `MarkDelivered` receipt. Two sources:
- the recipient's own client ACKs (im-delivery, `via = ws` or `push`);
- an external provider confirms delivery (im-providers, `via = provider` /
  `bot`).

**read** — a `MarkRead` receipt with **read-up-to** semantics. One receipt names
the newest message the recipient has seen (`up_to_message_id`); im-thread marks
that message and *every earlier unread message of that recipient in the thread*
as read in one go. (Message ids are UUIDv7, so id order equals creation order.)
Messages the recipient sent themselves are skipped.

**failed** — a `MarkFailed` receipt from a provider, carrying `error` (code +
message). Applies only to a `sent` row.

## Reporting receipts (gRPC → im-thread)

im-delivery and im-providers report confirmations through the
`webitel.im.service.thread.v1.MessageStatus` service. All three take a batch and
return the number of rows actually changed.

| RPC | Receipt fields |
|-----|----------------|
| `MarkDelivered` | `thread_id`, `message_id`, `member_id`, `delivered_at`, `via`, `domain_id` |
| `MarkRead` | `thread_id`, `member_id`, `up_to_message_id`, `read_at`, `via`, `domain_id` |
| `MarkFailed` | `thread_id`, `message_id`, `member_id`, `failed_at`, `via`, `domain_id`, `error_code`, `error_message` |

`via` is the confirmation source: `ws | push | provider | bot`.

Idempotency is enforced on top of the state-machine guards:

- Within one batch, repeated `(message_id, member_id)` pairs are collapsed
  (keep first); read receipts for the same `(thread_id, member_id)` collapse to
  the single greatest `up_to_message_id`.
- Across batches, the monotonic guards make replays no-ops.

From this repo, the WS `ack` / `read` frames are turned into `MarkDelivered` /
`MarkRead` calls in `internal/service/message_status_reporter.go`; timestamps
are stamped server-side and `via` is always `ws`.

## Status-change events

When (and only when) a row actually changes, im-thread emits a
`im.message.status` domain event through its transactional outbox:

```jsonc
{
  "thread_id": "…",
  "domain_id": 1,
  "member_id": "…",          // recipient whose status changed
  "message_ids": ["…", "…"], // one or many
  "status": "delivered",     // delivered | read | failed (never "sent")
  "via": "ws",
  "error": null,             // {code, message} — only when failed
  "occurred_at": "2026-07-25T10:30:45Z",
  "participants": ["…", "…"] // all current thread members, for fan-out
}
```

Batching: changes with the same `(thread, member, status, via)` are merged into
one event with multiple `message_ids` — so a read-up-to that flips 40 messages
is a single event. Failure events are **not** batched (one event per message, to
keep each `error`).

Transport: exchange `im_message.events` (topic), routing key
`im_message.<thread-id>.message.status.v1`. im-delivery binds with
`im_message.#.message.status.v1`
(`internal/handler/amqp/listener_message_status.go`), fans the event out to the
`participants`, strips the internal fields (`domain_id`, `participants`), and
pushes it to connected clients.

## WebSocket contract (client-facing)

What a client actually sees and sends over `/im/ws`. This is the relay of the
event above plus the two frames that generate the receipts.

**Receive** — one envelope, the update under `message_status_event`:

```jsonc
{
  "id": "0198f2c1-…",          // envelope id — echo it back in ack/read
  "created_at": 1753437045123,
  "payload": {
    "message_status_event": {
      "thread_id": "3f1a…",
      "member_id": "a92b…",       // recipient whose status changed
      "message_ids": ["c73d…", "c73e…"],
      "status": "read",           // delivered | read | failed
      "via": "ws",                // ws | push | provider | bot
      "error": null,              // {code, message} — only when failed
      "occurred_at": 1753437045000  // unix ms
    }
  }
}
```

**Send** — two frames, each carrying only the envelope `id` of a frame the
client previously received. The server resolves thread/message ids from it.

```jsonc
{ "type": "ack",  "event_id": "0198f2c1-…" }  // → delivered for that message
{ "type": "read", "event_id": "0198f2c1-…" }  // → read-up-to for that message
```

Send `ack` when a message frame reaches the client, and `read` once for the
newest message the user has actually seen (it covers everything before it).
Any other frame `type` is ignored.

## Consumer rules

- **Never downgrade a displayed status.** Compare by rank
  (`sent 1 < delivered 2 < read 3`) and ignore anything lower. The backend
  already drops backward receipts; the UI should mirror that.
- **Apply the change to every `message_ids`** — read-up-to and delivery batches
  routinely carry many.
- **Key by `member_id` in group threads** to render "read by 3 of 5".
- **Be idempotent** — the same change can arrive more than once (at-least-once
  delivery).
- **Send `read` sparingly** — one frame for the newest seen message is enough.

## Where it lives in the code

im-delivery-service (this repo):

- `internal/handler/ws/pumps.go` — parses inbound `ack` / `read` frames.
- `internal/service/message_status_reporter.go` — turns them into
  `MarkDelivered` / `MarkRead` calls.
- `internal/handler/amqp/listener_message_status.go` — consumes the domain event.
- `internal/handler/marshaller/ws/` — marshals the outbound `message_status_event`.

im-thread-service (owner of the state):

- `internal/domain/model/message_status.go` — the status enum and receipts.
- `internal/store/postgres/message_status_store.go` — the transition upserts.
- `internal/service/message_status.go` — event dispatch and batching.
- `migrations/…_create_message_statuses_table.sql` — the `message_statuses` table.

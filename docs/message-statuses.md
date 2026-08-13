# Message delivery & read statuses (socket client guide)

How a client connected to the **im-delivery-service** WebSocket/stream reports and
receives message delivery/read state. Statuses are **watermark-based**: a client
acknowledges *"delivered/read up to here"* rather than acking each message.

## Model in one minute

- Every message has a per-thread **`seq`** — a small monotonic integer (1, 2, 3, …)
  scoped to its thread. (`id` is the global UUID; `seq` is the per-thread order.)
- Read/delivered state is a **watermark per (thread, member)**:
  `last_read_seq` and `last_delivered_seq`. "I read up to seq N" ⇒ everything with
  `seq ≤ N` in that thread is read.
- A message's aggregate status is **derived**, not stored per-recipient:

  | status | meaning |
  |---|---|
  | `SENT` (1) | the message exists; not yet confirmed delivered |
  | `DELIVERED` (2) | every recipient's `last_delivered_seq ≥ message.seq` |
  | `READ` (3) | every recipient's `last_read_seq ≥ message.seq` |
  | `FAILED` (4) | a delivery failure was recorded for the message (per recipient) |

- Only **failures** are stored per message×recipient (they can't be a watermark:
  "X failed but X+1 succeeded"). Everything else is watermarks.

## Identity is never sent by the client

The client **never** sends `member_id` or `domain_id`. The socket is authenticated;
the server derives *who* is acking (contact id + domain) from the session token.
The client only ever tells the server **what** it has — a `seq` (or an envelope id).

---

## 1. Connect

Open the delivery stream. On reconnect, pass the last event id you durably
processed so the server can replay anything you missed:

```jsonc
// StreamRequest
{ "last_event_id": "01920a1c-....-...." }   // omit/empty on a fresh connect
```

## 2. Receive messages

Each pushed item is a `ServerEvent` with an envelope **`id`** and a payload:

```jsonc
// ServerEvent (payload = message_event)
{
  "id": "0192abcd-1111-7aaa-...",     // ENVELOPE id — use this to ACK a live message
  "created_at": 1765000000000,
  "message_event": {
    "message": {
      "id": "0192ffff-2222-7bbb-...", // message UUID
      "thread_id": "0192aaaa-....",
      "from": { "contact_id": "...", "identity": { "name": "Alice" } },
      "text": "hello",
      "type": 1
    }
  }
}
```

> A **live** pushed message is acknowledged by its **envelope `id`** (`event_id`).
> The per-thread **`seq`** is available from message **history** (`HistoryMessage.seq`)
> and from status events (`up_to_seq`) — see §4/§5. Use `event_id` live, `seq` on reconnect.

---

## 3. Acknowledge — delivered (`ack`) and read (`read`)

Send a small JSON frame up the socket. Two watermarks, two frame types.

### Live (you just received the push): ack by envelope `event_id`

```jsonc
// "I received up to this envelope"
{ "type": "ack",  "event_id": "0192abcd-1111-7aaa-..." }

// "I read up to this envelope" (user opened the chat)
{ "type": "read", "event_id": "0192abcd-1111-7aaa-..." }
```

The server resolves `event_id → message` via a 24h Redis reference, derives its
`seq`, and advances your `last_delivered_seq` / `last_read_seq`.

### Reconnect / from history: ack by `thread_id` + `seq`

After a reconnect (the 24h envelope reference may be gone), load history, take the
newest `seq` you hold, and ack by seq directly — no envelope needed:

```jsonc
// "In this thread I have received everything up to seq 42"
{ "type": "ack",  "thread_id": "0192aaaa-....", "seq": 42 }

// "In this thread I have read everything up to seq 42"
{ "type": "read", "thread_id": "0192aaaa-....", "seq": 42 }
```

### Rules (both frame types)

- **Watermark**: send the *newest* seq/envelope you have — never per message.
- **Idempotent**: re-sending the same (or an older) ack after a reconnect is a
  no-op; the server keeps the maximum. Nothing bad happens if you over-send.
- **Monotonic**: read and delivered are separate horizons and only move forward.
  A stale `ack` can never pull back a newer `read`.
- **Toggle off is empty**: there is no "unread" — you simply never advance past it.

---

## 4. `seq` — where the client gets it

| source | field | when |
|---|---|---|
| Message **history** | `HistoryMessage.seq` | paging a thread; the primary source for reconnect acks |
| **Status events** (below) | `MessageStatusEvent.up_to_seq` | to advance the peer's horizon locally |
| Live push | — (not carried; use `event_id`) | ack live messages by envelope id |

---

## 5. Receive status of the *other* participants

When someone else delivers/reads your messages, the server pushes a
`MessageStatusEvent` (one per connected participant; you never get your own):

```jsonc
// ServerEvent (payload = message_status_event)
{
  "id": "0192dddd-....",
  "message_status_event": {
    "thread_id": "0192aaaa-....",
    "member_id": "<the peer who advanced>",
    "status": 3,                    // 3 = READ, 2 = DELIVERED, 4 = FAILED
    "up_to_seq": 42,                // advance "read/delivered by peer" to seq 42 in O(1)
    "message_ids": ["...", "..."],  // also listed for older clients
    "occurred_at": 1765000000123
  }
}
```

The client advances a local "peer read/delivered up to seq 42" and re-renders the
check marks for every message with `seq ≤ 42`. A no-op ack on the server side
produces **no** event — you never receive a spurious/regressing status.

## 6. Failures

If a message can't be delivered to a recipient (e.g. an external provider rejects
it), its aggregate status becomes `FAILED (4)` and a `MessageStatusEvent` with
`status = 4` (and `error` details) is pushed. Failures are the only per-message
per-recipient state the server stores.

---

## 7. Worked example

Alice and Bob in thread `T` (seqs 40, 41, 42 are Alice's last three messages).

1. Server pushes Alice's msg (seq 42) to Bob as `ServerEvent{ id: E, message_event }`.
2. Bob's client received it → `{ "type": "ack", "event_id": "E" }`.
   → server: Bob `last_delivered_seq = 42`; Alice gets `MessageStatusEvent{ member_id: Bob, status: 2, up_to_seq: 42 }` → Alice shows **delivered** for 40–42.
3. Bob opens the chat → `{ "type": "read", "event_id": "E" }`.
   → server: Bob `last_read_seq = 42`; Alice gets `MessageStatusEvent{ member_id: Bob, status: 3, up_to_seq: 42 }` → Alice shows **read** for 40–42.
4. Bob's socket drops for a day, reconnects, loads history (newest `seq = 45`) →
   `{ "type": "read", "thread_id": "T", "seq": 45 }` → horizon jumps to 45; earlier
   acks in between were unnecessary. Fully self-healing, no envelope needed.

---

## Server-side (for reference)

- `im_message.messages.seq` — per-thread seq, assigned on insert from
  `im_thread.thread.last_seq` (row-locked → unique, monotonic).
- `im_thread.thread_dialog.last_read_seq` / `last_delivered_seq` — the watermarks
  (advanced monotonically; `read` also recomputes `unread_count`).
- `im_message.message_errors` — the only per message×recipient table (failures).
- `im_thread.v_messages.delivery_status` — derived from the watermarks
  (`READ if last_read_seq ≥ m.seq`, else `DELIVERED if last_delivered_seq ≥ m.seq`,
  else `SENT`; `FAILED` from `message_errors` overrides). Aggregate is
  all-failed ⇒ FAILED, else the minimum non-failed across recipients.
- Delivery is at-least-once; acks are idempotent + monotonic, so redelivery and
  reconnects are safe. The envelope→message reference lives 24h in Redis; past
  that, the client re-acks by `seq` from history.

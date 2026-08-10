package model

import (
	"testing"

	"github.com/google/uuid"
)

type fakeEvent struct {
	id   string
	uid  uuid.UUID
	kind string
	meta map[string]string
}

func (f fakeEvent) GetID() string                  { return f.id }
func (f fakeEvent) GetUserID() uuid.UUID           { return f.uid }
func (f fakeEvent) GetKindName() string            { return f.kind }
func (f fakeEvent) GetMetadata() map[string]string { return f.meta }

// The mobile client requires a non-empty data block with chat.id, message.id
// and type; an omitted map makes it reject the push, so all three must be set.
func TestFillFromEventPopulatesData(t *testing.T) {
	ev := fakeEvent{
		id:   uuid.NewString(),
		uid:  uuid.New(),
		kind: "MessageCreated",
		meta: map[string]string{
			"sender_name": "Ihor",
			"text":        "hi",
			"chat.id":     "chat-123",
			"message.id":  "msg-456",
		},
	}

	var req PushRequest
	req.FillFromEvent(ev)

	for _, key := range []string{"type", "chat.id", "message.id"} {
		if _, ok := req.Data[key]; !ok {
			t.Fatalf("data missing required key %q: %#v", key, req.Data)
		}
	}

	if req.Data["type"] != "UpdateNewMessage" {
		t.Fatalf("type = %q, want UpdateNewMessage", req.Data["type"])
	}

	if req.Data["chat.id"] != "chat-123" || req.Data["message.id"] != "msg-456" {
		t.Fatalf("routing keys not mapped: %#v", req.Data)
	}
}

// Even when metadata lacks routing keys, the fields must still be present
// (empty string) so strict deserializers do not fail on a missing field.
func TestFillFromEventDataKeysAlwaysPresent(t *testing.T) {
	ev := fakeEvent{id: uuid.NewString(), uid: uuid.New(), kind: "SomethingElse"}

	var req PushRequest
	req.FillFromEvent(ev)

	for _, key := range []string{"type", "chat.id", "message.id"} {
		if _, ok := req.Data[key]; !ok {
			t.Fatalf("data missing required key %q: %#v", key, req.Data)
		}
	}

	if req.Data["type"] != "SomethingElse" {
		t.Fatalf("type = %q, want passthrough kind", req.Data["type"])
	}
}

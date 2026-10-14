package payload

import (
	"encoding/json"
	"testing"
)

// im-thread-service stamps seq and updates_cursor on the event; both reach the domain model.
func TestMessageCreatedV1_SeqAndCursor(t *testing.T) {
	var in MessageCreatedV1
	if err := json.Unmarshal([]byte(`{"seq":7,"updates_cursor":"99.1790000000000"}`), &in); err != nil {
		t.Fatal(err)
	}

	got := in.ToDomain()
	if got.Seq != 7 || got.UpdatesCursor != "99.1790000000000" {
		t.Fatalf("message = seq %d cursor %q", got.Seq, got.UpdatesCursor)
	}
}

func TestMessageStatusV1_UnreadCount(t *testing.T) {
	var in MessageStatusV1
	if err := json.Unmarshal([]byte(`{"status":"read","up_to_seq":121,"unread_count":0}`), &in); err != nil {
		t.Fatal(err)
	}

	got := in.ToDomain()
	if got.UnreadCount == nil || *got.UnreadCount != 0 {
		t.Fatalf("unread_count = %v, want 0", got.UnreadCount)
	}
}

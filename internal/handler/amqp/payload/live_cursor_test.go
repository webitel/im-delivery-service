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

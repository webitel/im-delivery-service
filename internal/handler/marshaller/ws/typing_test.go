package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMapTypingFromMatchesSenderShape guards that the typing `from` is marshaled
// with the exact same nested shape as a message `sender` (WSPeer: id + contact +
// role), so clients can parse both with one type.
func TestMapTypingFromMatchesSenderShape(t *testing.T) {
	peer := model.Peer{
		ID:          uuid.New(),
		Type:        model.PeerUser,
		ContactType: "webitel",
		Sub:         "10",
		Issuer:      "webitel",
		Name:        "Ihor Ihor",
		MemberID:    uuid.NewString(),
		Role:        1, // ROLE_MEMBER
	}

	typing := &model.Typing{
		ThreadID:    uuid.NewString(),
		TimeoutMs:   12000,
		From:        peer,
		PreviewText: "Test text",
	}

	// Reference sender shape produced for a message.
	wantSender := mapPeer(&peer)
	gotFrom := mapTyping(typing).From

	wantJSON, err := json.Marshal(wantSender)
	if err != nil {
		t.Fatalf("marshal sender: %v", err)
	}

	gotJSON, err := json.Marshal(gotFrom)
	if err != nil {
		t.Fatalf("marshal typing from: %v", err)
	}

	if string(wantJSON) != string(gotJSON) {
		t.Fatalf("typing `from` shape differs from message `sender`:\n from   = %s\n sender = %s", gotJSON, wantJSON)
	}
}

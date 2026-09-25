package grpcmarshaller

import (
	"testing"

	"github.com/google/uuid"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func TestMarshal_Resync(t *testing.T) {
	ev := event.NewSystemEvent(uuid.New(), event.Resync, &model.ResyncPayload{})

	got, err := New().Marshal(ev, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := got.(*impb.ServerEvent).GetPayload().(*impb.ServerEvent_ResyncEvent); !ok {
		t.Fatalf("payload = %T, want resync_event", got.(*impb.ServerEvent).GetPayload())
	}
}

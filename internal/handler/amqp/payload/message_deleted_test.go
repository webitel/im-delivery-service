package payload

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestMessageDeletedV1_ToDomain_ReplacedBy(t *testing.T) {
	replacement := uuid.New()

	tests := []struct {
		name string
		raw  string
		want *uuid.UUID
	}{
		{"plain delete", `{"message_id":"` + uuid.NewString() + `"}`, nil},
		{"replaced by retry", `{"message_id":"` + uuid.NewString() + `","replaced_by":"` + replacement.String() + `"}`, &replacement},
		{"malformed id is ignored", `{"message_id":"` + uuid.NewString() + `","replaced_by":"not-a-uuid"}`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in MessageDeletedV1
			if err := json.Unmarshal([]byte(tt.raw), &in); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := in.ToDomain().ReplacedBy

			switch {
			case tt.want == nil && got != nil:
				t.Errorf("ReplacedBy = %s, want nil", got)
			case tt.want != nil && (got == nil || *got != *tt.want):
				t.Errorf("ReplacedBy = %v, want %s", got, tt.want)
			}
		})
	}
}

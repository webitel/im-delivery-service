package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

type fakeCursors struct {
	cursor string
	err    error
}

func (f fakeCursors) UpdatesCursor(context.Context, string) (string, error) { return f.cursor, f.err }

func TestConnectedUpdatesCursor(t *testing.T) {
	log := slog.New(slog.DiscardHandler)

	if got := ConnectedUpdatesCursor(context.Background(), fakeCursors{cursor: "94770179"}, "c", log); got != "94770179" {
		t.Errorf("cursor = %q, want 94770179", got)
	}

	// A thread-service blip must not fail the connection: the client falls back to GetUpdates.
	if got := ConnectedUpdatesCursor(context.Background(), fakeCursors{err: errors.New("down")}, "c", log); got != "" {
		t.Errorf("cursor on error = %q, want empty", got)
	}

	if got := ConnectedUpdatesCursor(context.Background(), nil, "c", log); got != "" {
		t.Errorf("cursor without client = %q, want empty", got)
	}
}

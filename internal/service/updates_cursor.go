package service

import (
	"context"
	"log/slog"
	"time"
)

// UpdatesCursors reads a contact's GetUpdates cursor from im-thread-service.
type UpdatesCursors interface {
	UpdatesCursor(ctx context.Context, contactID string) (string, error)
}

const updatesCursorTimeout = 2 * time.Second

// ConnectedUpdatesCursor is the cursor sent on the connected event. It is read after the
// connection is attached, so a change past it arrives live. Empty when unavailable: the
// client then falls back to GetUpdates with its saved cursor.
func ConnectedUpdatesCursor(ctx context.Context, cursors UpdatesCursors, contactID string, log *slog.Logger) string {
	if cursors == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, updatesCursorTimeout)
	defer cancel()

	cursor, err := cursors.UpdatesCursor(ctx, contactID)
	if err != nil {
		log.Warn("updates cursor unavailable for connected event", slog.Any("err", err))

		return ""
	}

	return cursor
}

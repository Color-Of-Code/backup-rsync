package testutil

import (
	"io"
	"log/slog"
)

// NewTestLogger creates a *slog.Logger that writes to w without timestamps,
// suitable for test assertions.
func NewTestLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
}

package deterministic

import (
	"context"
	"log/slog"
	"time"
)

// SlogHandler wraps a slog.Handler and replaces timestamps with deterministic values
// from a NowFunc.
type SlogHandler struct {
	handler slog.Handler
	nowFunc func() time.Time
}

var _ slog.Handler = (*SlogHandler)(nil)

// NewSlogHandler creates a new handler that uses the given NowFunc to generate
// deterministic timestamps for all log records.
func NewSlogHandler(handler slog.Handler, nowFunc func() time.Time) *SlogHandler {
	return &SlogHandler{
		handler: handler,
		nowFunc: nowFunc,
	}
}

// Enabled delegates to the underlying handler.
func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle intercepts the record, replaces its timestamp with a deterministic one, and delegates.
func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Time = h.nowFunc()
	return h.handler.Handle(ctx, r)
}

// WithAttrs delegates to the underlying handler and wraps the result.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SlogHandler{
		handler: h.handler.WithAttrs(attrs),
		nowFunc: h.nowFunc,
	}
}

// WithGroup delegates to the underlying handler and wraps the result.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return &SlogHandler{
		handler: h.handler.WithGroup(name),
		nowFunc: h.nowFunc,
	}
}

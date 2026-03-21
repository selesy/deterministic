package deterministic

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"
)

var _ slog.Handler = (*SlogRecorder)(nil)

// SlogRecorder is a slog.Handler that records all log records for later retrieval.
type SlogRecorder struct {
	records *[]slog.Record
	attrs   []slog.Attr
	groups  []string
}

// NewSlogRecorder creates a new SlogRecorder.
func NewSlogRecorder() *SlogRecorder {
	return &SlogRecorder{
		records: &[]slog.Record{},
	}
}

// Enabled returns true, indicating all log levels are enabled.
func (h *SlogRecorder) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// Get retrieves the record at the specified index.
func (h *SlogRecorder) Get(index int) (slog.Record, error) {
	return (*h.records)[index], nil
}

func (h *SlogRecorder) prefixAttrs(attrs []slog.Attr) []slog.Attr {
	if len(h.groups) == 0 {
		return attrs
	}
	prefix := strings.Join(h.groups, ".") + "."

	res := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		res[i] = a
		res[i].Key = prefix + a.Key
	}
	return res
}

// Handle records the log record with all attributes and group prefixes applied.
func (h *SlogRecorder) Handle(ctx context.Context, record slog.Record) error {
	newRec := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

	newRec.AddAttrs(h.attrs...)

	var rAttrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		rAttrs = append(rAttrs, a)
		return true
	})

	newRec.AddAttrs(h.prefixAttrs(rAttrs)...)

	*h.records = append(*h.records, newRec)

	return nil
}

// Len returns the number of recorded log records.
func (h *SlogRecorder) Len() int {
	return len(*h.records)
}

// WithAttrs returns a new handler with the given attributes added to all records.
func (h *SlogRecorder) WithAttrs(attrs []slog.Attr) slog.Handler {
	hand := h.clone()
	hand.attrs = append(hand.attrs, hand.prefixAttrs(attrs)...)

	return hand
}

// WithGroup returns a new handler with the given group name.
func (h *SlogRecorder) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	hand := h.clone()
	hand.groups = append(hand.groups, name)

	return hand
}

func (h *SlogRecorder) clone() *SlogRecorder {
	return &SlogRecorder{
		records: h.records,
		attrs:   slices.Clone(h.attrs),
		groups:  slices.Clone(h.groups),
	}
}

// SlogHandler wraps a slog.Handler and replaces timestamps with deterministic values
// from a NowFunc.
type SlogHandler struct {
	handler slog.Handler
	nowFunc func() time.Time
}

var _ slog.Handler = (*SlogHandler)(nil)

// NewSlogHandler creates a new handler that uses a default NowFunc to generate
// deterministic timestamps for all log records.
func NewSlogHandler(handler slog.Handler) *SlogHandler {
	return &SlogHandler{
		handler: handler,
		nowFunc: NowFunc(),
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

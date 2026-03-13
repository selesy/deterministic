package deterministic_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"
)

func TestNewSlogHandler_ReturnsHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)
	if h == nil {
		t.Error("NewSlogHandler returned nil")
	}
}

func TestSlogHandler_IsHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	var _ slog.Handler = deterministic.NewSlogHandler(underlying, nf)
}

func TestSlogHandler_ReplaceTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)
	ctx := context.Background()

	// Create a record with any time
	anyTime := nf() // Use one deterministic time
	r := slog.NewRecord(anyTime, slog.LevelInfo, "test message", 0)

	// When we handle it, the time should be replaced with next deterministic time
	err := h.Handle(ctx, r)

	if err != nil {
		t.Errorf("Handle returned error: %v", err)
	}

	// Verify timestamp was replaced with deterministic one
	output := buf.String()
	if !strings.Contains(output, "2006-01-02") {
		t.Errorf("output does not contain deterministic timestamp, got: %s", output)
	}
}

func TestSlogHandler_ConsecutiveCallsHaveDifferentTimestamps(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)
	ctx := context.Background()

	dummyTime := nf() // Get a time but discard for logging
	r1 := slog.NewRecord(dummyTime, slog.LevelInfo, "first", 0)
	r2 := slog.NewRecord(dummyTime, slog.LevelInfo, "second", 0)

	require.NoError(t, h.Handle(ctx, r1))
	require.NoError(t, h.Handle(ctx, r2))

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 log lines, got %d", len(lines))
	}

	// Extract timestamps from both lines
	ts1 := extractTime(lines[0])
	ts2 := extractTime(lines[1])

	if ts1 == "" || ts2 == "" {
		t.Fatalf("could not extract timestamps from output")
	}

	if ts1 == ts2 {
		t.Errorf("expected different timestamps, but got same: %s", ts1)
	}
}

func TestSlogHandler_Enabled(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)
	ctx := context.Background()

	// Both should return the same result
	if h.Enabled(ctx, slog.LevelInfo) != underlying.Enabled(ctx, slog.LevelInfo) {
		t.Error("Enabled returned different results")
	}
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)

	attrs := []slog.Attr{
		slog.String("key", "value"),
	}

	newH := h.WithAttrs(attrs)
	if newH == nil {
		t.Error("WithAttrs returned nil")
	}

	// Verify it's still a SlogHandler
	if _, ok := newH.(*deterministic.SlogHandler); !ok {
		t.Error("WithAttrs did not return a SlogHandler")
	}
}

func TestSlogHandler_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying, nf)

	newH := h.WithGroup("mygroup")
	if newH == nil {
		t.Error("WithGroup returned nil")
	}

	// Verify it's still a SlogHandler
	if _, ok := newH.(*deterministic.SlogHandler); !ok {
		t.Error("WithGroup did not return a SlogHandler")
	}
}

func TestSlogHandler_DifferentInstances(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	underlying1 := slog.NewTextHandler(buf1, nil)
	underlying2 := slog.NewTextHandler(buf2, nil)
	nf1 := deterministic.NowFunc()
	nf2 := deterministic.NowFunc()

	h1 := deterministic.NewSlogHandler(underlying1, nf1)
	h2 := deterministic.NewSlogHandler(underlying2, nf2)

	ctx := context.Background()
	dummyTime := nf1() // Advance nf1's internal state
	r := slog.NewRecord(dummyTime, slog.LevelInfo, "test", 0)

	require.NoError(t, h1.Handle(ctx, r))

	// nf2 is at the same initial state as nf1 started, so its first call gives the same time
	r = slog.NewRecord(dummyTime, slog.LevelInfo, "test", 0)
	require.NoError(t, h2.Handle(ctx, r))

	ts1 := extractTime(buf1.String())
	ts2 := extractTime(buf2.String())

	// Both handlers should have called their respective NowFunc once, so they get different times
	// (h1's second call, h2's first call)
	if ts1 == ts2 {
		t.Errorf("expected different timestamps for different sequences, but got same: %s", ts1)
	}
}

func extractTime(line string) string {
	// Extract time from slog text format: "time=2006-01-02T15:04:05.000Z07:00"
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasPrefix(part, "time=") {
			return strings.TrimPrefix(part, "time=")
		}
	}
	return ""
}

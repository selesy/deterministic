package deterministic_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"
)

func TestSlogRecorder(t *testing.T) {
	t.Parallel()

	newLoggerAndContent := func(h slog.Handler) {
		logger := slog.New(h)
		logger.Info("plain", slog.String("a", "a"), slog.String("b", "b"))
		logger = logger.With(slog.String("c", "c"), slog.String("d", "d"))
		logger.Info("with attrs", slog.String("e", "e"), slog.String("f", "f"))
		logger = logger.WithGroup("1")
		logger.Info("group1", slog.String("g", "g"), slog.String("h", "h"))
		logger = logger.With(slog.String("i", "i"), slog.String("j", "j"))
		logger.Info("group1 with attrs", slog.String("k", "k"), slog.String("l", "l"))
		logger = logger.WithGroup("2")
		logger.Info("group2", slog.String("m", "m"), slog.String("n", "n"))
		logger = logger.With(slog.String("o", "o"), slog.String("p", "p"))
		logger.Info("group2 with attrs", slog.String("q", "q"), slog.String("r", "r"))
	}

	textBuf := &bytes.Buffer{}
	var textHandler slog.Handler
	textHandler = slog.NewTextHandler(textBuf, nil)
	textHandler = deterministic.NewSlogHandler(textHandler)
	newLoggerAndContent(textHandler)

	wrappedRecorder := deterministic.NewSlogRecorder()
	var recHandler slog.Handler
	recHandler = wrappedRecorder
	recHandler = deterministic.NewSlogHandler(recHandler)
	newLoggerAndContent(recHandler)

	recBuf := &bytes.Buffer{}
	for i := range wrappedRecorder.Len() {
		rec, err := wrappedRecorder.Get(i)
		require.NoError(t, err)

		ts := rec.Time.Format("2006-01-02T15:04:05.000Z07:00")

		msg := rec.Message
		if strings.ContainsAny(msg, " \t\n") {
			msg = fmt.Sprintf("%q", msg)
		}
		fmt.Fprintf(recBuf, "time=%s level=%s msg=%s", ts, rec.Level, msg)
		rec.Attrs(func(attr slog.Attr) bool {
			fmt.Fprintf(recBuf, " %s=%s", attr.Key, attr.Value)

			return true
		})

		fmt.Fprintln(recBuf)
	}

	assert.Equal(t, textBuf.String(), recBuf.String())
}

func TestNewSlogHandler_ReturnsHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)

	h := deterministic.NewSlogHandler(underlying)
	if h == nil {
		t.Error("NewSlogHandler returned nil")
	}
}

func TestSlogHandler_IsHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)

	var _ slog.Handler = deterministic.NewSlogHandler(underlying)
}

func TestSlogHandler_ReplaceTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)
	nf := deterministic.NowFunc()

	h := deterministic.NewSlogHandler(underlying)
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

	h := deterministic.NewSlogHandler(underlying)
	ctx := context.Background()

	nf := deterministic.NowFunc()
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

	h := deterministic.NewSlogHandler(underlying)
	ctx := context.Background()

	// Both should return the same result
	if h.Enabled(ctx, slog.LevelInfo) != underlying.Enabled(ctx, slog.LevelInfo) {
		t.Error("Enabled returned different results")
	}
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	underlying := slog.NewTextHandler(buf, nil)

	h := deterministic.NewSlogHandler(underlying)

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

	h := deterministic.NewSlogHandler(underlying)

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

	h1 := deterministic.NewSlogHandler(underlying1)
	h2 := deterministic.NewSlogHandler(underlying2)

	ctx := context.Background()
	dummyTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	r := slog.NewRecord(dummyTime, slog.LevelInfo, "test1", 0)

	require.NoError(t, h1.Handle(ctx, r))

	r = slog.NewRecord(dummyTime, slog.LevelInfo, "test2", 0)
	require.NoError(t, h1.Handle(ctx, r))

	// h2's internal NowFunc should be at the first time since we haven't called it yet
	r = slog.NewRecord(dummyTime, slog.LevelInfo, "test3", 0)
	require.NoError(t, h2.Handle(ctx, r))

	// Get the timestamps from each handler
	output1 := buf1.String()
	output2 := buf2.String()

	// Extract all lines
	lines1 := strings.Split(strings.TrimSpace(output1), "\n")
	lines2 := strings.Split(strings.TrimSpace(output2), "\n")

	// Get the second timestamp from h1 (which should be one second after the first)
	ts1_second := extractTime(lines1[1])
	// Get the first timestamp from h2 (which should be the initial time)
	ts2_first := extractTime(lines2[0])

	// They should be different (h1's second call vs h2's first call)
	if ts1_second == ts2_first {
		t.Errorf("expected different timestamps for different instances, but got same: %s", ts1_second)
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

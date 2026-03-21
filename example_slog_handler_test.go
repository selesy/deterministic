package deterministic_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/selesy/deterministic"
)

func ExampleNewSlogHandler() {
	handler := deterministic.NewSlogHandler(
		slog.NewTextHandler(os.Stdout, nil),
	)
	logger := slog.New(handler)

	logger.InfoContext(context.Background(), "first message")
	logger.InfoContext(context.Background(), "second message")
	// Output:
	// time=2006-01-02T15:04:05.000Z level=INFO msg="first message"
	// time=2006-01-02T15:04:06.000Z level=INFO msg="second message"
}

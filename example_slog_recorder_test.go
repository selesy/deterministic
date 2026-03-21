package deterministic_test

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/selesy/deterministic"
)

func ExampleNewSlogRecorder() {
	recorder := deterministic.NewSlogRecorder()
	logger := slog.New(recorder)

	logger.InfoContext(context.Background(), "user logged in", slog.String("user_id", "123"))
	logger.WarnContext(context.Background(), "high memory usage", slog.Float64("percent", 85.5))

	for i := range recorder.Len() {
		record, _ := recorder.Get(i)
		fmt.Printf("[%s] %s\n", record.Level, record.Message)
	}
	// Output:
	// [INFO] user logged in
	// [WARN] high memory usage
}

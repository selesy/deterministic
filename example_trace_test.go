package deterministic_test

import (
	"context"
	"fmt"

	"github.com/selesy/deterministic"
)

func ExampleNewTraceRecorder() {
	ctx := context.Background()
	recorder := deterministic.NewTraceRecorder()
	defer func() {
		if err := recorder.Shutdown(ctx); err != nil {
			fmt.Println(err.Error())
		}
	}()

	tracer := recorder.Tracer("example")
	_, span := tracer.Start(ctx, "my-operation")
	span.End()

	spans, _ := recorder.FinishedSpans(ctx)
	s := spans[0]
	fmt.Println(s.Name())
	fmt.Println(s.SpanContext().TraceID())
	fmt.Println(s.SpanContext().SpanID())
	fmt.Println(s.StartTime())
	fmt.Println(s.EndTime())
	// Output:
	// my-operation
	// 538c7f96b164bf1b97bb9f4bb472e89f
	// 5b1484f25209c9d9
	// 2006-01-02 15:04:05 +0000 UTC
	// 2006-01-02 15:04:06 +0000 UTC
}

package deterministic_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"
)

func TestTraceRecorder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder := deterministic.NewTraceRecorder()
	t.Cleanup(func() {
		require.NoError(t, recorder.Shutdown(ctx))
	})
	tracer := recorder.Tracer("axtest_test")

	fn2 := func(ctx context.Context) {
		_, span := tracer.Start(ctx, "fn2")
		defer span.End()
	}

	fn1 := func(ctx context.Context) {
		ctx, span := tracer.Start(ctx, "fn1")
		defer span.End()

		for range 10 {
			fn2(ctx)
		}
	}

	fn1(ctx)

	spans, err := recorder.FinishedSpans(ctx)
	require.NoError(t, err)
	require.Len(t, spans, 11)

	parentSpan := spans[10]
	assert.Equal(t, "fn1", parentSpan.Name())
	assert.Equal(t, "538c7f96b164bf1b97bb9f4bb472e89f", parentSpan.SpanContext().TraceID().String())
	assert.Equal(t, "5b1484f25209c9d9", parentSpan.SpanContext().SpanID().String())
	assert.Equal(t, "2006-01-02 15:04:05 +0000 UTC", parentSpan.StartTime().String())
	assert.Equal(t, "2006-01-02 15:04:26 +0000 UTC", parentSpan.EndTime().String())

	lastEndTime := parentSpan.StartTime()
	for i := range 10 {
		childSpan := spans[i]

		// all "inner" spans are named fn2
		assert.Equal(t, "fn2", childSpan.Name())
		// fn1 is parent span
		assert.Equal(t, parentSpan.SpanContext().TraceID(), childSpan.Parent().TraceID())
		assert.Equal(t, parentSpan.SpanContext().SpanID(), childSpan.Parent().SpanID())
		// fn2 span is "within" the temporal range of fn1
		assert.Greater(t, childSpan.StartTime(), parentSpan.StartTime())
		assert.Less(t, childSpan.EndTime(), parentSpan.EndTime())
		// fn2 are sequential
		assert.Greater(t, childSpan.StartTime(), lastEndTime)
		lastEndTime = childSpan.EndTime()
	}
}

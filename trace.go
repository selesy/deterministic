package deterministic

import (
	"context"
	"errors"
	"math/rand"
	"time"

	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// TraceRecorder sets up an in-memory OpenTelemetry trace provider with
// deterministic span/trace IDs and timestamps, making traced code easy to
// test with reproducible output.
type TraceRecorder struct {
	provider *sdk.TracerProvider
	exporter *tracetest.InMemoryExporter
	nowFunc  func() time.Time
}

// NewTraceRecorder creates a [TraceRecorder] backed by an in-memory exporter
// and a deterministic ID generator seeded with a fixed value. The caller is
// responsible for calling [TraceRecorder.Shutdown] when the recorder is no
// longer needed.
func NewTraceRecorder() *TraceRecorder {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdk.NewTracerProvider(
		sdk.WithSpanProcessor(sdk.NewSimpleSpanProcessor(exporter)),
		sdk.WithIDGenerator(newIDGenerator()),
	)

	return &TraceRecorder{
		provider: provider,
		exporter: exporter,
		nowFunc:  NowFunc(),
	}
}

// Shutdown shuts down the exporter and provider, releasing any resources.
func (r *TraceRecorder) Shutdown(ctx context.Context) error {
	return errors.Join(
		r.exporter.Shutdown(ctx),
		r.provider.Shutdown(ctx),
	)
}

// FinishedSpans flushes the trace provider and returns all completed spans
// recorded so far.
func (r *TraceRecorder) FinishedSpans(ctx context.Context) ([]sdk.ReadOnlySpan, error) {
	if err := r.provider.ForceFlush(ctx); err != nil {
		return nil, err
	}

	return r.exporter.GetSpans().Snapshots(), nil
}

// Tracer returns a [trace.Tracer] that stamps every span with deterministic
// timestamps obtained from the recorder's [NowFunc].
func (r *TraceRecorder) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return &tracer{
		Tracer:  r.provider.Tracer(name, opts...),
		nowFunc: r.nowFunc,
	}
}

var _ trace.Tracer = (*tracer)(nil)

type tracer struct {
	trace.Tracer
	nowFunc func() time.Time
}

func (t *tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, s := t.Tracer.Start(ctx, name, append(opts, trace.WithTimestamp(t.nowFunc()))...)

	return ctx, &span{
		Span:    s,
		nowFunc: t.nowFunc,
	}
}

var _ trace.Span = (*span)(nil)

type span struct {
	trace.Span
	nowFunc func() time.Time
}

func (s *span) End(opts ...trace.SpanEndOption) {
	s.Span.End(append(opts, trace.WithTimestamp(s.nowFunc()))...)
}

var _ sdk.IDGenerator = (*idGenerator)(nil)

type idGenerator struct {
	rng *rand.Rand
}

func newIDGenerator() *idGenerator {
	return &idGenerator{rng: rand.New(rand.NewSource(42))} //nolint:gosec
}

func (g *idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	traceID := g.newTraceID(ctx)

	return traceID, g.NewSpanID(ctx, traceID)
}

func (g *idGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	var id trace.SpanID
	_, _ = g.rng.Read(id[:])

	return trace.SpanID(id)
}

func (g *idGenerator) newTraceID(_ context.Context) trace.TraceID {
	var id trace.TraceID
	_, _ = g.rng.Read(id[:])

	return id
}

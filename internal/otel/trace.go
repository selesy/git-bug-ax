package otel

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RecordError records an error on a span if err is non-nil.
// It sets the span status to Error with the error message.
// If err is nil, this is a no-op.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// EndSpan records an error on a span if err is non-nil and then ends the span.
// It is a convenience function combining RecordError and span.End().
// If err is nil, the span is ended without error recording.
func EndSpan(span trace.Span, err error) {
	RecordError(span, err)
	span.End()
}

// EndSpanOnError records an error on a span and ends it, but only if err is non-nil.
// If err is nil, this is a no-op and the span remains open.
func EndSpanOnError(span trace.Span, err error) {
	if err != nil {
		RecordError(span, err)
		span.End()
	}
}

package oteltest

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

// MockSpan is a minimal mock implementation of [trace.Span] that tracks
// RecordError and SetStatus calls for testing.
//
// Use MockSpan in unit tests where you need to verify that a function
// correctly records errors and sets span status. The RecordedErr, RecordedCode,
// RecordedMsg, and Ended fields capture the state after operations.
//
// Example:
//
//	span := &MockSpan{}
//	otel.RecordError(span, err)
//	if span.RecordedErr != err {
//		t.Errorf("expected error to be recorded")
//	}
type MockSpan struct {
	embedded.Span
	RecordedErr  error
	RecordedCode codes.Code
	RecordedMsg  string
	Ended        bool
}

var _ trace.Span = (*MockSpan)(nil)

// AddEvent is a no-op for MockSpan.
func (m *MockSpan) AddEvent(name string, options ...trace.EventOption) {}

// AddLink is a no-op for MockSpan.
func (m *MockSpan) AddLink(link trace.Link) {}

// IsRecording always returns true for MockSpan.
func (m *MockSpan) IsRecording() bool {
	return true
}

// RecordError stores the error in RecordedErr.
func (m *MockSpan) RecordError(err error, options ...trace.EventOption) {
	m.RecordedErr = err
}

// SetStatus stores the code and description in RecordedCode and RecordedMsg.
func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.RecordedCode = code
	m.RecordedMsg = description
}

// SetName is a no-op for MockSpan.
func (m *MockSpan) SetName(name string) {}

// SetAttributes is a no-op for MockSpan.
func (m *MockSpan) SetAttributes(kv ...attribute.KeyValue) {}

// End sets the Ended field to true.
func (m *MockSpan) End(options ...trace.SpanEndOption) {
	m.Ended = true
}

// SpanContext returns an empty trace.SpanContext.
func (m *MockSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

// TracerProvider returns the global TracerProvider.
func (m *MockSpan) TracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

package otel_test

import (
	"errors"
	"testing"

	"go.opentelemetry.io/otel/codes"

	"github.com/selesy/git-bug-agent/internal/otel"
	"github.com/selesy/git-bug-agent/internal/otel/oteltest"
)

func TestRecordError_WithNilError(t *testing.T) {
	span := &oteltest.MockSpan{}
	otel.RecordError(span, nil)

	if span.RecordedErr != nil {
		t.Errorf("expected no error recorded, got %v", span.RecordedErr)
	}
	if span.RecordedCode != codes.Code(0) {
		t.Errorf("expected no status set, got %v", span.RecordedCode)
	}
}

func TestRecordError_WithError(t *testing.T) {
	span := &oteltest.MockSpan{}
	testErr := errors.New("test error")
	otel.RecordError(span, testErr)

	if !errors.Is(span.RecordedErr, testErr) {
		t.Errorf("expected error %v, got %v", testErr, span.RecordedErr)
	}
	if span.RecordedCode != codes.Error {
		t.Errorf("expected status Error, got %v", span.RecordedCode)
	}
	if span.RecordedMsg != testErr.Error() {
		t.Errorf("expected message %q, got %q", testErr.Error(), span.RecordedMsg)
	}
}

func TestEndSpan_WithNilError(t *testing.T) {
	span := &oteltest.MockSpan{}
	otel.EndSpan(span, nil)

	if span.RecordedErr != nil {
		t.Errorf("expected no error recorded, got %v", span.RecordedErr)
	}
	if !span.Ended {
		t.Error("expected span to be ended")
	}
}

func TestEndSpan_WithError(t *testing.T) {
	span := &oteltest.MockSpan{}
	testErr := errors.New("test error")
	otel.EndSpan(span, testErr)

	if !errors.Is(span.RecordedErr, testErr) {
		t.Errorf("expected error %v, got %v", testErr, span.RecordedErr)
	}
	if span.RecordedCode != codes.Error {
		t.Errorf("expected status Error, got %v", span.RecordedCode)
	}
	if span.RecordedMsg != testErr.Error() {
		t.Errorf("expected message %q, got %q", testErr.Error(), span.RecordedMsg)
	}
	if !span.Ended {
		t.Error("expected span to be ended")
	}
}

func TestEndSpanOnError_WithNilError(t *testing.T) {
	span := &oteltest.MockSpan{}
	otel.EndSpanOnError(span, nil)

	if span.RecordedErr != nil {
		t.Errorf("expected no error recorded, got %v", span.RecordedErr)
	}
	if span.Ended {
		t.Error("expected span to remain open")
	}
}

func TestEndSpanOnError_WithError(t *testing.T) {
	span := &oteltest.MockSpan{}
	testErr := errors.New("test error")
	otel.EndSpanOnError(span, testErr)

	if !errors.Is(span.RecordedErr, testErr) {
		t.Errorf("expected error %v, got %v", testErr, span.RecordedErr)
	}
	if span.RecordedCode != codes.Error {
		t.Errorf("expected status Error, got %v", span.RecordedCode)
	}
	if span.RecordedMsg != testErr.Error() {
		t.Errorf("expected message %q, got %q", testErr.Error(), span.RecordedMsg)
	}
	if !span.Ended {
		t.Error("expected span to be ended")
	}
}

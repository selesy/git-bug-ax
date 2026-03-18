package oteltest_test

import (
	"errors"
	"fmt"

	"github.com/selesy/git-bug-agent/internal/otel"
	"github.com/selesy/git-bug-agent/internal/otel/oteltest"
)

func ExampleMockSpan() {
	span := &oteltest.MockSpan{}
	err := errors.New("test error")
	otel.RecordError(span, err)

	fmt.Printf("Error recorded: %v\n", errors.Is(span.RecordedErr, err))
	// Output: Error recorded: true
}

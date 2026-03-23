package backlog

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/selesy/deterministic"
)

func deterministicLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}
	handler := deterministic.NewSlogHandler(
		slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
	logger := slog.New(handler)

	return logger, buf
}

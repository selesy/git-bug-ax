package axtest

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/selesy/deterministic"
)

func DeterministicLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}
	handler := deterministic.NewSlogHandler(
		slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		deterministic.NowFunc(),
	)
	logger := slog.New(handler)

	return logger, buf
}

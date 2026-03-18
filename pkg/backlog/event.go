package backlog

import (
	"context"
	"log/slog"

	"github.com/git-bug/git-bug/cache"
	"github.com/lmittmann/tint"

	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/otel"
)

func newObservedEventConsumer(otelCfg *otel.Config) gitbug.EventConsumer {
	return func(ctx context.Context, events chan cache.BuildEvent) error {
		var hasEvents bool

		for event := range events {
			hasEvents = true

			if event.Err != nil {
				slog.ErrorContext(ctx, "cache build event error", tint.Err(event.Err))
				return event.Err
			}

			switch event.Event {
			case cache.BuildEventCacheIsBuilt:
				otelCfg.Logger().InfoContext(ctx, "CacheIsBuilt", slog.Any("event", event))
			case cache.BuildEventFinished:
				otelCfg.Logger().InfoContext(ctx, "Finished", slog.Any("event", event))
			case cache.BuildEventProgress:
				otelCfg.Logger().InfoContext(ctx, "Progress", slog.Any("event", event))
			case cache.BuildEventRemoveLock:
				otelCfg.Logger().InfoContext(ctx, "RemoveLock", slog.Any("event", event))
			case cache.BuildEventStarted:
				otelCfg.Logger().InfoContext(ctx, "Started", slog.Any("event", event))
			}
		}

		if !hasEvents {
			otelCfg.Logger().InfoContext(ctx, "no caches were built")
		}

		return nil
	}
}

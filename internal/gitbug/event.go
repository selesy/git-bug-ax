package gitbug

import (
	"context"

	"github.com/git-bug/git-bug/cache"
)

// EventConsumer is a function that consumes git-bug build events.
// It receives a context and a channel of events, and returns an error if processing fails.
type EventConsumer func(context.Context, chan cache.BuildEvent) error

// noopEventConsumer is a default event consumer that discards all events.
// It returns early if an event contains an error.
func noopEventConsumer(ctx context.Context, events chan cache.BuildEvent) error {
	for event := range events {
		// Consume events until the channel is closed but stop on error
		if event.Err != nil {
			return event.Err
		}
	}

	return nil
}

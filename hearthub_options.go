package heartfelt

import (
	"context"
	"time"
)

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the hearthub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.heartbeatTimeout = timeout
	}
}

// WithContextOption can set context to the hearthub.
func WithContextOption(ctx context.Context) HeartHubOption {
	return func(hub *HeartHub) {
		hub.ctx = ctx
	}
}

// WithLoggerOption can set logger to the hearthub.
func WithLoggerOption(logger Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

// WithLoggerOption can set watch events to hearthub.
func WithWatchEventOption(eventNames ...string) HeartHubOption {
	return func(hub *HeartHub) {
		hub.watchEvents = map[string]struct{}{}
		for _, eventName := range eventNames {
			hub.watchEvents[eventName] = struct{}{}
		}
	}
}

package heartfelt

import (
	"context"
	"time"
)

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the heartshub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.heartbeatTimeout = timeout
	}
}

// WithContextOption can set context to the heartshub.
func WithContextOption(ctx context.Context) HeartHubOption {
	return func(hub *HeartHub) {
		hub.ctx = ctx
	}
}

// WithLoggerOption can set logger to the heartshub.
func WithLoggerOption(logger Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

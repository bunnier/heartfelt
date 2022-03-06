package heartfelt

import (
	"time"
)

type heartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the hearthub.
func WithTimeoutOption(timeout time.Duration) heartHubOption {
	return func(hub *HeartHub) {
		hub.timeout = timeout
	}
}

// WithLoggerOption can set logger to the hearthub.
func WithLoggerOption(logger Logger) heartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

// WithSubscribeEventNamesOption can set watch events to hearthub.
func WithSubscribeEventNamesOption(eventNames ...string) heartHubOption {
	return func(hub *HeartHub) {
		hub.subscribedEvents = map[string]struct{}{}
		for _, eventName := range eventNames {
			hub.subscribedEvents[eventName] = struct{}{}
		}
	}
}

// WithDegreeOfParallelismOption can control degree of parallelism.
func WithDegreeOfParallelismOption(degree int) heartHubOption {
	return func(hub *HeartHub) {
		hub.parallelisms = make([]*heartHubParallelism, 0, degree)
	}
}

// WithEventBufferSizeOption can set event buffer size.
func WithEventBufferSizeOption(bufferSize int) heartHubOption {
	return func(hub *HeartHub) {
		hub.eventCh = make(chan *Event, bufferSize)
	}
}

// WithHeartbeatBufferSizeOption can set heartbeat buffer size.
func WithHeartbeatBufferSizeOption(bufferSize int) heartHubOption {
	return func(hub *HeartHub) {
		hub.beatChBufferSize = bufferSize
	}
}

// WithVerboseInfoOption can set log level.
func WithVerboseInfoOption() heartHubOption {
	return func(hub *HeartHub) {
		hub.verboseInfo = true
	}
}

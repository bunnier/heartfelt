package heartfelt

import (
	"time"
)

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the hearthub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.timeout = timeout
	}
}

// WithLoggerOption can set logger to the hearthub.
func WithLoggerOption(logger Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

// WithSubscribeEventNamesOption can set watch events to hearthub.
func WithSubscribeEventNamesOption(eventNames ...string) HeartHubOption {
	return func(hub *HeartHub) {
		hub.subscribedEvents = map[string]struct{}{}
		for _, eventName := range eventNames {
			hub.subscribedEvents[eventName] = struct{}{}
		}
	}
}

// WithDegreeOfParallelismOption can control degree of parallelism.
func WithDegreeOfParallelismOption(degree int) HeartHubOption {
	return func(hub *HeartHub) {
		hub.parallelisms = make([]*heartHubParallelism, 0, degree)
	}
}

// WithEventBufferSizeOption can set event buffer size.
func WithEventBufferSizeOption(bufferSize int) HeartHubOption {
	return func(hub *HeartHub) {
		hub.eventCh = make(chan *Event, bufferSize)
	}
}

// WithHeartbeatBufferSizeOption can set heartbeat buffer size.
func WithHeartbeatBufferSizeOption(bufferSize int) HeartHubOption {
	return func(hub *HeartHub) {
		hub.beatChBufferSize = bufferSize
	}
}

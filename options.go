package heartfelt

import (
	"time"
)

type heartHubOption func(hub *abstractHeartHub)

// withTimeoutOption can set timeout to the hearthub.
func withTimeoutOption(timeout time.Duration) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.timeout = timeout
	}
}

// WithLoggerOption can set logger to the hearthub.
func WithLoggerOption(logger Logger) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.logger = logger
	}
}

// WithSubscribeEventNamesOption can set watch events to hearthub.
func WithSubscribeEventNamesOption(eventNames ...string) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.subscribedEvents = map[string]struct{}{}
		for _, eventName := range eventNames {
			hub.subscribedEvents[eventName] = struct{}{}
		}
	}
}

// WithDegreeOfParallelismOption can control degree of parallelism.
func WithDegreeOfParallelismOption(degree int) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.parallelisms = make([]*heartHubParallelism, 0, degree)
	}
}

// WithEventBufferSizeOption can set event buffer size.
func WithEventBufferSizeOption(bufferSize int) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.eventCh = make(chan *Event, bufferSize)
	}
}

// WithHeartbeatBufferSizeOption can set heartbeat buffer size.
func WithHeartbeatBufferSizeOption(bufferSize int) heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.beatChBufferSize = bufferSize
	}
}

// WithVerboseInfoOption can set log level.
func WithVerboseInfoOption() heartHubOption {
	return func(hub *abstractHeartHub) {
		hub.verboseInfo = true
	}
}

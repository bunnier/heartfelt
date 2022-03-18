package heartfelt

import (
	"errors"
	"time"
)

// ErrHubClosed will be return from heartbeat method when the HeartHub be closed.
var ErrHubClosed error = errors.New("heartbeat: this HeartHub has been closed")

// ErrDynamicNotSupported will be return from heartbeatWithTimeout method when the HeartHub do not support dynamic timeout.
var ErrDynamicNotSupported error = errors.New("heartbeat: this HeartHub has been closed")

// ErrNoDefaultTimeout will be return from heartbeat method when the default timeout do not be set.
var ErrNoDefaultTimeout error = errors.New("heartbeat: this HeartHub has been closed")

// HeartHub is the api entrance of this package.
type HeartHub interface {
	// GetEventChannel return a channel for receiving subscribed events.
	GetEventChannel() <-chan *Event

	// Heartbeat will beat the heart of specified key.
	// This method will auto re-watch the key from heartHub after timeout.
	//   @key: The unique key of target service.
	//   @extra: It will be carried back by event data.
	Heartbeat(key string, extra interface{}) error

	// Heartbeat will beat the heart of specified key.
	// This method will auto remove the key from heartHub after timeout.
	//   @key: The unique key of target service.
	//   @extra: It will be carried back by event data.
	DisposableHeartbeat(key string, extra interface{}) error

	// Remove will stop watching the service of key from the heartHub.
	//   @key: The unique key of target service.
	Remove(key string) error

	// Close will stop watch all service keys and release all goroutines.
	Close()
}

// NewFixedTimeoutHeartHub will make a fixed timeout heartHub.
func NewFixedTimeoutHeartHub(timeout time.Duration, options ...heartHubOption) HeartHub {
	options = append(options, WithDefaultTimeoutOption(timeout))
	return newAbstractHeartHub(
		false,
		newBeatsUniqueQueue,
		options...,
	) // Implemented by unique queue.
}

// NewDynamicTimeoutHeartHub will make a dynamic timeout heartHub.
func NewDynamicTimeoutHeartHub(options ...heartHubOption) DynamicTimeoutHeartHub {
	return newAbstractHeartHub(
		true,
		newBeatsUniquePriorityQueue,
		options...,
	) // Implemented by unique priority queue.
}

// DynamicTimeoutHeartHub is a dynamic timeout heartHub.
type DynamicTimeoutHeartHub interface {
	HeartHub

	// Heartbeat will beat the heart of specified key.
	// This method will auto re-watch the key from heartHub after timeout.
	//   @key: The unique key of target service.
	//   @timeout: The timeout duration after this heartbeat.
	//   @extra: It will be carried back by event data.
	HeartbeatWithTimeout(key string, timeout time.Duration, extra interface{}) error

	// Heartbeat will beat the heart of specified key.
	// This method will auto remove the key from heartHub after timeout.
	//   @key: The unique key of target service.
	//   @timeout: The timeout duration after this heartbeat.
	//   @extra: It will be carried back by event data.
	DisposableHeartbeatWithTimeout(key string, timeout time.Duration, extra interface{}) error
}

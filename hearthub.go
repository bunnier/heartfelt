package heartfelt

import (
	"errors"
	"time"
)

// ErrHubClosed will be return from heartbeat method when the HeartHub be closed.
var ErrHubClosed error = errors.New("heartbeat: this HeartHub has been closed")

// ErrDynamicNotSupported will be return from heartbeatWithTimeout method when the HeartHub do not support dynamic timeout.
var ErrDynamicNotSupported error = errors.New("heartbeat: this HeartHub has been closed")

// HeartHub is the api entrance of this package.
type HeartHub interface {
	// GetEventChannel return a channel for receiving subscribed events.
	GetEventChannel() <-chan *Event

	// Heartbeat will beat the heart of specified key.
	// This method will auto re-watch the key from heartHub after timeout.
	//   @key: the unique key of target service.
	Heartbeat(key string) error

	// Heartbeat will beat the heart of specified key.
	// This method will auto remove the key from heartHub after timeout.
	//   @key: the unique key of target service.
	DisposableHeartbeat(key string) error

	// Remove will stop watching the service of key from the heartHub.
	//   @key: the unique key of target service.
	Remove(key string) error

	// Close will stop watch all service keys and release all goroutines.
	Close()
}

// NewFixedTimeoutHeartHub will make a fixed timeout heartHub watcher.
func NewFixedTimeoutHeartHub(timeout time.Duration, options ...heartHubOption) HeartHub {
	options = append(options, WithDefaultTimeoutOption(timeout))
	return newAbstractHeartHub(
		false,
		newBeatsUniqueQueue,
		options...,
	) // Implemented by unique queue.
}

// NewDynamicTimeoutHearthub will make a dynamic timeout heartHub watcher.
func NewDynamicTimeoutHearthub(options ...heartHubOption) DynamicTimeoutHearthub {
	return newAbstractHeartHub(
		true,
		newBeatsUniquePriorityQueue,
		options...,
	) // Implemented by unique priority queue.
}

// DynamicTimeoutHearthub is a dynamic timeout heartHub watcher
type DynamicTimeoutHearthub interface {
	HeartHub

	// Heartbeat will beat the heart of specified key.
	// This method will auto re-watch the key from heartHub after timeout.
	//   @key: the unique key of target service.
	//   @timeout: the timeout duration after this heartbeat.
	HeartbeatWithTimeout(key string, timeout time.Duration) error

	// Heartbeat will beat the heart of specified key.
	// This method will auto remove the key from heartHub after timeout.
	//   @key: the unique key of target service.
	//   @timeout: the timeout duration after this heartbeat.
	DisposableHeartbeatWithTimeout(key string, timeout time.Duration) error
}

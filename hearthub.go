package heartfelt

import (
	"errors"
)

// ErrHubClosed will be return from heartbeat method, after the HeartHub be closed.
var ErrHubClosed error = errors.New("heartbeat: this HeartHub has been closed")

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

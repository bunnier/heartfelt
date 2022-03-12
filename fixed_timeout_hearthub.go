package heartfelt

import (
	"time"
)

type FixedTimeoutHearthub struct {
	abstractHeartHub
}

// NewFixedTimeoutHeartHub will make a fixed timeout heartHub watcher.
func NewFixedTimeoutHeartHub(timeout time.Duration, options ...heartHubOption) HeartHub {
	options = append(options, withTimeoutOption(timeout))
	return newAbstractHeartHub(newBeatsUniqueQueue, options...)
}

// Heartbeat will beat the heart of specified key.
// This method will auto re-watch the key from heartHub after timeout.
//   @key: the unique key of target service.
func (hub *FixedTimeoutHearthub) Heartbeat(key string) error {
	parallelism := hub.getParallelism(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		parallelism.heartbeat(key, false, false)
		return nil
	}
}

// Heartbeat will beat the heart of specified key.
// This method will auto remove the key from heartHub after timeout.
//   @key: the unique key of target service.
func (hub *FixedTimeoutHearthub) DisposableHeartbeat(key string) error {
	parallelism := hub.getParallelism(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		parallelism.heartbeat(key, false, true)
		return nil
	}
}

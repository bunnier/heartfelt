package heartfelt

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"sync"
	"time"
)

var _ HeartHub = (*abstractHeartHub)(nil)
var _ DynamicTimeoutHeartHub = (*abstractHeartHub)(nil)

// abstractHeartHub is a abstract implementation of hearthub.
type abstractHeartHub struct {
	logger      Logger
	verboseInfo bool

	ctx         context.Context
	ctxCancelFn func()

	// In extreme cases, it may have large number of timeout heartbeats.
	// For avoid the timeout handle goroutine occupy much time,
	// we use batchPopThreshold to control maximum number of a batch.
	batchPopThreshold     int
	beatChBufferSize      int
	defaultTimeout        time.Duration
	supportDynamicTimeout bool

	subscribedEvents map[string]struct{}
	eventCh          chan *Event

	// More parallelisms mean more parallel capabilities.
	// A heartHubParallelism manages a part of hearts which are handled in same goroutines group.
	parallelisms []*heartHubParallelism
}

// beatsRepository stores all of alive(no timeout) heartbeats in the heartHubParallelism.
type beatsRepository interface {
	isEmpty() bool
	peek() *beat
	pop() *beat
	push(b *beat) *beat
	remove(key string) *beat
}

// beat means a heartbeat.
type beat struct {
	key         string        // target key
	timeout     time.Duration // timeout duration
	timeoutTime time.Time     // beaten time
	firstTime   time.Time     // first beaten time
	end         bool          // Set to true for removing the key from heartHub.
	disposable  bool          // Set to true for auto removing the key from heartHub after timeout.
	extra       interface{}   // The extra data you pass by calling heartbeat method, it will be carried back by event data.
}

// beatsPool for reuse beats.
var beatsPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return &beat{}
	},
}

// newAbstractHeartHub will make a AbstractHeartHub which is the api entrance of this package.
func newAbstractHeartHub(supportDynamicTimeout bool, repoFactory func() beatsRepository, options ...heartHubOption) *abstractHeartHub {
	hub := &abstractHeartHub{
		logger:                newDefaultLogger(),
		verboseInfo:           false,
		ctx:                   nil,
		ctxCancelFn:           nil,
		batchPopThreshold:     100,
		beatChBufferSize:      100,
		defaultTimeout:        0,
		supportDynamicTimeout: supportDynamicTimeout,
		subscribedEvents: map[string]struct{}{
			EventTimeout: {},
		},
		eventCh:      nil,
		parallelisms: nil,
	}

	for _, option := range options {
		option(hub)
	}

	hub.ctx, hub.ctxCancelFn = context.WithCancel(context.Background())

	if hub.eventCh == nil {
		hub.eventCh = make(chan *Event, 100)
	}

	if hub.parallelisms == nil {
		hub.parallelisms = make([]*heartHubParallelism, 0, 1)
	}

	for i := 1; i <= cap(hub.parallelisms); i++ {
		parallelism := newHeartHubParallelism(i, hub, repoFactory())
		hub.parallelisms = append(hub.parallelisms, parallelism)
		parallelism.startTimeoutCheck()
		parallelism.startHandleHeartbeat()
	}

	return hub
}

func (hub *abstractHeartHub) getParallelism(key string) *heartHubParallelism {
	fnv := fnv.New32a()
	fnv.Write([]byte(key))
	parallelismNum := int(fnv.Sum32()) % len(hub.parallelisms)
	return hub.parallelisms[parallelismNum]
}

// Heartbeat will beat the heart of specified key.
// This method will auto re-watch the key from heartHub after timeout.
//   @key: The unique key of target service.
//   @extra: It will be carried back by event data.
func (hub *abstractHeartHub) Heartbeat(key string, extra interface{}) error {
	if hub.defaultTimeout == 0 {
		return ErrNoDefaultTimeout
	}

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		hub.getParallelism(key).heartbeat(key, false, false, hub.defaultTimeout, extra)
		return nil
	}
}

// Heartbeat will beat the heart of specified key.
// This method will auto remove the key from heartHub after timeout.
//   @key: The unique key of target service.
//   @extra: It will be carried back by event data.
func (hub *abstractHeartHub) DisposableHeartbeat(key string, extra interface{}) error {
	if hub.defaultTimeout == 0 {
		return ErrNoDefaultTimeout
	}

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		hub.getParallelism(key).heartbeat(key, false, true, hub.defaultTimeout, extra)
		return nil
	}
}

// Heartbeat will beat the heart of specified key.
// This method will auto re-watch the key from heartHub after timeout.
//   @key: The unique key of target service.
//   @timeout: The timeout duration after this heartbeat.
//   @extra: It will be carried back by event data.
func (hub *abstractHeartHub) HeartbeatWithTimeout(key string, timeout time.Duration, extra interface{}) error {
	if !hub.supportDynamicTimeout {
		return ErrDynamicNotSupported
	}

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		hub.getParallelism(key).heartbeat(key, false, false, timeout, extra)
		return nil
	}
}

// Heartbeat will beat the heart of specified key.
// This method will auto remove the key from heartHub after timeout.
//   @key: The unique key of target service.
//   @timeout: The timeout duration after this heartbeat.
//   @extra: It will be carried back by event data.
func (hub *abstractHeartHub) DisposableHeartbeatWithTimeout(key string, timeout time.Duration, extra interface{}) error {
	if !hub.supportDynamicTimeout {
		return ErrDynamicNotSupported
	}

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		hub.getParallelism(key).heartbeat(key, false, true, timeout, extra)
		return nil
	}
}

// Remove will stop watching the service of key from the heartHub.
//   @key: the unique key of target service.
func (hub *abstractHeartHub) Remove(key string) error {
	parallelism := hub.getParallelism(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		parallelism.heartbeat(key, true, true, 0, nil)
		return nil
	}
}

// Close will stop watch all service keys and release all goroutines.
func (hub *abstractHeartHub) Close() {
	hub.ctxCancelFn()
	for _, parallelism := range hub.parallelisms {
		parallelism.wakeup()
	}
}

// GetEventChannel return a channel for receiving subscribed events.
func (hub *abstractHeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

const (
	// EventTimeout event will trigger when a heart meet timeout.
	EventTimeout = "TIME_OUT"

	// EventHeartBeat event will be triggered when a heart receives a heartbeat.
	EventHeartBeat = "HEART_BEAT"

	// EventRemoveKey event will be triggered when a heartbeat key be removed.
	EventRemoveKey = "REMOVE_KEY"
)

// Event just means an event, you can use GetEventChannel method to receive subscribed events.
type Event struct {
	EventName   string        `json:"event_name"`
	HeartKey    string        `json:"heart_key"`
	JoinTime    time.Time     `json:"join_time"`  // JoinTime is register time of the key.
	EventTime   time.Time     `json:"event_time"` // Event trigger time.
	Timeout     time.Duration `json:"timeout"`    // Timeout duration.
	TimeoutTime time.Time     `json:"beat_time"`
	Disposable  bool          `json:"disposable"`
	Extra       interface{}   // Extra is the extra data you pass by calling heartbeat method.
}

func (hub *abstractHeartHub) sendEvent(eventName string, beat *beat, eventTime time.Time) bool {
	if _, ok := hub.subscribedEvents[eventName]; !ok {
		return false
	}

	event := &Event{
		EventName:   eventName,
		HeartKey:    beat.key,
		JoinTime:    beat.firstTime,
		EventTime:   eventTime,
		Timeout:     beat.timeout,
		TimeoutTime: beat.timeoutTime,
		Disposable:  beat.disposable,
		Extra:       beat.extra,
	}

	select {
	case hub.eventCh <- event:
		return true
	default:
		eventJsonBytes, _ := json.Marshal(event)
		hub.logger.Error("event channel buffer is full, missed event:", string(eventJsonBytes))
		return false
	}
}

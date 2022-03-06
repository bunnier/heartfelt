package heartfelt

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

// HeartHub is the api entrance of this package.
type HeartHub struct {
	logger      Logger
	ctx         context.Context
	ctxCancelFn func()

	// In extreme cases, it may have large number of timeout heartbeats.
	// For avoid the timeout handle goroutine occupy much time,
	// we use batchPopThreshold to control maximum number of a batch.
	batchPopThreshold int
	beatChBufferSize  int
	timeout           time.Duration

	subscribedEvents map[string]struct{}
	eventCh          chan *Event

	// More parallelisms mean more parallel capability.
	// A heartHubParallelism manage a part of hearts which are handled in same goroutines group.
	parallelisms []*heartHubParallelism
}

// heartHubParallelism manage a part of hearts which are handled in same goroutines group.
type heartHubParallelism struct {
	id       int
	heartHub *HeartHub

	hearts   sync.Map   // hearts map, key => heart.key, value => heart
	beatLink beatLink   // all of alive(no timeout) heartbeats in this heartHubParallelism.
	cond     *sync.Cond // condition & mutex for doing modify

	beatCh chan string // for receive heartbeats
}

// heart just means a heart, the target of heartbeat.
type heart struct {
	key      string // Key is the unique id of a heart.
	lastBeat *beat
}

// beatLink is a doubly linked list,
// store all of alive(no timeout) heartbeats in one heartHubParallelism by time sequences.
type beatLink struct {
	headBeat *beat
	tailBeat *beat
}

// beat means a heartbeat.
type beat struct {
	Heart *heart    // target heart
	Time  time.Time // beaten time

	Prev *beat
	Next *beat
}

// beatsPool for reuse beats.
var beatsPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return &beat{}
	},
}

// After HeartHub be closed, heartbeat method will return this error.
var ErrHubClosed error = errors.New("heartbeat: this HeartHub has been closed")

// NewHeartHub will make a HeartHub which is the api entrance of this package.
func NewHeartHub(options ...HeartHubOption) *HeartHub {
	hub := &HeartHub{
		logger:            newDefaultLogger(),
		ctx:               nil,
		ctxCancelFn:       nil,
		batchPopThreshold: 100,
		beatChBufferSize:  100,
		timeout:           time.Second * 30,
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
		parallelism := newHeartHubParallelism(i, hub)
		hub.parallelisms = append(hub.parallelisms, parallelism)
		parallelism.startHealthCheck()
		parallelism.startHandleHeartbeat()
	}

	return hub
}

func (hub *HeartHub) getParallelism(key string) *heartHubParallelism {
	fnv := fnv.New32a()
	fnv.Write([]byte(key))
	parallelismNum := int(fnv.Sum32()) % len(hub.parallelisms)
	return hub.parallelisms[parallelismNum]
}

// Heartbeat will beat the heart of specified key.
func (hub *HeartHub) Heartbeat(key string) error {
	parallelism := hub.getParallelism(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		parallelism.heartbeat(key)
		now := time.Now()
		hub.sendEvent(EventHeartBeat, key, now, now)
		return nil
	}
}

// Close will release goroutines.
func (hub *HeartHub) Close() {
	hub.ctxCancelFn()
	for _, parallelism := range hub.parallelisms {
		parallelism.wakeup()
	}
}

// GetEventChannel return a channel for receiving events.
func (hub *HeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

const (
	EventTimeout   = "TIME_OUT"   // EventTimeout event will trigger when a heart meet timeout
	EventHeartBeat = "HEART_BEAT" // EventHeartBeat event will trigger when a heart receive a heartbeat.
)

type Event struct {
	EventName string    `json:"event_name"`
	HeartKey  string    `json:"heart_key"`
	BeatTime  time.Time `json:"beat_time"`
	EventTime time.Time `json:"event_time"`
}

func (hub *HeartHub) sendEvent(eventName string, heartKey string, beatTime time.Time, eventTime time.Time) bool {
	if _, ok := hub.subscribedEvents[eventName]; !ok {
		return false
	}

	event := &Event{
		EventName: eventName,
		HeartKey:  heartKey,
		BeatTime:  beatTime,
		EventTime: eventTime,
	}

	select {
	case hub.eventCh <- event:
		return true
	default:
		eventJsonBytes, _ := json.Marshal(event)
		hub.logger.Err("event channel buffer is full, missed event:", string(eventJsonBytes))
		return false
	}
}

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
	verboseInfo bool

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
	// A heartHubParallelism manages a part of hearts which are handled in same goroutines group.
	parallelisms []*heartHubParallelism
}

// heartHubParallelism manages a part of hearts which are handled in same goroutines group.
type heartHubParallelism struct {
	id       int
	heartHub *HeartHub

	hearts    map[string]*heart // hearts is a map: key => heart.key, value => *heart. Modification must be wrapped in cond.L.
	beatsLink beatsLink         // beatsLink stores all of alive(no timeout) heartbeats in the heartHubParallelism.

	// When beatsLink is empty, cond will be use to waiting heartbeat.
	// When modify beatsLink, cond.L will be use for mutual exclusion.
	cond         *sync.Cond
	beatSignalCh chan beatChSignal // for passing heartbeat signal to heartbeat handling goroutine.
}

// heart just means a heart, the target of heartbeat.
type heart struct {
	key        string // Key is the unique id of a heart.
	joinTime   time.Time
	latestBeat *beat
}

// beat means a heartbeat.
type beat struct {
	heart *heart    // target heart
	time  time.Time // beaten time

	prev *beat
	next *beat
}

// beatChSignal is a signal will be pass to heartbeat handling goroutine by beatSignalCh.
type beatChSignal struct {
	key    string // target key
	remove bool   // true will remove key
}

// beatsPool for reuse beats.
var beatsPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return &beat{}
	},
}

// ErrHubClosed will be return from heartbeat method, after the HeartHub be closed.
var ErrHubClosed error = errors.New("heartbeat: this HeartHub has been closed")

// NewHeartHub will make a HeartHub which is the api entrance of this package.
func NewHeartHub(options ...heartHubOption) *HeartHub {
	hub := &HeartHub{
		logger:            newDefaultLogger(),
		verboseInfo:       false,
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
		parallelism.startTimeoutCheck()
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
		parallelism.heartbeat(key, false)
		return nil
	}
}

// Remove will stop watching the service of key from the heartHub.
func (hub *HeartHub) Remove(key string) error {
	parallelism := hub.getParallelism(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		parallelism.heartbeat(key, true)
		return nil
	}
}

// Close will release all goroutines.
func (hub *HeartHub) Close() {
	hub.ctxCancelFn()
	for _, parallelism := range hub.parallelisms {
		parallelism.wakeup()
	}
}

// GetEventChannel return a channel for receiving subscribed events.
func (hub *HeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

const (
	// EventTimeout event will trigger when a heart meet timeout.
	EventTimeout = "TIME_OUT"

	// EventHeartBeat event will trigger when a heart receive a heartbeat.
	EventHeartBeat = "HEART_BEAT"
)

// Event just meats an event, you can use GetEventChannel method to receive subscribed events.
type Event struct {
	EventName string    `json:"event_name"`
	HeartKey  string    `json:"heart_key"`
	JoinTime  time.Time `json:"join_time"`
	BeatTime  time.Time `json:"beat_time"`
	EventTime time.Time `json:"event_time"`
}

func (hub *HeartHub) sendEvent(eventName string, heart *heart, beatTime time.Time, eventTime time.Time) bool {
	if _, ok := hub.subscribedEvents[eventName]; !ok {
		return false
	}

	event := &Event{
		EventName: eventName,
		HeartKey:  heart.key,
		JoinTime:  heart.joinTime,
		BeatTime:  beatTime,
		EventTime: eventTime,
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

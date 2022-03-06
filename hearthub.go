package heartfelt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// HeartHub is the center of this module.
type HeartHub struct {
	logger      Logger
	ctx         context.Context
	ctxCancelFn func()

	timeout          time.Duration
	onceMaxPopCount  int
	beatChBufferSize int

	eventCh          chan *Event
	subscribedEvents map[string]struct{}

	partitions []*heartHubPartition
}

// heartHubPartition manage a part of hearts.
type heartHubPartition struct {
	heartHub *HeartHub

	hearts   sync.Map   // hearts map, key=> heart.key, value => heart
	beatLink beatLink   // all of alive(no timeout) heartbeats in this partition.
	cond     *sync.Cond // condition & mutex for doing modify

	beatCh chan string // for receive heartbeats
}

// heart just means a heart, the target of heartbeat.
type heart struct {
	key      string // Key is the unique id of a heart.
	lastBeat *beat
}

// beatLink is a doubly linked list,
// store all of alive(no timeout) heartbeats in one partition by time sequences.
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
var ErrHubClosed error = errors.New("heartbeat: ErrHubClosed")

// NewHeartHub will make a HeartHub.
func NewHeartHub(options ...HeartHubOption) *HeartHub {
	hub := &HeartHub{
		logger:           newDefaultLogger(),
		ctx:              nil,
		ctxCancelFn:      nil,
		timeout:          time.Second * 30,
		onceMaxPopCount:  10,
		beatChBufferSize: 100,
		eventCh:          nil,
		subscribedEvents: map[string]struct{}{
			EventTimeout: {},
		},
		partitions: nil,
	}

	for _, option := range options {
		option(hub)
	}

	hub.ctx, hub.ctxCancelFn = context.WithCancel(context.Background())

	if hub.eventCh == nil {
		hub.eventCh = make(chan *Event, 100)
	}

	if hub.partitions == nil {
		hub.partitions = make([]*heartHubPartition, 0, 1)
	}

	for i := 0; i < cap(hub.partitions); i++ {
		p := newHeartHubPartition(hub)
		hub.partitions = append(hub.partitions, p)
		p.startHealthCheck()
		p.startHandleHeartbeat()
	}

	return hub
}

// Heartbeat will beat the heart of specified key.
func (hub *HeartHub) Heartbeat(key string) error {
	partition := hub.getPartition(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	default:
		partition.heartbeat(key)
		now := time.Now()
		hub.sendEvent(EventHeartBeat, key, now, now)
		return nil
	}
}

// Close will release goroutines.
func (hub *HeartHub) Close() {
	hub.ctxCancelFn()
	for _, partition := range hub.partitions {
		partition.wakeup()
	}
}

// GetEventChannel return a channel for receiving events.
func (hub *HeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

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
		hub.logger.Err(fmt.Sprintf("error: event buffer is full, miss event: %s", string(eventJsonBytes)))
		return false
	}
}

func (hub *HeartHub) getPartition(key string) *heartHubPartition {
	fnv := fnv.New32a()
	fnv.Write([]byte(key))
	partitionNum := int(fnv.Sum32()) % len(hub.partitions)
	return hub.partitions[partitionNum]
}

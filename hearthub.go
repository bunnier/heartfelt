package heartfelt

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type HeartHub struct {
	logger      Logger
	ctx         context.Context
	ctxCancelFn func()

	heartbeatTimeout time.Duration
	onceMaxPopCount  int

	hearts sync.Map // string->*Heart
	cond   *sync.Cond

	headBeat *beat
	tailBeat *beat

	heartbeatCh chan *heart
	eventCh     chan *Event
}

type heart struct {
	Key      string
	LastBeat *beat
}

type beat struct {
	Heart *heart
	Time  time.Time

	Prev *beat
	Next *beat
}

type Event struct {
	HeartKey  string    `json:"heart_key"`
	EventName string    `json:"event_name"`
	BeatTime  time.Time `json:"beat_time"`
}

var ErrHeartKeyExist error = errors.New("heartbeat: HeartKeyExistErr")

var ErrHeartKeyNoExist error = errors.New("heartbeat: HeartKeyNoExistErr")

var ErrHubClosed error = errors.New("heartbeat: ErrHubClosed")

func NewHeartHub(options ...HeartHubOption) *HeartHub {
	hearthub := &HeartHub{
		logger:      newDefaultLogger(),
		ctx:         context.Background(),
		ctxCancelFn: nil,

		heartbeatTimeout: time.Second * 5,
		onceMaxPopCount:  20,

		hearts: sync.Map{},
		cond:   sync.NewCond(&sync.Mutex{}),

		headBeat: nil,
		tailBeat: nil,

		heartbeatCh: make(chan *heart, 100),
		eventCh:     make(chan *Event, 100),
	}

	for _, option := range options {
		option(hearthub)
	}

	hearthub.ctx, hearthub.ctxCancelFn = context.WithCancel(hearthub.ctx)

	// start goroutines
	hearthub.startHealthCheck()
	hearthub.startHandleHeartbeat()

	return hearthub
}

func (hub *HeartHub) AddHeart(key string) error {
	if _, ok := hub.hearts.LoadOrStore(key, &heart{key, nil}); ok {
		return fmt.Errorf("%w: %s", ErrHeartKeyExist, key)
	}
	return nil
}

func (hub *HeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

func (hub *HeartHub) getHeart(key string) *heart {
	// TODO: add 2 level shard
	if h, ok := hub.hearts.Load(key); ok {
		return h.(*heart)
	} else {
		return nil
	}
}

func (hub *HeartHub) Close() {
	hub.ctxCancelFn()
	hub.cond.Broadcast()
}

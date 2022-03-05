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

	healthCheckingInterval time.Duration
	heartbeatTimeout       time.Duration
	onceMaxPopCount        int

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

		healthCheckingInterval: time.Second * 1,
		heartbeatTimeout:       time.Second * 5,
		onceMaxPopCount:        20,

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

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the heartshub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.heartbeatTimeout = timeout
	}
}

// WithHealthCheckingIntervalOption can set health checking interval to the heartshub.
func WithHealthCheckingIntervalOption(interval time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.healthCheckingInterval = interval
	}
}

// WithContextOption can set context to the heartshub.
func WithContextOption(ctx context.Context) HeartHubOption {
	return func(hub *HeartHub) {
		hub.ctx = ctx
	}
}

// WithLoggerOption can set logger to the heartshub.
func WithLoggerOption(logger Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

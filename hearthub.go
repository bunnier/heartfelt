package heartfelt

import (
	"context"
	"errors"
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

	beatLink beatLink

	heartbeatCh chan *heart
	eventCh     chan *Event
	watchEvents map[string]struct{}
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
	EventName string    `json:"event_name"`
	HeartKey  string    `json:"heart_key"`
	BeatTime  time.Time `json:"beat_time"`
	EventTime time.Time `json:"event_time"`
}

var ErrHubClosed error = errors.New("heartbeat: ErrHubClosed")

func NewHeartHub(options ...HeartHubOption) *HeartHub {
	hearthub := &HeartHub{
		logger:      newDefaultLogger(),
		ctx:         context.Background(),
		ctxCancelFn: nil,

		heartbeatTimeout: time.Second * 30,
		onceMaxPopCount:  10,

		hearts: sync.Map{},
		cond:   sync.NewCond(&sync.Mutex{}),

		beatLink: beatLink{},

		heartbeatCh: nil,
		eventCh:     nil,
		watchEvents: map[string]struct{}{
			EventTimeout: {},
		},
	}

	for _, option := range options {
		option(hearthub)
	}

	hearthub.ctx, hearthub.ctxCancelFn = context.WithCancel(hearthub.ctx)

	if hearthub.eventCh == nil {
		hearthub.eventCh = make(chan *Event, 100)
	}

	if hearthub.heartbeatCh == nil {
		hearthub.heartbeatCh = make(chan *heart, 100)
	}

	// start goroutines
	hearthub.startHealthCheck()
	hearthub.startHandleHeartbeat()

	return hearthub
}

func (hub *HeartHub) getHeart(key string) *heart {
	// TODO 2 level shard
	var h *heart
	if hi, ok := hub.hearts.Load(key); ok {
		h = hi.(*heart)
	} else {
		hi, _ = hub.hearts.LoadOrStore(key, &heart{key, nil})
		h = hi.(*heart)
	}
	return h
}

func (hub *HeartHub) GetEventChannel() <-chan *Event {
	return hub.eventCh
}

func (hub *HeartHub) Heartbeat(key string) error {
	heart := hub.getHeart(key)

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	case hub.heartbeatCh <- heart:
		now := time.Now()
		hub.sendEvent(EventHeartBeat, key, now, now)
		return nil
	}
}

func (hub *HeartHub) Close() {
	hub.ctxCancelFn()
	hub.cond.Broadcast()
}

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the hearthub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.heartbeatTimeout = timeout
	}
}

// WithContextOption can set context to the hearthub.
func WithContextOption(ctx context.Context) HeartHubOption {
	return func(hub *HeartHub) {
		hub.ctx = ctx
	}
}

// WithLoggerOption can set logger to the hearthub.
func WithLoggerOption(logger Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

// WithLoggerOption can set watch events to hearthub.
func WithWatchEventOption(eventNames ...string) HeartHubOption {
	return func(hub *HeartHub) {
		hub.watchEvents = map[string]struct{}{}
		for _, eventName := range eventNames {
			hub.watchEvents[eventName] = struct{}{}
		}
	}
}

// WithEventBufferSizeOption can set event buffer size.
func WithEventBufferSizeOption(bufferSize int) HeartHubOption {
	return func(hub *HeartHub) {
		hub.eventCh = make(chan *Event, bufferSize)
	}
}

// WithHeartbeatBufferSizeOption can set heartbeat buffer size.
func WithHeartbeatBufferSizeOption(bufferSize int) HeartHubOption {
	return func(hub *HeartHub) {
		hub.heartbeatCh = make(chan *heart, bufferSize)
	}
}

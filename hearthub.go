package heartfelt

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type HeartHub struct {
	HeartbeatTimeout         time.Duration
	HeartbeatTimeoutCallback func(string)
	logger                   log.Logger

	Hearts sync.Map // string->*Heart

	LinkMutex sync.Mutex
	HeadBeat  *Beat
	TailBeat  *Beat
}

type Heart struct {
	Key      string
	LastBeat *Beat
}

type Beat struct {
	Heart *Heart
	Time  time.Time

	Prev *Beat
	Next *Beat
}

var ErrHeartKeyExist error = errors.New("heartbeat: HeartKeyExistErr")

var ErrHeartKeyNoExist error = errors.New("heartbeat: HeartKeyNoExistErr")

func NewHeartHub(options ...HeartHubOption) *HeartHub {
	hub := &HeartHub{
		logger:           *log.Default(),
		HeartbeatTimeout: time.Second * 5,
		Hearts:           sync.Map{},
	}

	for _, option := range options {
		option(hub)
	}

	if hub.HeartbeatTimeoutCallback == nil {
		hub.HeartbeatTimeoutCallback = func(key string) {
			hub.logger.Printf("%s is timeout\n", key)
		}
	}

	return hub
}

func (hub *HeartHub) getHeart(key string) *Heart {
	// TODO: 2 level shard
	if heart, ok := hub.Hearts.Load(key); ok {
		return heart.(*Heart)
	} else {
		return nil
	}
}

func (hub *HeartHub) NewHeart(key string) error {
	if _, ok := hub.Hearts.LoadOrStore(key, &Heart{key, nil}); ok {
		return fmt.Errorf("new: %s: %w", key, ErrHeartKeyExist)
	}
	return nil
}

func (hub *HeartHub) Heartbeat(key string) error {
	var heart *Heart
	if heart = hub.getHeart(key); heart == nil {
		return fmt.Errorf("new: %s: %w", key, ErrHeartKeyNoExist)
	}

	// TODO: can be beaten by a channel
	hub.LinkMutex.Lock()
	defer hub.LinkMutex.Unlock()

	lastBeat := heart.LastBeat
	currentBeat := &Beat{Heart: heart, Time: time.Now()} // TODO: beat can be in a pool

	// remove old beat from link
	if lastBeat != nil {
		hub.HeadBeat = lastBeat.Next

		if hub.HeadBeat == lastBeat { // it is the head
			hub.HeadBeat = lastBeat.Next
		} else {
			lastBeat.Prev.Next = lastBeat.Next
		}

		if hub.TailBeat == lastBeat { // it is the tail
			hub.TailBeat = lastBeat.Prev
		} else {
			lastBeat.Next.Prev = lastBeat.Prev
		}
	}

	// push current beat to tail of the link
	if hub.TailBeat == nil { // the link is empty
		hub.TailBeat, hub.HeadBeat = currentBeat, currentBeat
	} else { // add current beat to the tail
		currentBeat.Prev = hub.TailBeat
		hub.TailBeat.Next = currentBeat
		hub.TailBeat = currentBeat
	}

	return nil
}

type HeartHubOption func(hub *HeartHub)

// WithTimeoutOption can set timeout to the heartshub.
func WithTimeoutOption(timeout time.Duration) HeartHubOption {
	return func(hub *HeartHub) {
		hub.HeartbeatTimeout = timeout
	}
}

// WithTimeoutCallbackOption can set timeout callback function to the heartshub.
func WithTimeoutCallbackOption(callback func(string)) HeartHubOption {
	return func(hub *HeartHub) {
		hub.HeartbeatTimeoutCallback = callback
	}
}

// WithLoggerOption can set logger to the heartshub.
func WithLoggerOption(logger log.Logger) HeartHubOption {
	return func(hub *HeartHub) {
		hub.logger = logger
	}
}

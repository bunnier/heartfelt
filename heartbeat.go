package heartfelt

import (
	"sync"
	"time"
)

const (
	EventTimeout   = "TIME_OUT"   // EventTimeout event will trigger when a heart meet timeout
	EventHeartBeat = "HEART_BEAT" // EventHeartBeat event will trigger when a heart receive a heartbeat.
)

func newHeartHubPartition(hub *HeartHub) *heartHubPartition {
	return &heartHubPartition{
		heartHub: hub,
		hearts:   sync.Map{},
		beatLink: beatLink{},
		cond:     sync.NewCond(&sync.Mutex{}),
		beatCh:   make(chan string, hub.beatChBufferSize),
	}
}

func (partition *heartHubPartition) getHeart(key string) *heart {
	var h *heart
	if hi, ok := partition.hearts.Load(key); ok {
		h = hi.(*heart)
	} else {
		hi, _ = partition.hearts.LoadOrStore(key, &heart{key, nil})
		h = hi.(*heart)
	}
	return h
}

func (partition *heartHubPartition) heartbeat(key string) {
	partition.beatCh <- key
}

// startHandleHeartbeat start a goroutine to handle heartbeats.
func (partition *heartHubPartition) startHandleHeartbeat() {
	go func() {
		for {
			var heart *heart
			select {
			case <-partition.heartHub.ctx.Done():
				return
			case key := <-partition.beatCh:
				heart = partition.getHeart(key)
			}

			partition.cond.L.Lock()

			partition.beatLink.remove(heart.lastBeat) // remove old beat

			beat := beatsPool.Get().(*beat)
			beat.Heart = heart
			beat.Time = time.Now()

			heart.lastBeat = beat
			partition.beatLink.push(heart.lastBeat) // push new beat

			partition.cond.L.Unlock()
			partition.cond.Signal() // notify checking health
		}
	}()
}

// startHealthCheck start a goroutine to find timeout hearts.
func (partition *heartHubPartition) startHealthCheck() {
	go func() {
		for {
			partition.cond.L.Lock()

			// when beatlink is empty, use cond to waiting heartbeat.
			for partition.beatLink.headBeat == nil {
				partition.cond.Wait()
				if partition.heartHub.ctx.Err() != nil {
					partition.cond.L.Unlock()
					return
				}
			}

			var nextTimeoutDuration time.Duration
			var popCount int
			for {
				var firstBeat *beat
				if firstBeat = partition.beatLink.peek(); firstBeat != nil {
					break
				}

				if nextTimeoutDuration = partition.heartHub.timeout - time.Since(firstBeat.Time); nextTimeoutDuration > 0 {
					break // have not be timeout yet
				}

				partition.hearts.Delete(firstBeat.Heart.key) // remove the heart from the hearts map
				beat := partition.beatLink.pop()             // pop the timeout heartbeat

				// clean beat and then put back to heartbeat pool
				beat.Heart = nil
				beat.Time = time.Time{}
				beat.Prev = nil
				beat.Next = nil
				beatsPool.Put(beat)

				partition.heartHub.sendEvent(EventTimeout, firstBeat.Heart.key, firstBeat.Time, time.Now()) // trigger timeout event

				if popCount = popCount + 1; popCount >= partition.heartHub.onceMaxPopCount {
					break
				}
			}

			partition.cond.L.Unlock()

			select {
			case <-partition.heartHub.ctx.Done():
				return
			case <-time.After(nextTimeoutDuration):
			}
		}
	}()
}

// wakeup waiting goroutine
func (partition *heartHubPartition) wakeup() {
	partition.cond.Broadcast()
}

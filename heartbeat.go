package heartfelt

import (
	"fmt"
	"sync"
	"time"
)

func newHeartHubParallelism(id int, hub *HeartHub) *heartHubParallelism {
	return &heartHubParallelism{
		id:        id,
		heartHub:  hub,
		hearts:    sync.Map{},
		beatsLink: beatsLink{},
		cond:      sync.NewCond(&sync.Mutex{}),
		beatCh:    make(chan string, hub.beatChBufferSize),
	}
}

func (parallelism *heartHubParallelism) getHeart(key string) *heart {
	var h *heart
	if hi, ok := parallelism.hearts.Load(key); ok {
		h = hi.(*heart)
	} else {
		hi, ok = parallelism.hearts.LoadOrStore(key, &heart{key, time.Now(), nil})
		if !ok {
			parallelism.heartHub.logger.Info("new heart key:", key)
		}
		h = hi.(*heart)
	}
	return h
}

func (parallelism *heartHubParallelism) heartbeat(key string) {
	parallelism.beatCh <- key
}

// startHandleHeartbeat start a goroutine to handle heartbeats.
func (parallelism *heartHubParallelism) startHandleHeartbeat() {
	go func() {
		for {
			var heart *heart
			select {
			case <-parallelism.heartHub.ctx.Done():
				parallelism.heartHub.logger.Info("ctx has been done, exit heartbeat handling goroutine, parallelismId:", parallelism.id)
				return
			case key := <-parallelism.beatCh:
				heart = parallelism.getHeart(key)
			}

			now := time.Now()
			parallelism.heartHub.sendEvent(EventHeartBeat, heart, now, now)

			parallelism.cond.L.Lock()

			parallelism.beatsLink.remove(heart.latestBeat) // remove old beat

			beat := beatsPool.Get().(*beat)
			beat.Heart = heart
			beat.Time = now

			heart.latestBeat = beat
			parallelism.beatsLink.push(heart.latestBeat) // push new beat

			parallelism.cond.L.Unlock()
			parallelism.cond.Signal() // Notify timeout checking goroutine.
		}
	}()
}

// startTimeoutCheck start a goroutine to find timeout hearts.
func (parallelism *heartHubParallelism) startTimeoutCheck() {
	go func() {
		for {
			parallelism.cond.L.Lock()

			parallelism.heartHub.logger.Info("begin to check timeout heartbeat, parallelismId:", parallelism.id)

			// When beatsLink is empty, use cond to waiting heartbeat.
			for parallelism.beatsLink.headBeat == nil {
				parallelism.heartHub.logger.Info("beatsLink is empty, waiting for cond, parallelismId:", parallelism.id)

				parallelism.cond.Wait()
				if parallelism.heartHub.ctx.Err() != nil {
					parallelism.heartHub.logger.Info("wakeup by context canceled, parallelismId:", parallelism.id)
					parallelism.cond.L.Unlock()
					return
				}

				parallelism.heartHub.logger.Info("wakeup by cond, parallelismId:", parallelism.id)
			}

			var nextTimeoutDuration time.Duration
			var popNum int
			for {
				var firstBeat *beat
				if firstBeat = parallelism.beatsLink.peek(); firstBeat != nil {
					break
				}

				if nextTimeoutDuration = parallelism.heartHub.timeout - time.Since(firstBeat.Time); nextTimeoutDuration > 0 {
					break
				}

				// time out workflow...
				parallelism.heartHub.logger.Info(fmt.Sprintf(
					"found a timeout heart, parallelismId: %d, key: %s, join time: %T, last beat: %T",
					parallelism.id,
					firstBeat.Heart.key,
					firstBeat.Heart.joinTime,
					firstBeat.Time))

				parallelism.hearts.Delete(firstBeat.Heart.key) // Remove the heart from the hearts map.
				beat := parallelism.beatsLink.pop()            // Pop the timeout heartbeat.

				// Clean beat and then put back to heartbeat pool.
				beat.Heart = nil
				beat.Time = time.Time{}
				beat.Prev = nil
				beat.Next = nil
				beatsPool.Put(beat)

				parallelism.heartHub.sendEvent(EventTimeout, firstBeat.Heart, firstBeat.Time, time.Now())

				// In extreme cases, it may have large number of timeout heartbeats.
				// For avoid the timeout handle goroutine occupy much time,
				// we use batchPopThreshold to control maximum number of a batch.
				if popNum = popNum + 1; popNum >= parallelism.heartHub.batchPopThreshold {
					nextTimeoutDuration = time.Millisecond * 1
					parallelism.heartHub.logger.Info("pop number has reached once loop maximun, parallelismId:", parallelism.id)
					break
				}
			}

			parallelism.cond.L.Unlock()

			parallelism.heartHub.logger.Info("end checking timeout heartbeat, parallelismId:", parallelism.id)

			select {
			case <-parallelism.heartHub.ctx.Done():
				parallelism.heartHub.logger.Info("ctx has been done, exit timeout checking goroutine, parallelismId:", parallelism.id)
				return
			case <-time.After(nextTimeoutDuration):
			}
		}
	}()
}

// wakeup waiting goroutine.
func (parallelism *heartHubParallelism) wakeup() {
	parallelism.cond.Broadcast()
}

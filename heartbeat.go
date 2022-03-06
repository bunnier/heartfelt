package heartfelt

import (
	"sync"
	"time"
)

func newHeartHubParallelism(id int, hub *HeartHub) *heartHubParallelism {
	return &heartHubParallelism{
		id:       id,
		heartHub: hub,
		hearts:   sync.Map{},
		beatLink: beatLink{},
		cond:     sync.NewCond(&sync.Mutex{}),
		beatCh:   make(chan string, hub.beatChBufferSize),
	}
}

func (parallelism *heartHubParallelism) getHeart(key string) *heart {
	var h *heart
	if hi, ok := parallelism.hearts.Load(key); ok {
		h = hi.(*heart)
	} else {
		hi, ok = parallelism.hearts.LoadOrStore(key, &heart{key, nil})
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
				return
			case key := <-parallelism.beatCh:
				heart = parallelism.getHeart(key)
			}

			parallelism.cond.L.Lock()

			parallelism.beatLink.remove(heart.lastBeat) // remove old beat

			beat := beatsPool.Get().(*beat)
			beat.Heart = heart
			beat.Time = time.Now()

			heart.lastBeat = beat
			parallelism.beatLink.push(heart.lastBeat) // push new beat

			parallelism.cond.L.Unlock()
			parallelism.cond.Signal() // notify checking health
		}
	}()
}

// startHealthCheck start a goroutine to find timeout hearts.
func (parallelism *heartHubParallelism) startHealthCheck() {
	go func() {
		for {
			parallelism.cond.L.Lock()

			// when beatlink is empty, use cond to waiting heartbeat.
			for parallelism.beatLink.headBeat == nil {
				parallelism.heartHub.logger.Info("beatLink is empty, waiting for cond, parallelismId:", parallelism.id)

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
				if firstBeat = parallelism.beatLink.peek(); firstBeat != nil {
					break
				}

				if nextTimeoutDuration = parallelism.heartHub.timeout - time.Since(firstBeat.Time); nextTimeoutDuration > 0 {
					parallelism.heartHub.logger.Info("timeout heartbeats have been handled, parallelismId:", parallelism.id)
					break // have not be timeout yet
				}

				parallelism.hearts.Delete(firstBeat.Heart.key) // remove the heart from the hearts map
				beat := parallelism.beatLink.pop()             // pop the timeout heartbeat

				// Clean beat and then put back to heartbeat pool.
				beat.Heart = nil
				beat.Time = time.Time{}
				beat.Prev = nil
				beat.Next = nil
				beatsPool.Put(beat)

				parallelism.heartHub.sendEvent(EventTimeout, firstBeat.Heart.key, firstBeat.Time, time.Now()) // trigger timeout event

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

			select {
			case <-parallelism.heartHub.ctx.Done():
				return
			case <-time.After(nextTimeoutDuration):
			}
		}
	}()
}

// wakeup waiting goroutine
func (parallelism *heartHubParallelism) wakeup() {
	parallelism.cond.Broadcast()
}

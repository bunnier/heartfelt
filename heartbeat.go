package heartfelt

import (
	"fmt"
	"sync"
	"time"
)

func newHeartHubParallelism(id int, hub *HeartHub) *heartHubParallelism {
	return &heartHubParallelism{
		id:           id,
		heartHub:     hub,
		hearts:       map[string]*heart{},
		beatsLink:    beatsLink{},
		cond:         sync.NewCond(&sync.Mutex{}),
		beatSignalCh: make(chan beatChSignal, hub.beatChBufferSize),
	}
}

func (parallelism *heartHubParallelism) heartbeat(key string, end bool, disposable bool) {
	parallelism.beatSignalCh <- beatChSignal{
		key:        key,
		end:        end,
		disposable: disposable,
	}
}

// startHandleHeartbeat starts a goroutine to handle heartbeats.
func (parallelism *heartHubParallelism) startHandleHeartbeat() {
	go func() {
		for {
			var signal beatChSignal
			select {
			case <-parallelism.heartHub.ctx.Done():
				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info("ctx has been done, exit heartbeat handling goroutine, parallelismId:", parallelism.id)
				}
				return
			case signal = <-parallelism.beatSignalCh:
			}

			parallelism.cond.L.Lock()

			h, ok := parallelism.hearts[signal.key]

			if signal.end { // it is a remove signal.
				if ok {
					delete(parallelism.hearts, h.key)          // Remove the heart from the hearts map.
					parallelism.beatsLink.remove(h.latestBeat) // Remove relative beat from beatlink.
				}
			} else { // So it is a beat signal.
				if !ok {
					// First beating, add the key to hearts.
					h = &heart{signal.key, time.Now(), nil}
					parallelism.hearts[signal.key] = h

					if parallelism.heartHub.verboseInfo {
						parallelism.heartHub.logger.Info("new heart key:", signal.key)
					}
				}

				parallelism.beatsLink.remove(h.latestBeat) // remove old beat

				now := time.Now()
				beat := beatsPool.Get().(*beat)
				beat.heart = h
				beat.time = now
				beat.disposable = signal.disposable
				h.latestBeat = beat
				parallelism.beatsLink.push(h.latestBeat) // push this new beat
				parallelism.heartHub.sendEvent(EventHeartBeat, h, now, now)
			}

			parallelism.cond.L.Unlock()
			parallelism.cond.Signal() // Notify timeout checking goroutine.
		}
	}()
}

// startTimeoutCheck starts a goroutine to find timeout hearts.
func (parallelism *heartHubParallelism) startTimeoutCheck() {
	go func() {
		for {
			parallelism.cond.L.Lock()

			if parallelism.heartHub.verboseInfo {
				parallelism.heartHub.logger.Info("begin to check timeout heartbeat, parallelismId:", parallelism.id)
			}

			// When beatsLink is empty, use cond to waiting heartbeat.
			for parallelism.beatsLink.headBeat == nil {
				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info("beatsLink is empty, waiting for cond, parallelismId:", parallelism.id)
				}

				parallelism.cond.Wait()
				if parallelism.heartHub.ctx.Err() != nil {
					if parallelism.heartHub.verboseInfo {
						parallelism.heartHub.logger.Info("wakeup by context canceled, parallelismId:", parallelism.id)
					}

					parallelism.cond.L.Unlock()
					return
				}

				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info("wakeup by cond, parallelismId:", parallelism.id)
				}
			}

			var nextTimeoutDuration time.Duration
			var popNum int
			for {
				var firstBeat *beat
				if firstBeat = parallelism.beatsLink.peek(); firstBeat == nil {
					break
				}

				if nextTimeoutDuration = parallelism.heartHub.timeout - time.Since(firstBeat.time); nextTimeoutDuration > 0 {
					break
				}

				// time out workflow...
				now := time.Now()

				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info(fmt.Sprintf(
						"found a timeout heart, parallelismId: %d, key: %s, find time: %d, join time: %d, last beat: %d, time offset: %d",
						parallelism.id,
						firstBeat.heart.key,
						now.UnixMilli(),
						firstBeat.heart.joinTime.UnixMilli(),
						firstBeat.time.UnixMilli(),
						now.UnixMilli()-firstBeat.time.UnixMilli()))
				}

				parallelism.heartHub.sendEvent(EventTimeout, firstBeat.heart, firstBeat.time, now)

				parallelism.beatsLink.pop() // Pop the timeout heartbeat.
				if firstBeat.disposable {
					delete(parallelism.hearts, firstBeat.heart.key) // Remove the heart from the hearts map.
					beatsPool.Put(firstBeat)                        // Clean beat and then put back to heartbeat pool.
				} else {
					// Put it to beatlist directly instead of beatChSignal, otherwise might have deadlock.
					// Reuse firstBeat here.
					firstBeat.time = now
					parallelism.beatsLink.push(firstBeat) // push this new beat
				}

				// In extreme cases, it may have large number of timeout heartbeats.
				// For avoid the timeout handle goroutine occupy much time,
				// we use batchPopThreshold to control maximum number of a batch.
				if popNum = popNum + 1; popNum >= parallelism.heartHub.batchPopThreshold {
					nextTimeoutDuration = time.Millisecond * 1

					if parallelism.heartHub.verboseInfo {
						parallelism.heartHub.logger.Info("pop number has reached once loop maximun, parallelismId:", parallelism.id)
					}

					break
				}
			}

			parallelism.cond.L.Unlock()

			if parallelism.heartHub.verboseInfo {
				parallelism.heartHub.logger.Info("end checking timeout heartbeat, parallelismId:", parallelism.id)
			}

			select {
			case <-parallelism.heartHub.ctx.Done():
				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info("ctx has been done, exit timeout checking goroutine, parallelismId:", parallelism.id)
				}
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

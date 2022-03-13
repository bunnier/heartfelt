package heartfelt

import (
	"fmt"
	"sync"
	"time"
)

// heartHubParallelism manages a part of hearts which are handled in same goroutines group.
type heartHubParallelism struct {
	id       int
	heartHub *abstractHeartHub

	beatsRepo beatsRepository // beatsRepo stores all of alive(no timeout) heartbeats in the heartHubParallelism.

	// When beatsLink is empty, cond will be used to waiting heartbeat.
	// When modify beatsLink, cond.L will be used to doing mutual exclusion.
	cond         *sync.Cond
	beatSignalCh chan beatChSignal // for passing heartbeat signal to heartbeat handling goroutine.
}

func newHeartHubParallelism(id int, hub *abstractHeartHub, beatsRepo beatsRepository) *heartHubParallelism {
	return &heartHubParallelism{
		id:           id,
		heartHub:     hub,
		beatsRepo:    beatsRepo,
		cond:         sync.NewCond(&sync.Mutex{}),
		beatSignalCh: make(chan beatChSignal, hub.beatChBufferSize),
	}
}

func (parallelism *heartHubParallelism) heartbeat(key string, end bool, disposable bool, timeout time.Duration) {
	parallelism.beatSignalCh <- beatChSignal{
		key:        key,
		end:        end,
		disposable: disposable,
		timeout:    timeout,
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

			firstBeat := parallelism.beatsRepo.peek() // Save first heartbeat for comparision later.

			now := time.Now()
			if signal.end {
				// It is a remove signal.
				beat := parallelism.beatsRepo.remove(signal.key) // Remove the heart from the repository.
				parallelism.heartHub.sendEvent(EventRemoveKey, beat, now)
			} else {
				// So it is a beat signal.
				beat := beatsPool.Get().(*beat)
				beat.key = signal.key
				beat.firstTime = now
				beat.timeout = signal.timeout
				beat.timeoutTime = now.Add(beat.timeout)
				beat.disposable = signal.disposable

				oldBeat := parallelism.beatsRepo.push(beat) // push this new beat
				if oldBeat != nil {
					beatsPool.Put(oldBeat)
				}

				parallelism.heartHub.sendEvent(EventHeartBeat, beat, now)
			}

			parallelism.cond.L.Unlock()

			// When first heartbeat been changed, notify timeout checking goroutine.
			if parallelism.beatsRepo.peek() != firstBeat {
				parallelism.cond.Signal()
			}
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
			for parallelism.beatsRepo.isEmpty() {
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
				if firstBeat = parallelism.beatsRepo.peek(); firstBeat == nil {
					break
				}

				now := time.Now()
				if nextTimeoutDuration = firstBeat.timeoutTime.Sub(now); nextTimeoutDuration > 0 {
					break
				}

				// time out workflow...
				if parallelism.heartHub.verboseInfo {
					parallelism.heartHub.logger.Info(fmt.Sprintf(
						"found a timeout heart, parallelismId: %d, key: %s, find time: %d, join time: %d, timeout time: %d, time offset: %d",
						parallelism.id,
						firstBeat.key,
						now.UnixMilli(),
						firstBeat.firstTime.UnixMilli(),
						firstBeat.timeoutTime.UnixMilli(),
						now.UnixMilli()-firstBeat.timeoutTime.UnixMilli()))
				}

				parallelism.heartHub.sendEvent(EventTimeout, firstBeat, now)

				parallelism.beatsRepo.pop() // Pop the timeout heartbeat.
				if firstBeat.disposable {
					parallelism.beatsRepo.remove(firstBeat.key) //  Remove the heart from the repository.
					beatsPool.Put(firstBeat)                    // Clean the beat and then put it back to heartbeat pool.
				} else {
					// Put it to the beatlist directly instead of beatChSignal, otherwise might have deadlock.
					// Reuse firstBeat here.
					firstBeat.timeoutTime = now.Add(firstBeat.timeout)
					parallelism.beatsRepo.push(firstBeat)
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

package heartfelt

import (
	"encoding/json"
	"fmt"
	"time"
)

const EventTimeout = "TIME_OUT"

func (hub *HeartHub) Heartbeat(key string) error {
	var heart *heart
	if heart = hub.getHeart(key); heart == nil {
		return fmt.Errorf("%w: %s", ErrHeartKeyNoExist, key)
	}

	select {
	case <-hub.ctx.Done():
		return ErrHubClosed
	case hub.heartbeatCh <- heart:
		return nil
	}
}

func (hub *HeartHub) sendEvent(event *Event) {
	select {
	case hub.eventCh <- event:
		return
	default:
		eventJsonBytes, _ := json.Marshal(event)
		hub.logger.Err(fmt.Sprintf("error: event buffer is full, miss event: %s", string(eventJsonBytes)))
	}
}

func (hub *HeartHub) startHandleHeartbeat() {
	go func() {
		for {
			var heart *heart
			select {
			case <-hub.ctx.Done():
				return
			case heart = <-hub.heartbeatCh:
			}

			hub.cond.L.Lock()

			currentBeat := &beat{Heart: heart, Time: time.Now()} // TODO: beat can reuse from a pool
			lastBeat := heart.LastBeat

			// remove old beat from link
			if lastBeat != nil {
				hub.headBeat = lastBeat.Next

				if hub.headBeat == lastBeat { // it is the head
					hub.headBeat = lastBeat.Next
				} else {
					lastBeat.Prev.Next = lastBeat.Next
				}

				if hub.tailBeat == lastBeat { // it is the tail
					hub.tailBeat = lastBeat.Prev
				} else {
					lastBeat.Next.Prev = lastBeat.Prev
				}
			}

			// push current beat to tail of the link
			if hub.tailBeat == nil { // the link is empty
				hub.tailBeat, hub.headBeat = currentBeat, currentBeat
			} else { // add current beat to the tail
				currentBeat.Prev = hub.tailBeat
				hub.tailBeat.Next = currentBeat
				hub.tailBeat = currentBeat
			}

			hub.cond.L.Unlock()
			hub.cond.Signal() // notify checking health
		}
	}()
}

func (hub *HeartHub) startHealthCheck() {
	go func() {
		for {
			hub.cond.L.Lock()

			for hub.headBeat == nil {
				hub.cond.Wait() // waiting for beat
				if hub.ctx.Err() != nil {
					hub.cond.L.Unlock()
					return
				}
			}

			var nextTimeoutDuration time.Duration
			var popCount int

			for hub.headBeat != nil {
				if nextTimeoutDuration = hub.heartbeatTimeout - time.Since(hub.headBeat.Time); nextTimeoutDuration > 0 {
					break
				}

				hub.sendEvent(&Event{
					HeartKey:  hub.headBeat.Heart.Key,
					EventName: EventTimeout,
					BeatTime:  hub.headBeat.Time,
				})

				hub.headBeat = hub.headBeat.Next

				if popCount = popCount + 1; popCount >= hub.onceMaxPopCount {
					break
				}
			}

			hub.cond.L.Unlock()

			select {
			case <-hub.ctx.Done():
				return
			default:
				time.Sleep(nextTimeoutDuration)
			}
		}
	}()
}

package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/bunnier/heartfelt"
)

func main() {
	// FixedTimeoutHeartHub is a heartbeats watcher of fixed timeout service.
	heartHub := heartfelt.NewFixedTimeoutHeartHub(
		time.Second, // Timeout duration is 1s.
		heartfelt.WithDegreeOfParallelismOption(2),
	)
	eventCh := heartHub.GetEventChannel() // Events will be sent to this channel later.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15) // For exit this example.
	defer cancel()

	// startFakeServices will start 10000 fake services, each service make heartbeat in 500ms regularly.
	// But these services: index in {67, 120, 100, 3456, 4000, 5221, 7899, 9999} will stop work after {its_id} ms.
	// Fortunately, heartHub will catch them all ^_^
	go startFakeServices(ctx, heartHub, 10000, []int{67, 120, 100, 3456, 4000, 5221, 7899, 9999})

	for {
		select {
		case event := <-eventCh:
			// The special service checking will be stop after timeout or heartHub.Remove(key) be called manually.
			log.Default().Printf("received an event: heartKey=%s eventName=%s, timeout duration=%d, timeoutTime=%d, eventTime=%d, offset=%dms",
				event.HeartKey, event.EventName, event.Timeout/time.Millisecond, event.TimeoutTime.UnixMilli(), event.EventTime.UnixMilli(), event.EventTime.Sub(event.TimeoutTime)/time.Millisecond)
		case <-ctx.Done():
			heartHub.Close()
			return
		}
	}
}

// startFakeServices will start fake services.
// Each service will make heartbeat in 500ms regularly.
func startFakeServices(ctx context.Context, heartHub heartfelt.HeartHub, serviceNum int, stuckIds []int) {
	// These ids will stuck later.
	stuckIdsMap := make(map[int]struct{})
	for _, v := range stuckIds {
		stuckIdsMap[v] = struct{}{}
	}

	for i := 1; i <= serviceNum; i++ {
		ctx := ctx
		if _, ok := stuckIdsMap[i]; ok {
			ctx, _ = context.WithTimeout(ctx, time.Duration(i)*time.Millisecond)
		}

		// Each goroutine below represents a service.
		key := strconv.Itoa(i)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Send heartbeat..
					heartHub.DisposableHeartbeat(key, nil)
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}

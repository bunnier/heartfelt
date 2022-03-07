package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/bunnier/heartfelt"
)

func main() {
	// HeartHub is the api entrance of this package.
	heartHub := heartfelt.NewHeartHub(
		heartfelt.WithDegreeOfParallelismOption(2),
		heartfelt.WithTimeoutOption(time.Second),
	)
	eventCh := heartHub.GetEventChannel() // Events will be sent to this channel later.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15) // exit context
	defer cancel()

	// startFakeServices will start 10000 fake services, each service make heartbeat in 200ms regularly.
	// But these services: index in {67, 120, 100, 3456, 4000, 5221, 7899, 9999} will stop work after {its_id} ms.
	// Fortunately, heartHub will catch them all ^_^
	go startFakeServices(ctx, heartHub, 10000, []int{67, 120, 100, 3456, 4000, 5221, 7899, 9999})

	for {
		select {
		case event := <-eventCh:
			// The special service cheking will be stop after timeout or heartHub.Remove(key) be called manually.
			log.Default().Printf("received an event: heartKey=%s eventName=%s, lastBeatTime=%d, eventTime=%d, findTime=%d",
				event.HeartKey, event.EventName, event.BeatTime.UnixMilli(), event.EventTime.UnixMilli(), event.EventTime.UnixMilli()-event.BeatTime.UnixMilli())
		case <-ctx.Done():
			heartHub.Close()
			return
		}
	}
}

// startFakeServices will start fake services.
func startFakeServices(ctx context.Context, heartHub *heartfelt.HeartHub, serviceNum int, stuckIds []int) {
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
		key := strconv.Itoa(i) // convert index to the service key
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					heartHub.Heartbeat(key) // send heartbeat..
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}

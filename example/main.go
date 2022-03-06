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
		heartfelt.WithDegreeOfParallelismOption(4),
		heartfelt.WithTimeoutOption(time.Millisecond*200),
	)

	eventCh := heartHub.GetEventChannel() // Event will be sent to this channel.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*15)
	go startFakeServices(ctx, heartHub) // Start fake service..

	for {
		select {
		case event := <-eventCh:
			log.Default().Printf("receive a event: heartKey=%s\n eventName=%s", event.HeartKey, event.EventName)
		case <-ctx.Done():
			heartHub.Close()
			return
		}
	}
}

// startFakeServices will start 10000 fake services, each service make heartbeat in 100ms regularly.
// But these services: id in {67, 120, 100, 3456, 4000, 5221, 7899, 9999} will stop work after {its_id} ms.
// Sure, heartHub will find them all ^_^
func startFakeServices(ctx context.Context, heartHub *heartfelt.HeartHub) {
	stuckIds := map[int]struct{}{
		67:   {},
		120:  {},
		100:  {},
		3456: {},
		4000: {},
		5221: {},
		7899: {},
		9999: {},
	}

	for i := 0; i < 10000; i++ {
		i := i
		ctx := ctx

		_, isStuckId := stuckIds[i]
		if isStuckId {
			ctx, _ = context.WithTimeout(ctx, time.Duration(i)*time.Millisecond)
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					heartHub.Heartbeat(strconv.Itoa(i))
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
}

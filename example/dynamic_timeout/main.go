package main

import (
	"context"
	"log"
	"time"

	"github.com/bunnier/heartfelt"
)

func main() {
	// DynamicTimeoutHearthub is a heartbeat watcher of dynamic timeout service.
	heartHub := heartfelt.NewDynamicTimeoutHearthub(
		heartfelt.WithDegreeOfParallelismOption(2),
	)
	eventCh := heartHub.GetEventChannel() // Events will be sent to this channel later.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15) // For exit this example.
	defer cancel()

	go func() {
		// This fake service will make heartbeat in 500ms regularly.
		// The timeout will be dynamically set by each heartbeat using {1000ms, 800ms 600ms, 400ms, 200ms}
		// So the heartHub will catch it after the fourth heartbeat.
		for i := 5; i > 0; i-- {
			select {
			case <-ctx.Done():
				return
			default:
				// Send heartbeat..you can also set timeout by the second parameter.
				timeout := time.Duration(i*200) * time.Millisecond
				log.Default().Printf("heartbeat with %v\n", timeout)
				heartHub.DisposableHeartbeatWithTimeout("fakeService", timeout)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	event := <-eventCh
	log.Default().Printf("received an event: heartKey=%s eventName=%s, timeoutTime=%d, eventTime=%d, offset=%dms",
		event.HeartKey, event.EventName, event.TimeoutTime.UnixMilli(), event.EventTime.UnixMilli(), event.EventTime.Sub(event.TimeoutTime)/time.Millisecond)
}

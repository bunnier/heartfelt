package main

import (
	"context"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/bunnier/heartfelt"
)

func main() {
	// DynamicTimeoutHeartHub is a heartbeats watcher of dynamic timeout service.
	heartHub := heartfelt.NewDynamicTimeoutHeartHub(
		heartfelt.WithDegreeOfParallelismOption(1),
	)
	eventCh := heartHub.GetEventChannel() // Events will be sent to this channel later.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	serviceNum := 20 // The number of fake services.

	// Make random timeout list.
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	var timeoutList []int
	for i := 0; i < serviceNum; i++ {
		timeout := rand.Int()%400 + 100
		timeoutList = append(timeoutList, timeout)
	}

	// Start fake services.
	go func() {
		for index, timeout := range timeoutList {
			key := strconv.Itoa(index)
			timeout := timeout
			go func() {
				heartHub.DisposableHeartbeatWithTimeout(key, time.Duration(timeout)*time.Millisecond)
			}()
		}
	}()

	receivedTimeoutList := make([]int, 0, len(timeoutList))
OVER:
	for {
		select {
		case event := <-eventCh:
			log.Default().Printf("received an event: eventName=%s, timeout duration=%d, timeoutTime=%d, eventTime=%d, offset=%dms",
				event.EventName, event.Timeout/time.Millisecond, event.TimeoutTime.UnixMilli(), event.EventTime.UnixMilli(), event.EventTime.Sub(event.TimeoutTime)/time.Millisecond)
			receivedTimeoutList = append(receivedTimeoutList, int(event.Timeout/time.Millisecond)) // Record receive timeout order.
		case <-ctx.Done():
			heartHub.Close()
			break OVER
		}
	}

	sort.IntSlice(timeoutList).Sort() // Received order must equal with sort result.
	if reflect.DeepEqual(timeoutList, receivedTimeoutList) {
		log.Default().Println("Exactly equal!")
	}
}

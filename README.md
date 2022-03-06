# heartfelt[WIP]

A high performance heartbeat watching manager.

## Example

```go
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
```

Output

```bash
2022/03/07 21:37:29 received an event: heartKey=67 eventName=TIME_OUT, lastBeatTime=1646660248392, eventTime=1646660249392, findTime=1000
2022/03/07 21:37:29 received an event: heartKey=100 eventName=TIME_OUT, lastBeatTime=1646660248392, eventTime=1646660249392, findTime=1000
2022/03/07 21:37:29 received an event: heartKey=120 eventName=TIME_OUT, lastBeatTime=1646660248392, eventTime=1646660249392, findTime=1000
2022/03/07 21:37:32 received an event: heartKey=3456 eventName=TIME_OUT, lastBeatTime=1646660251406, eventTime=1646660252406, findTime=1000
2022/03/07 21:37:32 received an event: heartKey=4000 eventName=TIME_OUT, lastBeatTime=1646660251906, eventTime=1646660252906, findTime=1000
2022/03/07 21:37:34 received an event: heartKey=5221 eventName=TIME_OUT, lastBeatTime=1646660253409, eventTime=1646660254409, findTime=1000
2022/03/07 21:37:36 received an event: heartKey=7899 eventName=TIME_OUT, lastBeatTime=1646660255914, eventTime=1646660256914, findTime=1000
2022/03/07 21:37:38 received an event: heartKey=9999 eventName=TIME_OUT, lastBeatTime=1646660257923, eventTime=1646660258923, findTime=1000
```

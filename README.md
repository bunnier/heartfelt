# heartfelt

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![Go](https://github.com/bunnier/heartfelt/actions/workflows/go.yml/badge.svg)](https://github.com/bunnier/heartfelt/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bunnier/heartfelt)](https://goreportcard.com/report/github.com/bunnier/heartfelt)
[![Go Reference](https://pkg.go.dev/badge/github.com/bunnier/heartfelt.svg)](https://pkg.go.dev/github.com/bunnier/heartfelt)

A high performance heartbeat watcher.

## Algorithm

### 1. Fixed timeout watcher

![Algorithm](./docs/fixedtime_algorithm.png)

## Usage

### Example 1: Fixed timeout watcher

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
	// FixedTimeoutHeartHub is a heartbeat watcher of fixed timeout service.
	heartHub := heartfelt.NewFixedTimeoutHeartHub(
		time.Second, // Timeout duration is 1s.
		heartfelt.WithDegreeOfParallelismOption(2),
	)
	eventCh := heartHub.GetEventChannel() // Events will be sent to this channel later.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15) // For exit this example.
	defer cancel()

	// startFakeServices will start 10000 fake services, each service make heartbeat in 500 regularly.
	// But these services: index in {67, 120, 100, 3456, 4000, 5221, 7899, 9999} will stop work after {its_id} ms.
	// Fortunately, heartHub will catch them all ^_^
	go startFakeServices(ctx, heartHub, 10000, []int{67, 120, 100, 3456, 4000, 5221, 7899, 9999})

	for {
		select {
		case event := <-eventCh:
			// The special service checking will be stop after timeout or heartHub.Remove(key) be called manually.
			log.Default().Printf("received an event: heartKey=%s eventName=%s, timeoutTime=%d, eventTime=%d, offset=%dms",
				event.HeartKey, event.EventName, event.TimeoutTime.UnixMilli(), event.EventTime.UnixMilli(), event.EventTime.Sub(event.TimeoutTime)/time.Millisecond)
		case <-ctx.Done():
			heartHub.Close()
			return
		}
	}
}

// startFakeServices will start fake services.
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
					heartHub.DisposableHeartbeat(key)
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
	}
}
```

Output

```bash
2022/03/12 22:07:49 received an event: heartKey=67 eventName=TIME_OUT, timeoutTime=1647094069675, eventTime=1647094069675, offset=0ms
2022/03/12 22:07:49 received an event: heartKey=100 eventName=TIME_OUT, timeoutTime=1647094069675, eventTime=1647094069676, offset=0ms
2022/03/12 22:07:49 received an event: heartKey=120 eventName=TIME_OUT, timeoutTime=1647094069676, eventTime=1647094069676, offset=0ms
2022/03/12 22:07:52 received an event: heartKey=3456 eventName=TIME_OUT, timeoutTime=1647094072684, eventTime=1647094072684, offset=0ms
2022/03/12 22:07:53 received an event: heartKey=4000 eventName=TIME_OUT, timeoutTime=1647094073185, eventTime=1647094073185, offset=0ms
2022/03/12 22:07:54 received an event: heartKey=5221 eventName=TIME_OUT, timeoutTime=1647094074686, eventTime=1647094074686, offset=0ms
2022/03/12 22:07:57 received an event: heartKey=7899 eventName=TIME_OUT, timeoutTime=1647094077193, eventTime=1647094077193, offset=0ms
2022/03/12 22:07:59 received an event: heartKey=9999 eventName=TIME_OUT, timeoutTime=1647094079196, eventTime=1647094079196, offset=0ms
```

### Example 2: Dynamic timeout watcher

```go
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

```

Output

```bash
2022/03/13 00:12:52 heartbeat with 1s
2022/03/13 00:12:53 heartbeat with 800ms
2022/03/13 00:12:53 heartbeat with 600ms
2022/03/13 00:12:54 heartbeat with 400ms
2022/03/13 00:12:54 received an event: heartKey=fakeService eventName=TIME_OUT, timeoutTime=1647101574404, eventTime=1647101574405, offset=0ms
```

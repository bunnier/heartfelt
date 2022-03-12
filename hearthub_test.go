package heartfelt

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

func Test_heartbeats(t *testing.T) {
	heartHub := NewFixedTimeoutHeartHub(
		time.Millisecond*10,
		WithDegreeOfParallelismOption(1),
	).(*abstractHeartHub)

	defer heartHub.Close()

	heartHub.DisposableHeartbeat("service1")
	heartHub.DisposableHeartbeat("service2")
	heartHub.Heartbeat("service3")

	time.Sleep(time.Millisecond * 2) // Waiting for inner goroutine done.

	heartHub.parallelisms[0].cond.L.Lock()

	// Check heartbeats number.
	beatsCount := 0
	tmpBeat := heartHub.parallelisms[0].beatsRepo.(*beatsUniqueQueue).link.head
	for tmpBeat != nil {
		tmpBeat = tmpBeat.next
		beatsCount++
	}

	if beatsCount != 3 {
		t.Errorf("hearthub heartbeats beats num error, want 3 get %v", beatsCount)
	}

	// Check disposable.
	if heartHub.parallelisms[0].beatsRepo.(*beatsUniqueQueue).link.tail.data.disposable {
		t.Errorf("hearthub heartbeats disposable error, want true get false")
	}

	heartHub.parallelisms[0].cond.L.Unlock()

	time.Sleep(time.Millisecond * 11) // Waiting for timeout.

	// Check disposable result.
	if len(heartHub.parallelisms[0].beatsRepo.(*beatsUniqueQueue).lastBeatsMap) != 1 {
		t.Errorf("hearthub heartbeats disposable error, want 1 get %v", len(heartHub.parallelisms[0].beatsRepo.(*beatsUniqueQueue).lastBeatsMap))
	}

	if _, ok := heartHub.parallelisms[0].beatsRepo.(*beatsUniqueQueue).lastBeatsMap["service3"]; !ok {
		t.Errorf("hearthub heartbeats disposable error, service3 is not existed")
	}
}

func Test_timeoutCheck(t *testing.T) {
	heartHub := NewFixedTimeoutHeartHub(
		time.Millisecond*10,
		WithDegreeOfParallelismOption(1),
	)
	defer heartHub.Close()
	eventCh := heartHub.GetEventChannel()

	heartHub.DisposableHeartbeat("service1")
	time.Sleep(time.Millisecond * 12) // Waiting for timeout.

	select {
	case event := <-eventCh:
		if event.HeartKey != "service1" || !event.Disposable || event.EventName != EventTimeout {
			t.Errorf("hearthub timeout checking error event1")
		}
	default:
		t.Errorf("hearthub timeout checking error event1, should have an event")
	}

	time.Sleep(time.Millisecond * 12) // Waiting for timeout.
	select {
	case event := <-eventCh:
		t.Errorf("hearthub timeout checking error, should be empty but get %v", event.EventName)
	default:
	}

	heartHub.Heartbeat("service1")
	time.Sleep(time.Millisecond * 24) // Waiting for timeout.
	for i := 0; i < 2; i++ {
		select {
		case event := <-eventCh:
			if event.HeartKey != "service1" || event.Disposable || event.EventName != EventTimeout {
				t.Errorf("hearthub timeout checking error event2")
			}
		default:
			t.Errorf("hearthub timeout checking error event1, should have an event")
		}
	}

	// Remove key.
	heartHub.Remove("service1")
	time.Sleep(time.Millisecond * 12) // Waiting for timeout.
	select {
	case event := <-eventCh:
		t.Errorf("hearthub timeout checking error, should be empty but get %v", event.EventName)
	default:
	}
}

func Test_workflow(t *testing.T) {
	heartHub := NewFixedTimeoutHeartHub(
		time.Millisecond*200,
		WithDegreeOfParallelismOption(2),
	)
	eventCh := heartHub.GetEventChannel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1) // exit context
	defer cancel()

	want := []int{67, 100, 123, 456, 789}
	get := make([]int, 0, len(want))
	go startFakeServices(ctx, heartHub, 2000, want)

OVER:
	for {
		select {
		case event := <-eventCh:
			id, _ := strconv.Atoi(event.HeartKey)
			get = append(get, id)
		case <-ctx.Done():
			heartHub.Close()
			break OVER
		}
	}

	sort.IntSlice(get).Sort()
	if !reflect.DeepEqual(want, get) {
		t.Errorf("hearthub workflow want %v get %v", want, get)
	}
}

// startFakeServices will start fake services.
func startFakeServices(ctx context.Context, heartHub HeartHub, serviceNum int, stuckIds []int) {
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

		key := strconv.Itoa(i) // convert index to the service key
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					heartHub.DisposableHeartbeat(key)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}
}

func Test_hearthubOptions(t *testing.T) {
	timeout := time.Millisecond * 200
	eventBufferSize := 10
	degreeOfParallelism := 10

	heartHub := NewFixedTimeoutHeartHub(
		timeout,
		WithEventBufferSizeOption(eventBufferSize),
		WithDegreeOfParallelismOption(degreeOfParallelism),
		WithSubscribeEventNamesOption(EventHeartBeat),
	).(*abstractHeartHub)
	defer heartHub.Close()
	eventCh := heartHub.GetEventChannel()

	if eventBufferSize != cap(eventCh) {
		t.Errorf("hearthub hearthubOptions timeout want %v get %v", eventBufferSize, cap(eventCh))
	}

	if timeout != heartHub.defaultTimeout {
		t.Errorf("hearthub hearthubOptions event buffer size want %v get %v", timeout, int(heartHub.defaultTimeout))
	}

	if degreeOfParallelism != len(heartHub.parallelisms) {
		t.Errorf("hearthub hearthubOptions event buffer size want %v get %v", degreeOfParallelism, len(heartHub.parallelisms))
	}

	if _, ok := heartHub.subscribedEvents[EventHeartBeat]; !ok {
		t.Errorf("hearthub hearthubOptions event name error")
	}

	if _, ok := heartHub.subscribedEvents[EventHeartBeat]; !ok {
		t.Errorf("hearthub hearthubOptions event name error")
	}
}

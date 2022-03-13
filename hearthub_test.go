package heartfelt

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"
)

func Test_fixedTimeoutWorkflow(t *testing.T) {
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

func Test_fixedTimeoutParallelWorkflow(t *testing.T) {
	heartHub := NewFixedTimeoutHeartHub(
		time.Millisecond*200,
		WithDegreeOfParallelismOption(4),
	)
	eventCh := heartHub.GetEventChannel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
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

func Test_dynamicTimeoutWorkflow(t *testing.T) {
	heartHub := NewDynamicTimeoutHeartHub(
		WithDegreeOfParallelismOption(4),
	)
	defer heartHub.Close()
	eventCh := heartHub.GetEventChannel()

	heartHub.DisposableHeartbeatWithTimeout("service1", time.Millisecond*10)
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

	heartHub.DisposableHeartbeatWithTimeout("service1", time.Millisecond*15)
	heartHub.DisposableHeartbeatWithTimeout("service2", time.Millisecond*12)
	heartHub.DisposableHeartbeatWithTimeout("service3", time.Millisecond*200)
	heartHub.DisposableHeartbeatWithTimeout("service4", time.Millisecond*1)
	wantTimeoutService := []string{"service4", "service2", "service1", "service3"}
	time.Sleep(time.Millisecond * 210) // Waiting for timeout.
	var gotTimeoutServices []string
	for i := 0; i < len(wantTimeoutService); i++ {
		select {
		case event := <-eventCh:
			gotTimeoutServices = append(gotTimeoutServices, event.HeartKey)
		default:
			t.Errorf("hearthub timeout checking error event1, should have an event")
		}
	}

	if !reflect.DeepEqual(wantTimeoutService, gotTimeoutServices) {
		t.Errorf("hearthub workflow order checking failed want %v got %v", wantTimeoutService, gotTimeoutServices)
	}

	heartHub.HeartbeatWithTimeout("service1", time.Millisecond*15)
	heartHub.HeartbeatWithTimeout("service2", time.Millisecond*12)
	heartHub.Remove("service1")
	heartHub.Remove("service2")
	time.Sleep(time.Millisecond * 16) // Waiting for timeout.
	select {
	case event := <-eventCh:
		t.Errorf("hearthub timeout checking error, should be empty but get %v", event.EventName)
	default:
	}
}

func Test_dynamicTimeoutParallelWorkflow(t *testing.T) {
	heartHub := NewDynamicTimeoutHeartHub(
		WithDegreeOfParallelismOption(1),
	)
	eventCh := heartHub.GetEventChannel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	serviceNum := 50
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	var timeoutList []int
	for i := 0; i < serviceNum; i++ {
		timeoutRand := rand.Int()%400 + 100 // Random timeout in {100-400}ms.
		timeoutList = append(timeoutList, timeoutRand)
	}

	go func() {
		for index, timeout := range timeoutList {
			key := strconv.Itoa(index)
			timeout := timeout
			go func() {
				heartHub.DisposableHeartbeatWithTimeout(key, time.Duration(timeout)*time.Millisecond)
			}()
		}
	}()

	get := make([]int, 0, len(timeoutList))
OVER:
	for {
		select {
		case event := <-eventCh:
			id, _ := strconv.Atoi(strconv.Itoa(int(event.Timeout / time.Millisecond)))
			get = append(get, id)
		case <-ctx.Done():
			heartHub.Close()
			break OVER
		}
	}

	if len(get) != len(timeoutList) {
		t.Errorf("hearthub workflow miss timeout want %v get %v", len(timeoutList), len(get))
	}

	sort.IntSlice(timeoutList).Sort()
	if !reflect.DeepEqual(timeoutList, get) {
		t.Errorf("hearthub workflow want %v get %v", timeoutList, get)
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

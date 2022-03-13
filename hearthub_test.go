package heartfelt

import (
	"reflect"
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

func Test_dynamicTimeoutWorkflow(t *testing.T) {
	heartHub := NewDynamicTimeoutHearthub(
		WithDegreeOfParallelismOption(1),
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

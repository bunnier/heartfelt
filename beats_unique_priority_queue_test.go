package heartfelt

import (
	"reflect"
	"testing"
	"time"
)

func Test_beatsUniquePriorityQueue_isEmpty(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}
			if got := queue.isEmpty(); got != tt.want {
				t.Errorf("beatsUniquePriorityQueue.isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_peek(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}
			if got := queue.peek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.peek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_pop(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}
			if got := queue.pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_push(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	type args struct {
		b *beat
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}
			if got := queue.push(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.push() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_remove(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}
			if got := queue.remove(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_heap_peek(t *testing.T) {
	nodeSlice := []*beat{
		{key: "2", timeoutTime: time.Now().Add(2 * time.Second)},
		{key: "1", timeoutTime: time.Now().Add(1 * time.Second)},
		{key: "4", timeoutTime: time.Now().Add(4 * time.Second)},
		{key: "5", timeoutTime: time.Now().Add(5 * time.Second)},
		{key: "3", timeoutTime: time.Now().Add(3 * time.Second)},
	}

	h := heap{make([]*beat, 0)}
	h.push(nodeSlice[0])
	h.push(nodeSlice[1])
	h.push(nodeSlice[2])
	h.push(nodeSlice[3])
	h.push(nodeSlice[4])

	// check length
	if len(h.items) != 5 {
		t.Errorf("heap.push() length %v, want 3", len(h.items))
	}
	checkHeapOrder(t, &h)

	// duplicate
	h.push(nodeSlice[0])
	h.push(nodeSlice[1])
	h.push(nodeSlice[2])
	h.push(nodeSlice[3])
	h.push(nodeSlice[4])
	h.push(nodeSlice[0])
	h.push(nodeSlice[1])
	if len(h.items) != 7 {
		t.Errorf("heap.push() length %v, want 7", len(h.items))
	}
	checkHeapOrder(t, &h)

	// single
	h.push(nodeSlice[0])
	if len(h.items) != 1 {
		t.Errorf("heap.push() length %v, want 1", len(h.items))
	}
	checkHeapOrder(t, &h)

	// empty
	if len(h.items) != 0 {
		t.Errorf("heap.push() length %v, want 0", len(h.items))
	}

	if b := h.pop(); b != nil {
		t.Errorf("heap.pop() got %v, want nil", b)
	}
}

func checkHeapOrder(t *testing.T, h *heap) {
	ti := time.Time{}
	len := len(h.items)
	for i := 0; i < len; i++ {
		currentTi := h.pop().timeoutTime
		if ti.After(currentTi) {
			t.Errorf("heap.pop() order is wrong, %v", h.items)
			break
		}
		ti = currentTi
	}
}

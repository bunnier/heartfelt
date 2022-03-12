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
		{
			name: "isEmpty_empty",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{make([]*beat, 0)},
			},
			want: true,
		},
		{
			name: "isEmpty_notEmpty",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{items: []*beat{{}}},
			},
			want: false,
		},
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
	b := &beat{}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		{
			name: "peek_empty",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{make([]*beat, 0)},
			},
			want: nil,
		},
		{
			name: "peek_notEmpty",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{[]*beat{b}},
			},
			want: b,
		},
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

func Test_beatsUniquePriorityQueue_push(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]int
		minHeap      heap
	}
	type args struct {
		b *beat
	}
	b1 := &beat{key: "service1"}
	b2 := &beat{key: "service2"}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		{
			name: "push_notExist",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{make([]*beat, 0)},
			},
			args: args{b1},
		},
		{
			name: "push_exist",
			fields: fields{
				lastBeatsMap: map[string]int{"service1": 0},
				minHeap:      heap{[]*beat{b1}},
			},
			args: args{b1},
		},
		{
			name: "push_existOther",
			fields: fields{
				lastBeatsMap: map[string]int{"service2": 1},
				minHeap:      heap{[]*beat{b2}},
			},
			args: args{b1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
			}

			if got := queue.push(tt.args.b); got != nil && got.key != tt.args.b.key {
				t.Errorf("beatsUniquePriorityQueue.push() return = %v, want %v", got.key, tt.args.b.key)
			}

			if index, ok := queue.lastBeatsMap[tt.args.b.key]; !ok {
				t.Errorf("beatsUniquePriorityQueue.push() map key not found")
				if queue.minHeap.items[index] != tt.args.b {
					t.Errorf("beatsUniquePriorityQueue.push() map value not equal")
				}
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
	b1 := &beat{key: "service1"}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		{
			name: "remove_empty",
			fields: fields{
				lastBeatsMap: make(map[string]int),
				minHeap:      heap{make([]*beat, 0)},
			},
			args: args{"service1"},
			want: nil,
		},
		{
			name: "remove_notEmpty",
			fields: fields{
				lastBeatsMap: map[string]int{"service1": 0},
				minHeap:      heap{[]*beat{b1}},
			},
			args: args{"service1"},
			want: b1,
		},
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

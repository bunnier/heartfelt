package heartfelt

import (
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test_beatsUniquePriorityQueue_isEmpty(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*heapNode
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
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{make([]*heapNode, 0)},
			},
			want: true,
		},
		{
			name: "isEmpty_notEmpty",
			fields: fields{
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{items: []*heapNode{{}}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
				nodePool: sync.Pool{
					New: func() interface{} {
						return &heapNode{}
					},
				},
			}
			if got := queue.isEmpty(); got != tt.want {
				t.Errorf("beatsUniquePriorityQueue.isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_peek(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*heapNode
		minHeap      heap
	}
	b := &beat{}
	node := &heapNode{b, 0}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		{
			name: "peek_empty",
			fields: fields{
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{make([]*heapNode, 0)},
			},
			want: nil,
		},
		{
			name: "peek_notEmpty",
			fields: fields{
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{[]*heapNode{node}},
			},
			want: b,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
				nodePool: sync.Pool{
					New: func() interface{} {
						return &heapNode{}
					},
				},
			}
			if got := queue.peek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.peek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_push(t *testing.T) {

	type fields struct {
		lastBeatsMap map[string]*heapNode
		minHeap      heap
	}
	type args struct {
		b *beat
	}

	b1 := &beat{key: "service1"}
	node1 := &heapNode{b1, 0}
	b2 := &beat{key: "service2"}
	node2 := &heapNode{b2, 0}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		{
			name: "push_notExist",
			fields: fields{
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{make([]*heapNode, 0)},
			},
			args: args{b1},
		},
		{
			name: "push_exist",
			fields: fields{
				lastBeatsMap: map[string]*heapNode{"service1": node1},
				minHeap:      heap{[]*heapNode{node1}},
			},
			args: args{b1},
		},
		{
			name: "push_existOther",
			fields: fields{
				lastBeatsMap: map[string]*heapNode{"service2": node2},
				minHeap:      heap{[]*heapNode{node2}},
			},
			args: args{b1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
				nodePool: sync.Pool{
					New: func() interface{} {
						return &heapNode{}
					},
				},
			}

			if got := queue.push(tt.args.b); got != nil && got.key != tt.args.b.key {
				t.Errorf("beatsUniquePriorityQueue.push() return = %v, want %v", got.key, tt.args.b.key)
			}

			if index, ok := queue.lastBeatsMap[tt.args.b.key]; !ok {
				t.Errorf("beatsUniquePriorityQueue.push() map key not found")
				if queue.minHeap.items[index.heapIndex].data != tt.args.b {
					t.Errorf("beatsUniquePriorityQueue.push() map value not equal")
				}
			}
		})
	}
}

func Test_beatsUniquePriorityQueue_remove(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*heapNode
		minHeap      heap
	}
	type args struct {
		key string
	}

	b1 := &beat{key: "service1"}
	node1 := &heapNode{b1, 0}
	b2 := &beat{key: "service2"}
	node2 := &heapNode{b2, 1}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		{
			name: "remove_empty",
			fields: fields{
				lastBeatsMap: make(map[string]*heapNode),
				minHeap:      heap{make([]*heapNode, 0)},
			},
			args: args{"service1"},
			want: nil,
		},
		{
			name: "remove_notEmpty",
			fields: fields{
				lastBeatsMap: map[string]*heapNode{"service1": node1},
				minHeap:      heap{[]*heapNode{node1}},
			},
			args: args{"service1"},
			want: b1,
		},
		{
			name: "remove_notEmpty2",
			fields: fields{
				lastBeatsMap: map[string]*heapNode{"service1": node1, "service2": node2},
				minHeap:      heap{[]*heapNode{node1, node2}},
			},
			args: args{"service2"},
			want: b2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &beatsUniquePriorityQueue{
				lastBeatsMap: tt.fields.lastBeatsMap,
				minHeap:      tt.fields.minHeap,
				nodePool: sync.Pool{
					New: func() interface{} {
						return &heapNode{}
					},
				},
			}
			if got := queue.remove(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniquePriorityQueue.remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_queue(t *testing.T) {
	queue := newBeatsUniquePriorityQueue()

	// randon test
	loop := 1000
	removeKeys := make([]string, loop/2)
	for i := 0; i < loop; i++ {
		rand := rand.New(rand.NewSource(time.Now().Unix()))
		timeout := rand.Int()
		key := strconv.Itoa(timeout)

		if i%2 == 0 {
			removeKeys = append(removeKeys, key)
		}

		queue.push(&beat{
			key:         strconv.Itoa(timeout),
			timeoutTime: time.Now().Add(time.Duration(timeout) * time.Second),
		})
	}

	for _, removeKey := range removeKeys {
		queue.remove(removeKey)
	}

	checkHeapOrderByPop(t, &(queue.(*beatsUniquePriorityQueue)).minHeap)

	var lastBeat *beat
	var currentBeat *beat
	for i := 0; i < loop/2; i++ {
		currentBeat = queue.pop()
		if lastBeat != nil {
			if lastBeat.timeoutTime.After(currentBeat.timeoutTime) {
				t.Errorf("priority_queue.pop() order is wrong.")
			}
		}
		lastBeat = currentBeat
	}

	if !queue.isEmpty() {
		t.Errorf("priority_queue.isEmpty() want true got false")
	}
}

func Test_heap(t *testing.T) {
	h := heap{make([]*heapNode, 0)}

	// empty
	if len(h.items) != 0 {
		t.Errorf("heap.push() length %v, want 0", len(h.items))
	}

	if bPop := h.pop(); bPop != nil {
		t.Errorf("heap.pop() got %v, want nil", bPop)
	}

	// single
	node := &heapNode{&beat{key: "1", timeoutTime: time.Now().Add(1 * time.Second)}, 0}
	h.push(node)
	if len(h.items) != 1 {
		t.Errorf("heap.push() length %v, want 1", len(h.items))
	}

	if nodePop := h.pop(); node != nodePop {
		t.Errorf("heap.pop() got %v, want %v", nodePop, node)
	}

	nodeSlice := []*heapNode{
		{&beat{key: "2", timeoutTime: time.Now().Add(2 * time.Second)}, 0},
		{&beat{key: "1", timeoutTime: time.Now().Add(1 * time.Second)}, 0},
		{&beat{key: "4", timeoutTime: time.Now().Add(4 * time.Second)}, 0},
		{&beat{key: "5", timeoutTime: time.Now().Add(5 * time.Second)}, 0},
		{&beat{key: "3", timeoutTime: time.Now().Add(3 * time.Second)}, 0},
		{&beat{key: "7", timeoutTime: time.Now().Add(7 * time.Second)}, 0},
		{&beat{key: "6", timeoutTime: time.Now().Add(6 * time.Second)}, 0},
	}

	for _, node := range nodeSlice {
		h.push(node)
	}

	checkHeapOrderByPop(t, &h)

	// randon test
	loop := 1000
	for i := 0; i < loop; i++ {
		rand := rand.New(rand.NewSource(time.Now().Unix()))
		timeout := rand.Int()
		h.push(&heapNode{&beat{key: strconv.Itoa(timeout), timeoutTime: time.Now().Add(time.Duration(timeout) * time.Second)}, 0})
	}

	if len(h.items) != loop {
		t.Errorf("heap.push() length %v, want %v", len(h.items), len(nodeSlice))
	}

	checkHeapOrderByPop(t, &h)
}

func checkHeapOrderByPop(t *testing.T, h *heap) {
	before := time.Time{}
	loop := len(h.items)
	for i := 0; i < loop; i++ {
		current := h.pop().data.timeoutTime
		if before.After(current) {
			t.Errorf("heap.pop() order is wrong, %v", h.items)
			break
		}
		before = current
	}

	// empty
	if len(h.items) != 0 {
		t.Errorf("heap.push() length %v, want 0", len(h.items))
	}
}

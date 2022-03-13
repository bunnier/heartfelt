package heartfelt

import (
	"reflect"
	"sync"
	"testing"
)

func Test_beatsUniqueQueue_isEmpty(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*linkNode
		link         doublyLink
	}
	n := &linkNode{}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "isEmpty_empty",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link:         doublyLink{},
			},
			want: true,
		},
		{
			name: "isEmpty_notEmpty",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link: doublyLink{
					head: n,
					tail: n,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := newBeatsUniqueQueueForTest(
				tt.fields.lastBeatsMap,
				tt.fields.link,
			)

			if got := queue.isEmpty(); got != tt.want {
				t.Errorf("beatsUniqueQueue.isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniqueQueue_peek(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*linkNode
		link         doublyLink
	}

	b := &beat{}
	n := &linkNode{data: b}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		{
			name: "peek_nil",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link:         doublyLink{},
			},
			want: nil,
		},
		{
			name: "peek_notNil",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link: doublyLink{
					head: n,
					tail: n,
				},
			},
			want: b,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := newBeatsUniqueQueueForTest(
				tt.fields.lastBeatsMap,
				tt.fields.link,
			)

			if got := queue.peek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniqueQueue.peek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniqueQueue_pop(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*linkNode
		link         doublyLink
	}

	b := &beat{}
	n := &linkNode{data: b}
	tests := []struct {
		name   string
		fields fields
		want   *beat
	}{
		{
			name: "pop_nil",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link:         doublyLink{},
			},
			want: nil,
		},
		{
			name: "pop_notNil",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link: doublyLink{
					head: n,
					tail: n,
				},
			},
			want: b,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := newBeatsUniqueQueueForTest(
				tt.fields.lastBeatsMap,
				tt.fields.link,
			)

			if got := queue.pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniqueQueue.pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beatsUniqueQueue_push(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*linkNode
		link         doublyLink
	}
	type args struct {
		b *beat
	}

	b1 := &beat{key: "service1"}
	n1 := &linkNode{data: b1}

	b2 := &beat{key: "service2"}
	n2 := &linkNode{data: b2}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "push_notExist",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link:         doublyLink{},
			},
			args: args{b1},
		},
		{
			name: "push_exist",
			fields: fields{
				lastBeatsMap: map[string]*linkNode{
					"service1": n1,
				},
				link: doublyLink{
					head: n1,
					tail: n1,
				},
			},
			args: args{b1},
		},
		{
			name: "push_existOther",
			fields: fields{
				lastBeatsMap: map[string]*linkNode{
					"service2": n2,
				},
				link: doublyLink{
					head: n2,
					tail: n2,
				},
			},
			args: args{b1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := newBeatsUniqueQueueForTest(
				tt.fields.lastBeatsMap,
				tt.fields.link,
			)

			if got := queue.push(tt.args.b); got != nil && got.key != tt.args.b.key {
				t.Errorf("beatsUniqueQueue.push() return = %v, want %v", got.key, tt.args.b.key)
			}

			if newbeat, ok := queue.lastBeatsMap[tt.args.b.key]; !ok {
				t.Errorf("beatsUniqueQueue.push() map key not found")
				if newbeat.data != tt.args.b {
					t.Errorf("beatsUniqueQueue.push() map value not equal")
				}
			}

			if !reflect.DeepEqual(queue.link.tail.data, tt.args.b) {
				t.Errorf("beatsUniqueQueue.push()  %v, want %v", queue.link.tail, tt.args.b)
			}
		})
	}
}

func Test_beatsUniqueQueue_remove(t *testing.T) {
	type fields struct {
		lastBeatsMap map[string]*linkNode
		link         doublyLink
	}
	type args struct {
		key string
	}

	b1 := &beat{key: "service1"}
	n1 := &linkNode{data: b1}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *beat
	}{
		{
			name: "remove_empty",
			fields: fields{
				lastBeatsMap: make(map[string]*linkNode),
				link:         doublyLink{},
			},
			args: args{"service1"},
			want: nil,
		},
		{
			name: "remove_notEmpty",
			fields: fields{
				lastBeatsMap: map[string]*linkNode{
					"service1": n1,
				},
				link: doublyLink{
					head: n1,
					tail: n1,
				},
			},
			args: args{"service1"},
			want: n1.data,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := newBeatsUniqueQueueForTest(
				tt.fields.lastBeatsMap,
				tt.fields.link,
			)

			if got := queue.remove(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beatsUniqueQueue.remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_doublyLink(t *testing.T) {
	nodeSlice := []*linkNode{
		{data: &beat{key: "0"}},
		{data: &beat{key: "1"}},
		{data: &beat{key: "2"}},
		{data: &beat{key: "3"}},
		{data: &beat{key: "4"}},
	}

	beatsLink := doublyLink{}
	beatsLink.push(nodeSlice[0])

	if beatsLink.head != nodeSlice[0] || beatsLink.tail != nodeSlice[0] {
		t.Errorf("doublyLink.push() head %v, tail %v, want %v", beatsLink.head.data.key, beatsLink.tail.data.key, nodeSlice[0].data.key)
	}

	beatsLink.push(nodeSlice[1])
	if beatsLink.head != nodeSlice[0] || beatsLink.tail != nodeSlice[1] {
		t.Errorf("doublyLink.push() head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	if beatsLink.head.prev != nil || beatsLink.tail.next != nil || beatsLink.head.next != nodeSlice[1] || beatsLink.tail.prev != nodeSlice[0] {
		t.Errorf("doublyLink.push() pointer error, head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	if peek := beatsLink.peek(); peek != nodeSlice[0] {
		t.Errorf("doublyLink.peek() head %v, want %v", peek.data.key, nodeSlice[0].data.key)
	}

	beatsLink.push(nodeSlice[2])
	beatsLink.push(nodeSlice[3])
	beatsLink.pop()
	if pop := beatsLink.pop(); pop != nodeSlice[1] {
		t.Errorf("doublyLink.pop() head %v, want %v", pop.data.key, nodeSlice[1].data.key)
	}

	if beatsLink.head != nodeSlice[2] || beatsLink.tail != nodeSlice[3] {
		t.Errorf("doublyLink.pop() head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	if beatsLink.head.prev != nil || beatsLink.tail.next != nil || beatsLink.head.next != nodeSlice[3] || beatsLink.tail.prev != nodeSlice[2] {
		t.Errorf("doublyLink.pop() pointer error, head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	beatsLink.push(nodeSlice[4])
	beatsLink.remove(nodeSlice[4])

	if beatsLink.head != nodeSlice[2] || beatsLink.tail != nodeSlice[3] {
		t.Errorf("doublyLink.pop() head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	if beatsLink.head.prev != nil || beatsLink.tail.next != nil || beatsLink.head.next != nodeSlice[3] || beatsLink.tail.prev != nodeSlice[2] {
		t.Errorf("doublyLink.pop() pointer error, head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	beatsLink.push(nodeSlice[4])
	beatsLink.remove(nodeSlice[3])

	if beatsLink.head != nodeSlice[2] || beatsLink.tail != nodeSlice[4] {
		t.Errorf("doublyLink.pop() head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	if beatsLink.head.prev != nil || beatsLink.tail.next != nil || beatsLink.head.next != nodeSlice[4] || beatsLink.tail.prev != nodeSlice[2] {
		t.Errorf("doublyLink.pop() pointer error, head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	beatsLink.pop()
	if beatsLink.head != nodeSlice[4] || beatsLink.tail != nodeSlice[4] {
		t.Errorf("doublyLink.push() head %v, tail %v, want %v", beatsLink.head.data.key, beatsLink.tail.data.key, nodeSlice[4].data.key)
	}

	if beatsLink.head.prev != nil || beatsLink.tail.next != nil || beatsLink.head.next != nil || beatsLink.tail.prev != nil {
		t.Errorf("beatsLink.pop() pointer error, head %v, tail %v", beatsLink.head.data.key, beatsLink.tail.data.key)
	}

	beatsLink.pop()
	if beatsLink.head != nil || beatsLink.tail != nil {
		t.Errorf("doublyLink.push() head %v, tail %v, want %v", beatsLink.head.data.key, beatsLink.tail.data.key, nodeSlice[4].data.key)
	}
}

func newBeatsUniqueQueueForTest(lastBeatsMap map[string]*linkNode, link doublyLink) *beatsUniqueQueue {
	return &beatsUniqueQueue{
		lastBeatsMap: lastBeatsMap,
		link:         link,
		nodePool: sync.Pool{
			New: func() interface{} {
				return &linkNode{}
			},
		},
	}
}

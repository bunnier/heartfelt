package heartfelt

var _ beatsRepository = (*beatsUniquePriorityQueue)(nil)

// beatsUniqueQueue maintenance a beats unique queue for fixed timeout heartbeat.
type beatsUniquePriorityQueue struct {
	lastBeatsMap map[string]int // lastBeatsMap is a map: key => beat.key, value => beat index of heap.
	minHeap      heap           // link is a heartbeat unique queue
}

func newBeatsUniquePriorityQueue() beatsRepository {
	return &beatsUniquePriorityQueue{
		lastBeatsMap: make(map[string]int),
		minHeap:      heap{make([]*beat, 0)},
	}
}

func (queue *beatsUniquePriorityQueue) isEmpty() bool {
	return queue.minHeap.isEmpty()
}

func (queue *beatsUniquePriorityQueue) peek() *beat {
	return queue.minHeap.peek()
}

func (queue *beatsUniquePriorityQueue) pop() *beat {
	beat := queue.minHeap.pop()
	delete(queue.lastBeatsMap, beat.key)
	return beat
}

func (queue *beatsUniquePriorityQueue) push(b *beat) *beat {
	var oldBeat *beat
	if oldIndex, ok := queue.lastBeatsMap[b.key]; ok {
		oldBeat := queue.minHeap.remove(oldIndex)
		if b != oldBeat {
			b.firstTime = oldBeat.firstTime
		} else {
			oldBeat = nil
		}
	}

	newIndex := queue.minHeap.push(b)
	queue.lastBeatsMap[b.key] = newIndex
	return oldBeat
}

func (queue *beatsUniquePriorityQueue) remove(key string) *beat {
	if i, ok := queue.lastBeatsMap[key]; ok {
		delete(queue.lastBeatsMap, key)
		return queue.minHeap.remove(i)
	}
	return nil
}

// heap is a heap of heartbeats
type heap struct {
	items []*beat
}

func (h *heap) isEmpty() bool {
	return len(h.items) == 0
}

func (h *heap) peek() *beat {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}

func (h *heap) pop() *beat {
	if len(h.items) == 0 {
		return nil
	}
	return h.remove(0)
}

func (h *heap) push(b *beat) int {
	h.items = append(h.items, b)
	index := len(h.items) - 1

	for {
		var parentIndex int
		if parentIndex := (index - 1) / 2; parentIndex < 0 || parentIndex == index {
			break
		}

		if h.items[parentIndex].timeoutTime.Before(h.items[index].timeoutTime) {
			break
		}

		// swap with parent node.
		h.items[parentIndex], h.items[index] = h.items[index], h.items[parentIndex]
		index = parentIndex
	}

	return index
}

func (h *heap) remove(index int) *beat {
	if len(h.items) == 0 {
		return nil
	}

	b := h.items[index]
	h.items[index] = h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]

	for {
		var minChildIndex int
		if leftChildIndex := index*2 + 1; leftChildIndex >= len(h.items) {
			break
		} else {
			minChildIndex = leftChildIndex
			if rightChildIndex := leftChildIndex + 1; rightChildIndex < len(h.items) {
				if h.items[leftChildIndex].timeoutTime.After(h.items[rightChildIndex].timeoutTime) {
					minChildIndex = rightChildIndex
				}
			}
		}

		if h.items[index].timeoutTime.Before(h.items[minChildIndex].timeoutTime) {
			break
		}

		// Swap with min child node.
		h.items[minChildIndex], h.items[index] = h.items[index], h.items[minChildIndex]
		index = minChildIndex
	}

	return b
}

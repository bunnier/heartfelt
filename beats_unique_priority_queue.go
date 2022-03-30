package heartfelt

import "sync"

var _ beatsRepository = (*beatsUniquePriorityQueue)(nil)

// beatsUniqueQueue maintenance a beats unique queue for fixed timeout heartbeat.
type beatsUniquePriorityQueue struct {
	lastBeatsMap map[string]*heapNode // lastBeatsMap is a map: key => beat.key, value => beat.
	minHeap      heap                 // link is a heartbeat unique queue
	nodePool     sync.Pool
}

type heapNode struct {
	data      *beat
	heapIndex int // the index of beat on minHeap
}

func newBeatsUniquePriorityQueue() beatsRepository {
	return &beatsUniquePriorityQueue{
		lastBeatsMap: make(map[string]*heapNode),
		minHeap:      heap{make([]*heapNode, 0)},
		nodePool: sync.Pool{
			New: func() interface{} {
				return &heapNode{}
			},
		},
	}
}

func (queue *beatsUniquePriorityQueue) isEmpty() bool {
	return queue.minHeap.isEmpty()
}

func (queue *beatsUniquePriorityQueue) peek() *beat {
	if heapNode := queue.minHeap.peek(); heapNode != nil {
		return heapNode.data
	}
	return nil
}

func (queue *beatsUniquePriorityQueue) pop() *beat {
	if heapNode := queue.minHeap.pop(); heapNode != nil {
		delete(queue.lastBeatsMap, heapNode.data.key)
		return heapNode.data
	}
	return nil
}

func (queue *beatsUniquePriorityQueue) push(b *beat) *beat {
	var oldBeat *beat
	var oldHeapNode *heapNode
	var ok bool

	if oldHeapNode, ok = queue.lastBeatsMap[b.key]; ok {
		queue.minHeap.remove(oldHeapNode.heapIndex)
		oldBeat = oldHeapNode.data
		if b != oldBeat {
			b.firstTime = oldBeat.firstTime
		} else {
			oldBeat = nil
		}
	}

	heapNode := queue.nodePool.Get().(*heapNode)
	heapNode.data = b
	queue.minHeap.push(heapNode)
	queue.lastBeatsMap[b.key] = heapNode
	return oldBeat
}

func (queue *beatsUniquePriorityQueue) remove(key string) *beat {
	if heapNode, ok := queue.lastBeatsMap[key]; ok {
		delete(queue.lastBeatsMap, key)
		heapNode := queue.minHeap.remove(heapNode.heapIndex)

		beat := heapNode.data
		queue.nodePool.Put(heapNode)
		return beat
	}
	return nil
}

// heap is a heap of heartbeats
type heap struct {
	items []*heapNode
}

func (h *heap) isEmpty() bool {
	return len(h.items) == 0
}

func (h *heap) peek() *heapNode {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}

func (h *heap) pop() *heapNode {
	if len(h.items) == 0 {
		return nil
	}
	return h.remove(0)
}

func (h *heap) push(node *heapNode) {
	h.items = append(h.items, node)
	node.heapIndex = len(h.items) - 1

	// Heapify up.
	for {
		currentIndex := node.heapIndex

		var parentIndex int
		if parentIndex = (currentIndex - 1) / 2; parentIndex < 0 || parentIndex == currentIndex {
			break
		}

		if h.items[parentIndex].data.timeoutTime == node.data.timeoutTime ||
			h.items[parentIndex].data.timeoutTime.Before(node.data.timeoutTime) {
			break
		}

		// swap with parent node.
		h.items[parentIndex].heapIndex, node.heapIndex = currentIndex, parentIndex
		h.items[parentIndex], h.items[currentIndex] = node, h.items[parentIndex]
	}
}

func (h *heap) remove(index int) *heapNode {
	if len(h.items) == 0 {
		return nil
	}

	removedNode := h.items[index]

	node := h.items[len(h.items)-1]
	node.heapIndex = index
	h.items[index] = node
	h.items = h.items[:len(h.items)-1]

	// Heapify down.
	for {
		currentIndex := node.heapIndex

		var minChildIndex int
		if minChildIndex = currentIndex*2 + 1; minChildIndex >= len(h.items) {
			break
		} else {
			if rightChildIndex := minChildIndex + 1; rightChildIndex < len(h.items) {
				if h.items[rightChildIndex].data.timeoutTime.Before(h.items[minChildIndex].data.timeoutTime) {
					minChildIndex = rightChildIndex
				}
			}
		}

		if h.items[currentIndex].data.timeoutTime == h.items[minChildIndex].data.timeoutTime ||
			h.items[currentIndex].data.timeoutTime.Before(h.items[minChildIndex].data.timeoutTime) {
			break
		}

		// Swap with min child node.
		h.items[minChildIndex].heapIndex, node.heapIndex = currentIndex, minChildIndex
		h.items[minChildIndex], h.items[currentIndex] = node, h.items[minChildIndex]
	}

	return removedNode
}

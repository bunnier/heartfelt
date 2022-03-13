package heartfelt

import "sync"

var _ beatsRepository = (*beatsUniqueQueue)(nil)

// beatsUniqueQueue maintenance a beats unique queue for fixed timeout heartbeat.
type beatsUniqueQueue struct {
	lastBeatsMap map[string]*linkNode // lastBeatsMap is a map: key => beat.key, value => beat node of link.
	link         doublyLink           // link is a heartbeat unique queue
	nodePool     sync.Pool
}

func newBeatsUniqueQueue() beatsRepository {
	return &beatsUniqueQueue{
		lastBeatsMap: make(map[string]*linkNode),
		link:         doublyLink{},
		nodePool: sync.Pool{
			New: func() interface{} {
				return &linkNode{}
			},
		},
	}
}

func (queue *beatsUniqueQueue) isEmpty() bool {
	return queue.link.head == nil
}

func (queue *beatsUniqueQueue) peek() *beat {
	if queue.isEmpty() {
		return nil
	}
	return queue.link.head.data
}

func (queue *beatsUniqueQueue) pop() *beat {
	if queue.isEmpty() {
		return nil
	}

	n := queue.link.pop()
	b := n.data
	delete(queue.lastBeatsMap, b.key)
	queue.nodePool.Put(n) // Reuse the *linkNode.
	return b
}

func (queue *beatsUniqueQueue) push(b *beat) *beat {
	var node *linkNode
	var ok bool
	var oldBeat *beat
	if node, ok = queue.lastBeatsMap[b.key]; ok {
		queue.link.remove(node)
		if b != node.data {
			oldBeat = node.data
			b.firstTime = oldBeat.firstTime
			node.data = b
		}
	} else {
		node = queue.nodePool.Get().(*linkNode) // Reuse the *linkNode.
		node.data = b
		queue.lastBeatsMap[b.key] = node
	}

	queue.link.push(node)
	return oldBeat
}

func (queue *beatsUniqueQueue) remove(key string) *beat {
	if n, ok := queue.lastBeatsMap[key]; ok {
		delete(queue.lastBeatsMap, key) // Remove the heart from the hearts map.
		queue.link.remove(n)            // Remove relative beat from beatlink.

		d := n.data
		queue.nodePool.Put(n) // Reuse the *linkNode.
		return d
	}
	return nil
}

// doublyLink is a doubly link of heartbeats
type doublyLink struct {
	head *linkNode
	tail *linkNode
}

type linkNode struct {
	data *beat
	prev *linkNode
	next *linkNode
}

func (link *doublyLink) peek() *linkNode {
	return link.head
}

func (link *doublyLink) pop() *linkNode {
	if link.head == nil {
		return nil
	}

	n := link.head
	if link.head == link.tail {
		link.head = nil
		link.tail = nil
	} else {
		n.next.prev = n.prev
		link.head = n.next
	}

	n.prev = nil
	n.next = nil
	return n
}

func (link *doublyLink) push(node *linkNode) bool {
	if node == nil {
		return false
	}

	if link.tail == nil {
		link.head, link.tail = node, node
	} else {
		link.tail.next = node
		node.prev = link.tail
		link.tail = node
	}

	return true
}

func (link *doublyLink) remove(node *linkNode) bool {
	if node == nil {
		return false
	}

	if node == link.head {
		link.head = node.next
	} else {
		node.prev.next = node.next
	}

	if node == link.tail {
		link.tail = node.prev
	} else {
		node.next.prev = node.prev
	}

	node.prev = nil
	node.next = nil

	return true
}

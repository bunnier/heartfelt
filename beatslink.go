package heartfelt

// beatsLink is a doubly linked list.
// It stores all of alive(no timeout) heartbeats in one heartHubParallelism by time sequences.
type beatsLink struct {
	headBeat *beat
	tailBeat *beat
}

// remove a beat from this link.
func (link *beatsLink) remove(b *beat) bool {
	if b == nil {
		return false
	}

	if b == link.headBeat {
		link.headBeat = b.next
	} else {
		b.prev.next = b.next
	}

	if b == link.tailBeat {
		link.tailBeat = b.prev
	} else {
		b.next.prev = b.prev
	}

	b.prev = nil
	b.next = nil

	return true
}

// push a beat to the tail of this link.
func (link *beatsLink) push(b *beat) bool {
	if b == nil {
		return false
	}

	if link.tailBeat == nil {
		link.headBeat, link.tailBeat = b, b
	} else {
		link.tailBeat.next = b
		b.prev = link.tailBeat
		link.tailBeat = b
	}

	return true
}

// pop a beat from the head of this link.
func (link *beatsLink) pop() *beat {
	b := link.headBeat
	link.remove(b)
	return b
}

// peek the head of this link.
func (link *beatsLink) peek() *beat {
	return link.headBeat
}

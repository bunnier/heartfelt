package heartfelt

// beatsLink is a doubly linked list,
// store all of alive(no timeout) heartbeats in one heartHubParallelism by time sequences.
type beatsLink struct {
	headBeat *beat
	tailBeat *beat
}

// remove a beat from link.
func (link *beatsLink) remove(b *beat) bool {
	if b == nil {
		return false
	}

	if b == link.headBeat {
		link.headBeat = b.Next
	} else {
		b.Prev.Next = b.Next
	}

	if b == link.tailBeat {
		link.tailBeat = b.Prev
	} else {
		b.Next.Prev = b.Prev
	}

	return true
}

// push a beat to tail of the link.
func (link *beatsLink) push(b *beat) {
	if link.tailBeat == nil { // So the link must be empty.
		link.tailBeat, link.headBeat = b, b
	} else { // Add current beat to the tail.
		b.Prev = link.tailBeat
		b.Prev.Next = b
		link.tailBeat = b
	}
}

// pop a beat from head of the link.
func (link *beatsLink) pop() *beat {
	if link.headBeat == nil {
		return nil
	}

	h := link.headBeat
	link.headBeat = link.headBeat.Next
	return h
}

// peek the head of the link.
func (link *beatsLink) peek() *beat {
	return link.headBeat
}

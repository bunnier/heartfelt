package heartfelt

type beatLink struct {
	headBeat *beat
	tailBeat *beat
}

// remove a beat from link
func (bLink *beatLink) remove(b *beat) bool {
	if b == nil {
		return false
	}

	if b == bLink.headBeat { // it is the head
		bLink.headBeat = b.Next
	} else {
		b.Prev.Next = b.Next
	}

	if b == bLink.tailBeat { // it is the tail
		bLink.tailBeat = b.Prev
	} else {
		b.Next.Prev = b.Prev
	}

	return true
}

// push a beat to tail of the link
func (bLink *beatLink) push(b *beat) {
	if bLink.tailBeat == nil { // so the link is empty
		bLink.tailBeat, bLink.headBeat = b, b
	} else { // add current beat to the tail
		b.Prev = bLink.tailBeat
		b.Prev.Next = b
		bLink.tailBeat = b
	}
}

// pop a beat from head of the link
func (bLink *beatLink) pop() *beat {
	if bLink.headBeat == nil {
		return nil
	}

	h := bLink.headBeat
	bLink.headBeat = bLink.headBeat.Next
	return h
}

// peek the head of the link
func (bLink *beatLink) peek() *beat {
	return bLink.headBeat
}

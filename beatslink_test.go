package heartfelt

import (
	"testing"
)

func Test_beatsLink(t *testing.T) {
	beatsSlice := []*beat{
		{heart: &heart{key: "0"}},
		{heart: &heart{key: "1"}},
		{heart: &heart{key: "2"}},
		{heart: &heart{key: "3"}},
		{heart: &heart{key: "4"}},
	}

	beatsLink := beatsLink{}
	beatsLink.push(beatsSlice[0])

	if beatsLink.headBeat != beatsSlice[0] || beatsLink.tailBeat != beatsSlice[0] {
		t.Errorf("beatsLink.push() head %v, tail %v, want %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key, beatsSlice[0].heart.key)
	}

	beatsLink.push(beatsSlice[1])
	if beatsLink.headBeat != beatsSlice[0] || beatsLink.tailBeat != beatsSlice[1] {
		t.Errorf("beatsLink.push() head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	if beatsLink.headBeat.prev != nil || beatsLink.tailBeat.next != nil || beatsLink.headBeat.next != beatsSlice[1] || beatsLink.tailBeat.prev != beatsSlice[0] {
		t.Errorf("beatsLink.push() pointer error, head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	if peek := beatsLink.peek(); peek != beatsSlice[0] {
		t.Errorf("beatsLink.peek() head %v, want %v", peek.heart.key, beatsSlice[0].heart.key)
	}

	beatsLink.push(beatsSlice[2])
	beatsLink.push(beatsSlice[3])
	beatsLink.pop()
	if pop := beatsLink.pop(); pop != beatsSlice[1] {
		t.Errorf("beatsLink.pop() head %v, want %v", pop.heart.key, beatsSlice[1].heart.key)
	}

	if beatsLink.headBeat != beatsSlice[2] || beatsLink.tailBeat != beatsSlice[3] {
		t.Errorf("beatsLink.pop() head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	if beatsLink.headBeat.prev != nil || beatsLink.tailBeat.next != nil || beatsLink.headBeat.next != beatsSlice[3] || beatsLink.tailBeat.prev != beatsSlice[2] {
		t.Errorf("beatsLink.pop() pointer error, head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	beatsLink.push(beatsSlice[4])
	beatsLink.remove(beatsSlice[4])

	if beatsLink.headBeat != beatsSlice[2] || beatsLink.tailBeat != beatsSlice[3] {
		t.Errorf("beatsLink.pop() head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	if beatsLink.headBeat.prev != nil || beatsLink.tailBeat.next != nil || beatsLink.headBeat.next != beatsSlice[3] || beatsLink.tailBeat.prev != beatsSlice[2] {
		t.Errorf("beatsLink.pop() pointer error, head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	beatsLink.push(beatsSlice[4])
	beatsLink.remove(beatsSlice[3])

	if beatsLink.headBeat != beatsSlice[2] || beatsLink.tailBeat != beatsSlice[4] {
		t.Errorf("beatsLink.pop() head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	if beatsLink.headBeat.prev != nil || beatsLink.tailBeat.next != nil || beatsLink.headBeat.next != beatsSlice[4] || beatsLink.tailBeat.prev != beatsSlice[2] {
		t.Errorf("beatsLink.pop() pointer error, head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	beatsLink.pop()
	if beatsLink.headBeat != beatsSlice[4] || beatsLink.tailBeat != beatsSlice[4] {
		t.Errorf("beatsLink.push() head %v, tail %v, want %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key, beatsSlice[4].heart.key)
	}

	if beatsLink.headBeat.prev != nil || beatsLink.tailBeat.next != nil || beatsLink.headBeat.next != nil || beatsLink.tailBeat.prev != nil {
		t.Errorf("beatsLink.pop() pointer error, head %v, tail %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key)
	}

	beatsLink.pop()
	if beatsLink.headBeat != nil || beatsLink.tailBeat != nil {
		t.Errorf("beatsLink.push() head %v, tail %v, want %v", beatsLink.headBeat.heart.key, beatsLink.tailBeat.heart.key, beatsSlice[4].heart.key)
	}
}

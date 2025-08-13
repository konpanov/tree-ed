package main

import (
	"fmt"
	"log"
	"slices"
)

type CompositeChange struct {
	changes []Change
}

func (self CompositeChange) Apply(win *Window) {
	for _, change := range self.changes {
		change.Apply(win)
	}
}

func (self CompositeChange) Reverse() Change {
	reversed := CompositeChange{}
	reversed.changes = slices.Clone(self.changes)
	for i := range reversed.changes {
		reversed.changes[i] = reversed.changes[i].Reverse()
	}
	slices.Reverse(reversed.changes)
	return reversed
}

func (self CompositeChange) IsEmpty() bool {
	if len(self.changes) == 0 {
		return true
	}
	for _, change := range self.changes {
		if change.IsEmpty() {
			return true
		}
	}
	return false
}

func NewSwapChange(win *Window, startA int, endA int, startB int, endB int) CompositeChange {
	startA, endA = order(startA, endA)
	startB, endB = order(startB, endB)
	if isIntersection(startA, endA, startB, endB) {
		msg := fmt.Sprintf(
			"Swap change can be created only with nonintersecting regions, but %d, %d, %d, %d is given\n",
			startA, endA, startB, endB,
		)
		if debug {
			log.Panic(msg)
		} else {
			log.Print(msg)
		}
	}
	if startB < startA {
		startA, endA, startB, endB = startB, endB, startA, endA
	}

	a := win.buffer.Content()[startA:endA]
	b := win.buffer.Content()[startB:endB]
	change := CompositeChange{}
	mod1 := NewReplacementChange(startB, b, a)
	change.changes = append(change.changes, mod1)
	mod2 := NewReplacementChange(startA, a, b)
	change.changes = append(change.changes, mod2)
	return change
}

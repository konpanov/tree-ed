package main

import (
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

func (self CompositeChange) Shift(index int) int {
	for _, change := range self.changes {
		index = change.Shift(index)
	}
	return index
}

func (self CompositeChange) IsEmpty() bool {
	for _, change := range self.changes {
		if change.IsEmpty() {
			return true
		}
	}
	return false
}

func NewSwapChange(win *Window, startA int, endA int, startB int, endB int) CompositeChange {
	is_ordered := startA <= endA && endA <= startB && startB <= endB
	if !is_ordered {
		log.Panicf(
			"Swap change can be created only with ordered indices, but %d, %d, %d, %d is given\n",
			startA, endA, startB, endB,
		)
	}

	change := CompositeChange{}

	a := win.buffer.Content()[startA:endA]
	b := win.buffer.Content()[startB:endB]

	mod1 := NewReplacementModification(startB, b, a)
	mod1.cursorBefore = win.cursor.Index()
	mod1.cursorAfter = mod1.Shift(win.cursor.Index())
	change.changes = append(change.changes, mod1)

	mod2 := NewReplacementModification(startA, a, b)
	mod2.cursorBefore = win.cursor.Index()
	mod2.cursorAfter = mod2.Shift(win.cursor.Index())
	change.changes = append(change.changes, mod2)

	return change
}

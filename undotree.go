package main

import (
	"log"
	"slices"
	"unicode/utf8"
)

type Modification interface {
	Apply(buf *Window)
	Reverse() Modification
	IsEmpty() bool
}

type ReplacementModification struct {
	at           int
	before       []byte
	after        []byte
	cursorBefore int
	cursorAfter  int
}

func (self ReplacementModification) Apply(win *Window) {
	win.buffer.Edit(ReplacementInput{
		start:       self.at,
		end:         self.at + len(self.before),
		replacement: self.after,
	})
	var err error
	win.cursor, err = win.cursor.ToIndex(clip(self.cursorAfter, 0, win.buffer.LastIndex()))
	panic_if_error(err)
}

func (self ReplacementModification) Reverse() Modification {
	self.after, self.before = self.before, self.after
	self.cursorAfter, self.cursorBefore = self.cursorBefore, self.cursorAfter
	return self
}

func (self ReplacementModification) IsEmpty() bool {
	return slices.Compare(self.before, self.after) == 0
}

func NewReplacementModification(at int, before []byte, after []byte) ReplacementModification {
	return ReplacementModification{at: at, before: slices.Clone(before), after: slices.Clone(after)}
}

func NewEraseModification(win *Window, start int, end int) ReplacementModification {
	start, end = min(start, end), max(start, end)
	return NewReplacementModification(start, win.buffer.Content()[start:end], []byte{})
}

func NewEraseRuneModification(win *Window, index int) ReplacementModification {
	_, length := utf8.DecodeRune(win.buffer.Content()[index:])
	return NewEraseModification(win, index, index+length)
}

func NewEraseLineModification(win *Window, row int) ReplacementModification {
	buf := win.buffer
	lines := buf.Lines()
	if row < 0 || row >= len(lines) {
		log.Panicf("Cannot erase nonexisting line %d. number of line: %d.", row, len(lines))
	}
	line := lines[row]
	end := min(line.end+len(buf.Nl_seq()), len(buf.Content()))
	return NewEraseModification(win, line.start, end)
}

type CompositeModification struct {
	modifications []Modification
}

func (self CompositeModification) Apply(win *Window) {
	for _, mod := range self.modifications {
		mod.Apply(win)
	}
}

func (self CompositeModification) Reverse() Modification {
	reversed := CompositeModification{}
	reversed.modifications = slices.Clone(self.modifications)
	for i := range reversed.modifications {
		reversed.modifications[i] = reversed.modifications[i].Reverse()
	}
	slices.Reverse(reversed.modifications)
	return reversed
}

func (self CompositeModification) IsEmpty() bool {
	for _, mod := range self.modifications {
		if mod.IsEmpty() {
			return true
		}
	}
	return false
}

type ChangeTree struct {
	buffer        IBuffer
	modifications []Modification
	current       int
}

func (self *ChangeTree) Push(mod Modification) {
	if !mod.IsEmpty() {
		self.modifications = self.modifications[:self.current]
		self.modifications = append(self.modifications, mod)
		self.current++
	}
}

func (self *ChangeTree) Back() Modification {
	if self.current == 0 {
		return nil
	}
	self.current--
	return self.modifications[self.current]
}

func (self *ChangeTree) Forward() Modification {
	if self.current == len(self.modifications) {
		return nil
	}
	self.current++
	return self.modifications[self.current-1]
}

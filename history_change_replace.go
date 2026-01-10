package main

import (
	"log"
	"slices"
	"unicode/utf8"
)

type ReplaceChange struct {
	at int

	before []byte
	after  []byte

	cursorAfter  int
	cursorBefore int

	anchorAfter  int
	anchorBefore int
}

func (self ReplaceChange) Apply(win *Window) {
	win.buffer.Edit(ReplacementInput{
		start:       self.at,
		end:         self.at + len(self.before),
		replacement: self.after,
	})
	win.setCursor(win.cursor.ToIndex(self.cursorAfter), true)
	win.setAnchor(win.anchor.ToIndex(self.anchorAfter))
}

func (self ReplaceChange) Reverse() Change {
	self.after, self.before = self.before, self.after
	self.cursorBefore, self.cursorAfter = self.cursorAfter, self.cursorBefore
	self.anchorBefore, self.anchorAfter = self.anchorAfter, self.anchorBefore
	return self
}

func NewReplacementChange(at int, before []byte, after []byte) ReplaceChange {
	return ReplaceChange{
		at:           at,
		before:       slices.Clone(before),
		after:        slices.Clone(after),
		cursorAfter:  at,
		cursorBefore: at,
		anchorAfter:  at,
		anchorBefore: at,
	}
}

func NewEraseChange(win *Window, start int, end int) ReplaceChange {
	start, end = min(start, end), max(start, end)
	return NewReplacementChange(start, win.buffer.Content()[start:end], []byte{})
}

func NewEraseRuneChange(win *Window, index int) ReplaceChange {
	_, length := utf8.DecodeRune(win.buffer.Content()[index:])
	return NewEraseChange(win, index, index+length)
}

func NewEraseLineChange(win *Window, row int) ReplaceChange {
	buf := win.buffer
	lines := buf.Lines()
	if row < 0 || row >= len(lines) {
		log.Panicf("Cannot erase nonexisting line %d. number of line: %d.", row, len(lines))
	}
	line := lines[row]
	end := len(buf.Content())
	if row+1 < len(lines) {
		end = lines[row+1].start
	}
	return NewEraseChange(win, line.start, end)
}

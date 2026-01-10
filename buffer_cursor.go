package main

import (
	"fmt"
	"unicode/utf8"
)

type BufferCursor struct {
	buffer  IBuffer
	index   int
	as_edge bool
}

var ErrSequenceNotFound = fmt.Errorf("Sequence not found")

func (self BufferCursor) Index() int {
	return self.index
}

func (self BufferCursor) ToIndex(index int) BufferCursor {
	self.index = clip(index, 0, self.buffer.Length())
	line := self.buffer.Lines()[self.Row()]
	self.index = min(self.index, self.LineEnd(line))
	return self
}

func (self BufferCursor) Rune() (rune, int) {
	return utf8.DecodeRune(self.buffer.Content()[self.index:])
}

func (self BufferCursor) Class() RuneClass {
	r, _ := self.Rune()
	return rune_class(r)
}

func (self BufferCursor) AsEdge() BufferCursor {
	self.as_edge = true
	return self.ToIndex(self.index)
}

func (self BufferCursor) AsChar() BufferCursor {
	self.as_edge = false
	return self.ToIndex(self.index)
}

func (self BufferCursor) LineEnd(line Line) int {
	if self.as_edge {
		return line.end
	} else {
		return max(line.start, line.end-1)
	}
}

func (self BufferCursor) Row() int {
	return self.buffer.Row(self.Index())
}

func (self BufferCursor) Pos() Pos {
	return self.buffer.RunePos(self.index)
}

func (self BufferCursor) BytesForward(count int) BufferCursor {
	return self.ToIndex(self.index + count)
}

func (self BufferCursor) BytesBackward(count int) BufferCursor {
	return self.ToIndex(self.index - count)
}

func (self BufferCursor) MoveToRow(row int) BufferCursor {
	lines := self.buffer.Lines()
	row_clipped := clip(row, 0, len(lines)-1)
	line := lines[row_clipped]
	return self.ToIndex(line.start)
}

func (self BufferCursor) MoveToCol(number int) BufferCursor {
	row := self.Row()
	line := self.buffer.Lines()[row]
	width := utf8.RuneCount(self.buffer.Content()[line.start:line.end])
	if !self.as_edge {
		width--
	}
	col := clip(number, 0, max(width, 0))
	rune_pos := Pos{col: col, row: row}
	index := self.buffer.Index(rune_pos)
	cursor := self.ToIndex(index)
	return cursor
}

func (self BufferCursor) MoveToRunePos(rune_pos Pos) BufferCursor {
	self = self.MoveToRow(rune_pos.row)
	self = self.MoveToCol(rune_pos.col)
	return self
}

func (self BufferCursor) RuneNext() BufferCursor {
	if self.IsEnd() {
		return self
	}
	pos := self.Pos()
	line := self.buffer.Lines()[pos.row]
	if pos.col == line.end {
		self.index = line.next_start
		return self
	} else {
		_, width := self.Rune()
		self.index += width
		return self
	}
}

func (self BufferCursor) RunePrev() BufferCursor {
	if self.IsBegining() {
		return self
	}
	pos := self.Pos()
	if pos.col == 0 {
		prev_line := self.buffer.Lines()[pos.row-1]
		self.index = prev_line.end
		return self
	} else {
		_, size := utf8.DecodeLastRune(self.buffer.Content()[:self.index])
		self.index -= size
		return self
	}
}

func (self BufferCursor) WordStartNext() BufferCursor {
	hardstop := self.buffer.Length()
	for prev := self; self.Index() < hardstop; prev, self = self, self.RuneNext() {
		if self.Class() != RuneClassSpace && self.Class() != prev.Class() {
			break
		}
	}
	return self
}

func (self BufferCursor) WordEndNext() BufferCursor {
	hardstop := self.buffer.Length()
	self = self.RuneNext()
	for next := self; self.Index() < hardstop; self, next = next, next.RuneNext() {
		if self.Class() != RuneClassSpace && self.Class() != next.Class() {
			break
		}
	}
	return self
}

func (self BufferCursor) WordStartPrev() BufferCursor {
	self = self.RunePrev()
	for prev := self; self.Index() > 0; self, prev = prev, prev.RunePrev() {
		if self.Class() != RuneClassSpace && self.Class() != prev.Class() {
			break
		}
	}
	return self
}

func (self BufferCursor) WordEndPrev() BufferCursor {
	for next := self; self.Index() > 0; next, self = self, self.RunePrev() {
		if self.Class() != RuneClassSpace && self.Class() != next.Class() {
			break
		}
	}
	return self
}

func (self BufferCursor) ToLineEnd() BufferCursor {
	line := self.buffer.Lines()[self.Row()]
	return self.ToIndex(self.LineEnd(line))
}

func (self BufferCursor) ToLineStart() BufferCursor {
	line := self.buffer.Lines()[self.Row()]
	return self.ToIndex(line.start)
}

func (self BufferCursor) ToLineTextStart() BufferCursor {
	line := self.buffer.Lines()[self.Row()]
	self = self.ToIndex(line.start)
	for self.Class() == RuneClassSpace && self.Index() <= self.LineEnd(line) {
		self = self.RuneNext()
	}
	return self
}

func (self BufferCursor) Match(seq []byte) bool {
	return matchBytes(self.buffer.Content()[self.index:], seq)
}

func (self BufferCursor) IsLineBreak() bool {
	is_line_break, _ := IsLineBreak(self.buffer.Content()[self.index:])
	return is_line_break
}

func (self BufferCursor) IsLineStart() bool {
	return self.Pos().col == 0
}

func (self BufferCursor) IsEnd() bool {
	return self.Index() >= self.buffer.Length()
}

func (self BufferCursor) IsBegining() bool {
	return self.Index() == 0
}

func (self BufferCursor) SearchForward(seq []byte) (BufferCursor, error) {
	cursor := self
	hardstop := cursor.buffer.Length()
	for cursor.Index() != hardstop {
		cursor = cursor.BytesForward(1)
		if cursor.Match(seq) {
			return cursor, nil
		}
	}
	return self, ErrSequenceNotFound
}

func (self BufferCursor) SearchBackward(seq []byte) (BufferCursor, error) {
	cursor := self
	for cursor.Index() != 0 {
		cursor = cursor.BytesBackward(1)
		if cursor.Match(seq) {
			return cursor, nil
		}
	}
	return self, ErrSequenceNotFound
}

func (self BufferCursor) Update(edit ReplacementInput) BufferCursor {
	if self.Index() >= edit.end {
		offset := edit.start - edit.end + len(edit.replacement)
		return self.ToIndex(self.Index() + offset)
	} else if self.Index() > edit.start {
		edit_end := edit.start + len(edit.replacement)
		return self.ToIndex(min(self.Index(), edit_end))
	} else {
		return self
	}
}

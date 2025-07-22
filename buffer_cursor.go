package main

import (
	"fmt"
	"log"
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

func (self BufferCursor) Rune() rune {
	value, _ := utf8.DecodeRune(self.buffer.Content()[self.index:])
	return value
}

func (self BufferCursor) Class() RuneClass {
	return rune_class(self.Rune())
}

func (self BufferCursor) Hardstop() int {
	hardstop := self.buffer.LastIndex()
	if self.as_edge {
		hardstop++
	}
	return hardstop
}

func (self BufferCursor) Rowend(row Region) int {
	if self.as_edge {
		return row.end
	} else {
		return max(row.start, row.end-1)
	}
}

func (self BufferCursor) Row() int {
	row, err := self.buffer.Row(self.Index())
	if err != nil {
		log.Panicln(err)
	}
	return row
}

func (self BufferCursor) BytePosition() Point {
	coord, err := self.buffer.Coord(self.index)
	if err != nil {
		log.Fatalf("Could not find buffer coord at index %d\n", self.index)
	}
	return coord
}

func (self BufferCursor) RunePosition() Point {
	coord, err := self.buffer.RuneCoord(self.index)
	if err != nil {
		log.Fatalf("Could not find buffer coord at index %d\n", self.index)
	}
	return coord
}

func (self BufferCursor) ToIndex(index int) BufferCursor {
	self.index = clip(index, 0, self.Hardstop())
	return self
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
	index, err := self.buffer.IndexFromRuneCoord(Point{col: col, row: row})
	panic_if_error(err)
	cursor := self.ToIndex(index)
	return cursor
}

func (self BufferCursor) MoveToRunePos(pos Point) BufferCursor {
	self = self.MoveToRow(pos.row)
	self = self.MoveToCol(pos.col)
	return self
}

func (self BufferCursor) RuneNext() BufferCursor {
	if !self.IsEnd() {
		self.index += utf8.RuneLen(self.Rune())
	}
	return self
}

func (self BufferCursor) RunePrev() BufferCursor {
	if !self.IsBegining() {
		_, size := utf8.DecodeLastRune(self.buffer.Content()[:self.index])
		self.index -= size
	}
	return self
}

func (self BufferCursor) WordStartNext() BufferCursor {
	hardstop := self.buffer.LastIndex()
	for prev := self; self.Index() < hardstop; prev, self = self, self.RuneNext() {
		if self.Class() != RuneClassSpace && self.Class() != prev.Class() {
			break
		}
	}
	return self
}

func (self BufferCursor) WordEndNext() BufferCursor {
	hardstop := self.buffer.LastIndex()
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

func (self BufferCursor) ToRowEnd() BufferCursor {
	row, err := self.buffer.Line(self.Row())
	panic_if_error(err)
	return self.ToIndex(self.Rowend(row))
}

func (self BufferCursor) ToRowStart() BufferCursor {
	row, err := self.buffer.Line(self.Row())
	panic_if_error(err)
	return self.ToIndex(row.start)
}

func (self BufferCursor) ToRowTextStart() BufferCursor {
	row, err := self.buffer.Line(self.Row())
	panic_if_error(err)
	self = self.ToIndex(row.start)
	for self.Class() == RuneClassSpace && self.Index() <= self.Rowend(row) {
		self = self.RuneNext()
	}
	return self
}

func (self BufferCursor) Match(seq []byte) bool {
	return matchBytes(self.buffer.Content()[self.index:], seq)
}

func (self BufferCursor) IsNewLine() bool {
	return self.Match(self.buffer.Nl_seq())
}

func (self BufferCursor) IsLineStart() bool {
	return self.BytePosition().col == 0
}

func (self BufferCursor) IsEnd() bool {
	return self.Index() == len(self.buffer.Content())
}

func (self BufferCursor) IsBegining() bool {
	return self.Index() == 0
}

func (self BufferCursor) SearchForward(seq []byte) (BufferCursor, error) {
	cursor := self
	hardstop := cursor.Hardstop()
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

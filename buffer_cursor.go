package main

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"
)

type BufferCursor struct {
	buffer IBuffer
	index  int
}

var ErrReachBufferEnd = fmt.Errorf("Reached buffer end")
var ErrReachBufferBeginning = fmt.Errorf("Reached buffer beginning")
var ErrRuneError = fmt.Errorf("Unrecognized rune")
var ErrLastWord = fmt.Errorf("Alread on the last word")
var ErrFirstWord = fmt.Errorf("Alread on the first word")
var ErrNodeNotFound = fmt.Errorf("Failed to find closes node")
var ErrSequenceNotFount = fmt.Errorf("Sequence not found")

func NewBufferCursor(buffer IBuffer) BufferCursor {
	return BufferCursor{buffer: buffer, index: 0}
}

func (self BufferCursor) Index() int {
	return self.index
}

func (self BufferCursor) Rune() rune {
	value, _ := utf8.DecodeRune(self.buffer.Content()[self.index:])
	return value
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

func (self BufferCursor) ToIndex(index int) (BufferCursor, error) {
	if err := self.buffer.CheckIndex(index); err == nil {
		return BufferCursor{buffer: self.buffer, index: index}, nil
	} else {
		return self, err
	}
}

func (self BufferCursor) BytesForward(count int) (BufferCursor, error) {
	return self.ToIndex(self.index + count)
}

func (self BufferCursor) BytesBackward(count int) (BufferCursor, error) {
	return self.ToIndex(self.index - count)
}

func (self BufferCursor) MoveToRow(number int) (BufferCursor, error) {
	lines := self.buffer.Lines()
	number = clip(number, 0, len(lines)-1)
	line := lines[number]
	return BufferCursor{self.buffer, line.start}, nil
}

func (self BufferCursor) MoveToCol(number int) (BufferCursor, error) {
	pos := self.RunePosition()
	line := self.buffer.Lines()[pos.row]
	width := utf8.RuneCount(self.buffer.Content()[line.start:line.end])
	pos.col = clip(number, 0, max(width-1, 0))
	index, err := self.buffer.IndexFromRuneCoord(pos)
	panic_if_error(err)
	return self.ToIndex(index)
}

func (self BufferCursor) MoveToRunePos(pos Point) (BufferCursor, error) {
	var err error
	self, err = self.MoveToRow(pos.row)
	panic_if_error(err)
	self, err = self.MoveToCol(pos.col)
	panic_if_error(err)
	return self, nil
}

func (self BufferCursor) VerticalShift(number int, column_anchor int) (BufferCursor, error) {
	lines := self.buffer.Lines()
	pos := self.RunePosition()
	pos.row += number
	if pos.row < 0 || pos.row >= len(lines) {
		return self, nil
	}
	line := lines[pos.row]
	width := utf8.RuneCount(self.buffer.Content()[line.start:line.end])
	pos.col = clip(column_anchor, 0, width-1)
	index, err := self.buffer.IndexFromRuneCoord(pos)
	if err != nil {
		return self, err
	}
	return self.ToIndex(index)
}

// Treats incorrect runes as a 1 byte rune
func (self BufferCursor) RunesForward(count int) (BufferCursor, error) {
	if self.IsEnd() {
		return self, ErrReachBufferEnd
	}
	sum_size := 0
	for i := 0; i < count; i++ {
		value, size := utf8.DecodeRune(self.buffer.Content()[self.index+sum_size:])
		if value == utf8.RuneError && size == 0 {
			return self, ErrReachBufferEnd
		}
		sum_size += size
	}
	return self.BytesForward(sum_size)
}

// Treats incorrect runes as a 1 byte rune
func (self BufferCursor) RunesBackward(count int) (BufferCursor, error) {
	if self.IsBegining() {
		return self, ErrReachBufferBeginning
	}
	sum_size := 0
	for i := 0; i < count; i++ {
		value, size := utf8.DecodeLastRune(self.buffer.Content()[:self.index-sum_size])
		if value == utf8.RuneError && size == 0 {
			return self, ErrReachBufferBeginning
		}
		sum_size += size
	}
	return self.BytesBackward(sum_size)
}

func (self BufferCursor) WordStartForward() (BufferCursor, error) {
	var err error
	lines := self.buffer.Lines()
	self, err = self.SkipRuneClassForward()
	if err == nil {
		self, err = self.SkipSpace(lines[len(lines)-1].end)
	}
	return self, err
}

func (self BufferCursor) WordEndForward() (BufferCursor, error) {
	var err error
	lines := self.buffer.Lines()
	self, err = self.SkipSpace(lines[len(lines)-1].end)
	if err == nil {
		self, err = self.SkipRuneClassForward()
	}
	return self, err
}

func (self BufferCursor) SkipSpace(stop int) (BufferCursor, error) {
	var err error
	for err == nil && self.Index() < stop && unicode.IsSpace(self.Rune()) {
		self, err = self.RunesForward(1)
	}
	return self, nil
}

func (self BufferCursor) SkipRuneClassForward() (BufferCursor, error) {
	var err error
	prev := self
	for err == nil && rune_class(self.Rune()) == rune_class(prev.Rune()) {
		prev = self
		self, err = self.RunesForward(1)
	}
	return self, err
}

func (self BufferCursor) WordStartBackward() (BufferCursor, error) {
	var err error
	self, err = self.RunesBackward(1)
	for err == nil && unicode.IsSpace(self.Rune()) {
		self, err = self.RunesBackward(1)
	}
	prev := self
	for err == nil && rune_class(self.Rune()) == rune_class(prev.Rune()) {
		self = prev
		prev, err = self.RunesBackward(1)
	}
	return self, err
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
	var err error
	cursor := self
	for {
		cursor, err = cursor.BytesForward(1)
		if err != nil {
			return self, fmt.Errorf("%w: %s", ErrSequenceNotFount, seq)
		}
		if cursor.Match(seq) {
			return cursor, nil
		}
	}
}

func (self BufferCursor) SearchBackward(seq []byte) (BufferCursor, error) {
	var err error
	cursor := self
	for {
		cursor, err = cursor.BytesBackward(1)
		if err != nil {
			return self, fmt.Errorf("%w: %s", ErrSequenceNotFount, seq)
		}
		if cursor.Match(seq) {
			return cursor, nil
		}
	}
}

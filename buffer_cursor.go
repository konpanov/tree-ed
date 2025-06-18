package main

import (
	"fmt"
	"log"
	"unicode/utf8"
)

// TODO Make updatable on buffer changes
// TODO Track their own lines and columns in bytes and runes
type BufferCursor struct {
	buffer IBuffer
	index  int
}

var ErrReachBufferEnd = fmt.Errorf("Reached buffer end")
var ErrRuneError = fmt.Errorf("Unrecognized rune")

func NewBufferCursor(buffer IBuffer) BufferCursor {
	return BufferCursor{buffer: buffer, index: 0}
}

func (self BufferCursor) Index() int {
	return self.index
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
	if self.IsEnd() {
		return self, ErrReachBufferEnd
	}
	sum_size := 0
	for i := 0; i < count; i++ {
		value, size := utf8.DecodeLastRune(self.buffer.Content()[:self.index-sum_size])
		if value == utf8.RuneError && size == 0 {
			return self, ErrReachBufferEnd
		}
		sum_size += size
	}
	return self.BytesBackward(sum_size)
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

var ErrSequenceNotFount = fmt.Errorf("Sequence not found")

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

func (self BufferCursor) UpdateToChange(change BufferChange) (BufferCursor, error) {
	log.Println(change)
	if change.start_index <= self.Index() {
		offset := len(change.after) - len(change.before)
		if offset > 0 {
			return self.BytesForward(offset)
		} else {
			return self.BytesBackward(-offset)
		}
	}
	return self, nil
}

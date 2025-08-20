package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

type ScannerState struct {
	keys  []*tcell.EventKey
	curr  int
	start int
}

type ScanResult int

const (
	ScanFull ScanResult = iota
	ScanNone
	ScanStop
)

type ScanFunc func() ScanResult

func (self *ScannerState) Push(ev tcell.Event) {
	switch value := ev.(type) {
	case *tcell.EventKey:
		self.keys = append(self.keys, value)
	default:
		log.Printf("Scanner: ignoring non key event %+v\n", ev)
	}
}

func (self *ScannerState) Peek() *tcell.EventKey {
	return self.keys[self.curr]
}

func (self *ScannerState) Advance() *tcell.EventKey {
	curr := self.Peek()
	self.curr++
	return curr
}

func (self *ScannerState) IsEnd() bool {
	return self.curr >= len(self.keys)
}

func (self *ScannerState) Input() []*tcell.EventKey {
	return self.keys[self.start:]
}

func (self *ScannerState) Scanned() []*tcell.EventKey {
	return self.keys[self.start:self.curr]
}

func (self *ScannerState) Clear() {
	self.start = self.curr
}

func (self *ScannerState) Reset() {
	self.curr = self.start
}

func (self *ScannerState) ScanWithCondition(cond func() bool) ScanResult {
	if self.IsEnd() {
		return ScanStop
	} else if cond() {
		self.Advance()
		return ScanFull
	} else {
		return ScanNone
	}
}

func (self *ScannerState) ScanZeroOrMore(scan ScanFunc) ScanResult {
	for {
		switch scan() {
		case ScanStop:
			return ScanStop
		case ScanNone:
			return ScanFull
		}
	}
}

func (self *ScannerState) ScanKey(key tcell.Key) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Key() == key
	})
}

func (self *ScannerState) ScanRune(r rune) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Rune() == r
	})
}

func (self *ScannerState) ScanDigit() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsDigitKey(self.Peek())
	})
}

func (self *ScannerState) ScanTextInput() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsTextInputKey(self.Peek())
	})
}

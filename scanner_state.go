package main

import (
	"github.com/gdamore/tcell/v2"
)

type ScanResult int

const (
	ScanFull ScanResult = iota
	ScanNone
	ScanStop
)

type ScanFunc func() ScanResult

func (self *Scanner) Peek() *tcell.EventKey {
	return self.keys[self.curr]
}

func (self *Scanner) Advance() *tcell.EventKey {
	curr := self.Peek()
	self.curr++
	return curr
}

func (self *Scanner) IsEnd() bool {
	return self.curr >= len(self.keys)
}

func (self *Scanner) Input() []*tcell.EventKey {
	return self.keys[self.start:]
}

func (self *Scanner) Scanned() []*tcell.EventKey {
	return self.keys[self.start:self.curr]
}

func (self *Scanner) Clear() {
	self.start = self.curr
}

func (self *Scanner) Reset() {
	self.curr = self.start
}

func (self *Scanner) ScanWithCondition(cond func() bool) ScanResult {
	if self.IsEnd() {
		return ScanStop
	} else if cond() {
		self.Advance()
		return ScanFull
	} else {
		return ScanNone
	}
}

func (self *Scanner) ScanZeroOrMore(scan ScanFunc) ScanResult {
	for {
		switch scan() {
		case ScanStop:
			return ScanStop
		case ScanNone:
			return ScanFull
		}
	}
}

func (self *Scanner) ScanKey(key tcell.Key) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Key() == key
	})
}

func (self *Scanner) ScanRune(r rune) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Rune() == r
	})
}

func (self *Scanner) ScanDigit() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsDigitKey(self.Peek())
	})
}

func (self *Scanner) ScanTextInput() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsTextInputKey(self.Peek())
	})
}

package main

import (
	"github.com/gdamore/tcell/v2"
)

type InsertScanner struct {
	state      *ScannerState
	continuous bool
	change     ReplacementInput
	input      []byte
}

func (self *InsertScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

// TODO Make erasing after insert continuous (single modification, single undo)
func (self *InsertScanner) Scan() (Operation, error) {
	if self.state == nil {
		self.state = &ScannerState{}
	}
	self.state.Reset()

	op, err := self.ScanOperation()
	self.state.Clear()
	return op, err
}

func (self *InsertScanner) ScanOperation() (Operation, error) {
	ek, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
	input, err := self.ScanInput()
	switch {
	case input != nil:
		self.input = append(self.input, input...)
		op = InsertContentOperation{
			content:              self.input,
			continue_last_insert: self.continuous,
		}
		self.continuous = true
	case ek.Key() == tcell.KeyEsc:
		self.continuous = false
		self.state.Advance()
		op = SwitchFromInsertToNormalMode{}
	case ek.Key() == tcell.KeyBackspace, ek.Key() == tcell.KeyBackspace2:
		op = EraseCharInsertMode{continue_last_erase: self.continuous}
		self.state.Advance()
		self.continuous = false
		// case ek.Key() == tcell.KeyEnter:
		// 	log.Println("Scanned KeyEntr")
		// 	self.state.Advance()
		// 	self.continuous = true
		// 	op = InsertNewLine{}
	}
	if !self.continuous {
		self.input = []byte{}
	}
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

func (self *InsertScanner) ScanInput() ([]byte, error) {
	input := []byte{}
	ek, err := self.state.Curr()
	for err == nil {
		if ek.Key() == tcell.KeyRune {
			input = append(input, []byte(string(ek.Rune()))...)
		} else if ek.Key() == tcell.KeyTab {
			input = append(input, '\t')
		} else if ek.Key() == tcell.KeyCR {
			input = append(input, '\r')
		} else if ek.Key() == tcell.KeyLF {
			input = append(input, '\n')
		} else {
			break
		}
		ek, err = self.state.Advance()
	}
	if len(input) == 0 {
		input = nil
	}
	return input, err
}

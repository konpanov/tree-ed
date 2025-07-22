package main

import "github.com/gdamore/tcell/v2"

type InsertScanner struct {
	state      *ScannerState
	continuous bool
	change     ReplacementInput
	input      []byte
}

// TODO Make erasing after insert continuous (single modification, single undo)
func (self *InsertScanner) Scan(ev tcell.Event) (Operation, error) {
	if self.state == nil {
		self.state = &ScannerState{}
	}
	if err := self.state.Push(ev); err != nil {
		return nil, err
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
	switch {
	case ek.Key() == tcell.KeyRune:
		if !self.continuous {
			self.input = []byte{}
		}
		self.input = append(self.input, string(ek.Rune())...)
		op = InsertContent{
			content:              self.input,
			continue_last_insert: self.continuous,
		}
		self.continuous = true
	case ek.Key() == tcell.KeyEsc:
		self.continuous = false
		op = SwitchFromInsertToNormalMode{}
	case ek.Key() == tcell.KeyBackspace, ek.Key() == tcell.KeyBackspace2:
		op = EraseCharInsertMode{continue_last_erase: self.continuous}
		self.continuous = false
	case ek.Key() == tcell.KeyEnter:
		self.continuous = false
		op = InsertNewLine{}
	}
	self.state.Advance()
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

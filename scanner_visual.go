package main

import "github.com/gdamore/tcell/v2"

type VisualScanner struct {
	state *ScannerState
}

func (self *VisualScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *VisualScanner) Scan() (Operation, error) {
	if self.state == nil {
		self.state = &ScannerState{}
	}
	self.state.Reset()

	count, count_err := ScanCount(self.state)
	op, err := self.ScanOperation()
	if count_err != nil {
		self.state.Clear()
		return op, err
	} else if err == nil {
		self.state.Clear()
		return CountOperation{op: op, count: count}, nil
	} else if err == ErrNoKey {
		return nil, ErrNoKey
	} else {
		self.state.Clear()
		return nil, err
	}
}

func (self *VisualScanner) ScanOperation() (Operation, error) {
	ek, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
	if op, err = self.state.ScanCursorMovement(); op != nil {
	} else {
		switch {
		case ek.Key() == tcell.KeyEsc:
			self.state.Advance()
			op = SwitchToNormalMode{}
		case ek.Rune() == 'i':
			self.state.Advance()
			op = SwitchToInsertMode{}
		case ek.Rune() == 'a':
			self.state.Advance()
			op = SwitchToInsertModeAsAppend{}
		case ek.Rune() == 'v':
			self.state.Advance()
			op = SwitchToNormalMode{}
		case ek.Rune() == 'd':
			self.state.Advance()
			op = EraseSelectionOperation{}
		case ek.Rune() == 't':
			self.state.Advance()
			op = SwitchFromVisualToTreeMode{}
		case ek.Rune() == 'y':
			self.state.Advance()
			op = CopyToClipboardOperation{}
		}
	}
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

package main

import "github.com/gdamore/tcell/v2"

type VisualScanner struct {
	state *ScannerState
}

func (self *VisualScanner) Scan(ev tcell.Event) (Operation, error) {
	if self.state == nil {
		self.state = &ScannerState{}
	}
	if err := self.state.Push(ev); err != nil {
		return nil, err
	}
	self.state.Reset()

	count, count_err := ScanInteger(self.state)
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
	switch {
	case ek.Key() == tcell.KeyEsc:
		op = SwitchToNormalMode{}
	case ek.Rune() == 'i':
		op = SwitchToInsertMode{}
	case ek.Rune() == 'a':
		op = SwitchToInsertModeAsAppend{}
	case ek.Rune() == 'v':
		op = SwitchToNormalMode{}
	case ek.Rune() == 'd':
		op = EraseSelectionOperation{}
	case ek.Rune() == 'j':
		op = NormalCursorDown{}
	case ek.Rune() == 'k':
		op = NormalCursorUp{}
	case ek.Rune() == 'h':
		op = NormalCursorLeft{}
	case ek.Rune() == 'l':
		op = NormalCursorRight{}
	case ek.Rune() == 't':
		op = SwitchFromVisualToTreeMode{}
	case ek.Rune() == 'w':
		op = WordStartForwardOperation{}
	case ek.Rune() == 'e':
		op = WordEndForwardOperation{}
	case ek.Rune() == 'b':
		op = WordBackwardOperation{}
	}
	self.state.Advance()
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

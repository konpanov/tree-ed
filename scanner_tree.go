package main

import "github.com/gdamore/tcell/v2"

type TreeScanner struct {
	state *ScannerState
}

func (self *TreeScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *TreeScanner) Scan() (Operation, error) {
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

func (self *TreeScanner) ScanOperation() (Operation, error) {
	ek, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
	switch {
	case ek.Key() == tcell.KeyEsc:
		op = SwitchToNormalMode{}
	case ek.Rune() == 't':
		op = SwitchToNormalMode{}
	case ek.Rune() == 'k':
		op = NodeUpOperation{}
	case ek.Rune() == 'j':
		op = NodeDownOperation{}
	case ek.Rune() == 'H':
		op = NodePrevSiblingOperation{}
	case ek.Rune() == 'L':
		op = NodeNextSiblingOperation{}
	case ek.Rune() == 'h':
		op = NodePrevCousinOperation{}
	case ek.Rune() == 'l':
		op = NodeNextCousinOperation{}
	case ek.Rune() == 'd':
		op = EraseSelectionOperation{}
	case ek.Rune() == 'f':
		op = SwapNodeForwardEndOperation{}
	case ek.Rune() == 'b':
		op = SwapNodeBackwardEndOperation{}
	case ek.Rune() == '$':
		op = NodeLastSiblingOperation{}
	case ek.Rune() == '_', ek.Rune() == '0':
		op = NodeFirstSiblingOperation{}
	}
	self.state.Advance()
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

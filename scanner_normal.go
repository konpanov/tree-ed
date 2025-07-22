package main

import (
	"github.com/gdamore/tcell/v2"
)

type NormalScanner struct {
	state *ScannerState
}

func (self *NormalScanner) Scan(ev tcell.Event) (Operation, error) {
	if self.state == nil {
		self.state = &ScannerState{}
	}
	if err := self.state.Push(ev); err != nil {
		return nil, err
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

func (self *NormalScanner) ScanOperation() (Operation, error) {
	ek, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
	self.state.Advance()
	switch {
	case ek.Rune() == 'j':
		op = NormalCursorDown{}
	case ek.Rune() == 'k':
		op = NormalCursorUp{}
	case ek.Rune() == 'h':
		op = NormalCursorLeft{}
	case ek.Rune() == 'l':
		op = NormalCursorRight{}
	case ek.Rune() == 'w':
		op = WordStartForwardOperation{}
	case ek.Rune() == 'b':
		op = WordBackwardOperation{}
	case ek.Rune() == 'e':
		op = WordEndForwardOperation{}
	case ek.Rune() == 'E':
		op = WordEndBackwardOperation{}
	case ek.Rune() == 'g':
		op = GoOperation{}
	case ek.Rune() == '$':
		op = LineEndOperation{}
	case ek.Rune() == '0':
		op = LineStartOperation{}
	case ek.Rune() == '_':
		op = LineTextStartOperation{}
	// Modification
	case ek.Rune() == 'd':
		op = EraseLineAtCursor{}
	case ek.Rune() == 'x':
		op = EraseCharNormalMode{}
	// Modes
	case ek.Rune() == 'a':
		op = SwitchToInsertModeAsAppend{}
	case ek.Rune() == 'i':
		op = SwitchToInsertMode{}
	case ek.Rune() == 'v':
		op = SwitchToVisualmode{}
	case ek.Rune() == 't':
		op = SwitchToTreeMode{}
	case ek.Rune() == 'p':
		op = GetClipboardOperation{}
	case ek.Rune() == 'u':
		op = UndoChangeOperation{}
	case ek.Key() == tcell.KeyCtrlR:
		op = RedoChangeOperation{}
	}
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

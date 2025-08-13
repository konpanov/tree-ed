package main

import (
	"github.com/gdamore/tcell/v2"
)

type NormalScanner struct {
	state *ScannerState
}

func (self *NormalScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *NormalScanner) Scan() (Operation, error) {
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

func (self *ScannerState) ScanCursorMovement() (Operation, error) {
	ek, err := self.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
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
	case ek.Rune() == 'G':
		op = GoEndOperation{}
	case ek.Rune() == '$':
		op = LineEndOperation{}
	case ek.Rune() == '0':
		op = LineStartOperation{}
	case ek.Rune() == '_':
		op = LineTextStartOperation{}
	}
	if op != nil {
		self.Advance()
		return op, err
	}
	return nil, ErrNoMatch
}

func (self *NormalScanner) ScanOperation() (Operation, error) {
	ek, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	var op Operation
	if op, err = self.state.ScanCursorMovement(); op != nil {
	} else {
		self.state.Advance()
		switch {
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
			op = PasteClipboardOperation{}
		case ek.Rune() == 'u':
			op = UndoChangeOperation{}
		case ek.Key() == tcell.KeyCtrlR:
			op = RedoChangeOperation{}
		case ek.Rune() == 's':
			op = DeleteSelectionAndInsert{}
		}
	}
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

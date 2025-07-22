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

func (self *NormalScanner) ScanOperation() (Operation, error) {
	key_event, err := self.state.Curr()
	if err != nil {
		return nil, err
	}
	self.state.Advance()
	if key_event.Key() == tcell.KeyRune {
		switch key_event.Rune() {
		// Navigation
		case 'j':
			return NormalCursorDown{}, nil
		case 'k':
			return NormalCursorUp{}, nil
		case 'h':
			return NormalCursorLeft{}, nil
		case 'l':
			return NormalCursorRight{}, nil
		case 'w':
			return WordStartForwardOperation{}, nil
		case 'b':
			return WordBackwardOperation{}, nil
		case 'e':
			return WordEndForwardOperation{}, nil
		case 'E':
			return WordEndBackwardOperation{}, nil
		case 'g':
			return GoOperation{}, nil
		// Modification
		case 'd':
			return EraseLineAtCursor{}, nil
		case 'x':
			return EraseCharNormalMode{}, nil
		// Modes
		case 'a':
			return SwitchToInsertModeAsAppend{}, nil
		case 'i':
			return SwitchToInsertMode{}, nil
		case 'v':
			return SwitchToVisualmode{}, nil
		case 't':
			return SwitchToTreeMode{}, nil
		case 'u':
			return UndoChangeOperation{}, nil
		}
	}
	switch key_event.Key() {
	case tcell.KeyCtrlR:
		return RedoChangeOperation{}, nil
	}
	return nil, ErrNoMatch
}

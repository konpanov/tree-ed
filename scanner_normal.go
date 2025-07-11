package main

import (
	"github.com/gdamore/tcell/v2"
)

type NormalScanner struct {
	history []*tcell.EventKey
	curr    int
}

func (self *NormalScanner) Scan(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	self.history = append(self.history, key_event)

	self.curr = 0
	count, count_err := ScanInteger(self)
	op, err := self.ScanOperation()
	if count_err != nil {
		self.Clear()
		return op, err
	} else if err == nil {
		self.Clear()
		return CountOperation{op: op, count: count}, nil
	} else if err == ErrNoKey {
		return nil, ErrNoKey
	} else {
		self.Clear()
		return nil, err
	}
}

func (self *NormalScanner) Clear() {
	self.history = self.history[self.curr:]
}

func (self *NormalScanner) Advance() (*tcell.EventKey, error) {
	if self.curr < len(self.history) {
		self.curr++
	} else {
		return nil, ErrNoKey
	}
	return self.Curr()
}

func (self *NormalScanner) Curr() (*tcell.EventKey, error) {
	if self.curr >= len(self.history) {
		return nil, ErrNoKey
	}
	return self.history[self.curr], nil
}

func (self *NormalScanner) ScanOperation() (Operation, error) {
	key_event, err := self.Curr()
	if err != nil {
		return nil, err
	}
	if key_event.Key() == tcell.KeyRune {
		switch key_event.Rune() {
		// Navigation
		case 'j':
			self.Advance()
			return NormalCursorDown{}, nil
		case 'k':
			self.Advance()
			return NormalCursorUp{}, nil
		case 'h':
			self.Advance()
			return NormalCursorLeft{}, nil
		case 'l':
			self.Advance()
			return NormalCursorRight{}, nil
		case 'w':
			self.Advance()
			return WordStartForwardOperation{}, nil
		case 'e':
			self.Advance()
			return WordEndForwardOperation{}, nil
		case 'b':
			self.Advance()
			return WordBackwardOperation{}, nil
		case 'g':
			self.Advance()
			return GoOperation{}, nil
		// Modification
		case 'd':
			self.Advance()
			return EraseLineAtCursor{}, nil
		case 'x':
			self.Advance()
			return EraseCharNormalMode{}, nil
		// Modes
		case 'a':
			self.Advance()
			return SwitchToInsertModeAsAppend{}, nil
		case 'i':
			self.Advance()
			return SwitchToInsertMode{}, nil
		case 'v':
			self.Advance()
			return SwitchToVisualmode{}, nil
		case 't':
			self.Advance()
			return SwitchToTreeMode{}, nil
		case 'u':
			self.Advance()
			return UndoChangeOperation{}, nil
		}
	}
	switch key_event.Key() {
	case tcell.KeyCtrlR:
		self.Advance()
		return RedoChangeOperation{}, nil
	}
	self.Advance()
	return nil, ErrNoMatch
}

func (self *NormalScanner) ParseKeySequence(seq []*tcell.EventKey, op Operation) (Operation, error) {
	for i := 0; i < len(seq); i++ {
		key_event, err := self.Curr()
		if err != nil {
			return nil, err
		}
		if key_event.Key() != seq[i].Key() || key_event.Rune() == seq[i].Rune() {
			return nil, ErrNoMatch
		}
		self.Advance()
	}
	return op, nil
}

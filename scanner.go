package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

var ErrNotAnEventKey = fmt.Errorf("Accepting only event keys")
var ErrAmbiguous = fmt.Errorf("Ambiguous sequence")
var ErrNoMatch = fmt.Errorf("No match for sequence")

type KeyTable map[tcell.Key]Operation
type RuneTable map[rune]Operation

type Parser interface {
	Prase(ev tcell.Event) (Operation, error)
}

type GlobalParser struct{}

func (self GlobalParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	if key_event.Key() == tcell.KeyCtrlC {
		return QuitOperation{}, nil
	}
	return nil, ErrNoMatch
}

type NormalParser struct{}

func (self NormalParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
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
	return nil, ErrNoMatch
}

type InsertParser struct {
	continuous bool
	change     ReplacementInput
	input      []byte
}

func (self *InsertParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	switch key_event.Key() {
	case tcell.KeyRune:
		operation := InsertContent{
			content:              []byte(string(key_event.Rune())),
			continue_last_insert: self.continuous,
		}
		self.continuous = true
		return operation, nil
	case tcell.KeyEsc:
		self.continuous = false
		return SwitchToNormalMode{}, nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		operation := EraseCharInsertMode{continue_last_erase: self.continuous}
		self.continuous = true
		return operation, nil
	case tcell.KeyEnter:
		self.continuous = false
		return InsertNewLine{}, nil
	default:
		return nil, ErrNoMatch

	}
}

type TreeParser struct{}

func (self TreeParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	keys := KeyTable{
		tcell.KeyEsc: SwitchToNormalMode{},
	}
	runes := RuneTable{
		't': SwitchToNormalMode{},
		'k': NodeUpOperation{},
		'j': NodeDownOperation{},
		'l': NodeRightOperation{},
		'h': NodeLeftOperation{},
		'd': DeleteSelectionOperation{},
	}
	return parseKeysAndRunes(keys, runes, key_event)
}

type VisualParser struct{}

func (self VisualParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	keys := KeyTable{
		tcell.KeyEsc: SwitchToNormalMode{},
	}
	runes := RuneTable{
		'i': SwitchToInsertMode{},
		'a': SwitchToInsertModeAsAppend{},
		'v': SwitchToNormalMode{},
		'd': DeleteSelectionOperation{},

		// Navigation
		'j': NormalCursorDown{},
		'k': NormalCursorUp{},
		'h': NormalCursorLeft{},
		'l': NormalCursorRight{},
	}
	return parseKeysAndRunes(keys, runes, key_event)
}

func parseKeysAndRunes(keys KeyTable, runes RuneTable, ev *tcell.EventKey) (Operation, error) {
	if op, ok := keys[ev.Key()]; ok {
		return op, nil
	}

	if op, ok := runes[ev.Rune()]; ok {
		return op, nil
	}

	return nil, ErrNoMatch
}

package main

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var ErrNotAnEventKey = fmt.Errorf("Accepting only event keys")
var ErrAmbiguous = fmt.Errorf("Ambiguous sequence")
var ErrNoMatch = fmt.Errorf("No match for sequence")
var ErrNoKey = fmt.Errorf("No more keys to scan")

type KeyTable map[tcell.Key]Operation
type RuneTable map[rune]Operation

type Scanner interface {
	Scan(ev tcell.Event) (Operation, error)
	Advance() (*tcell.EventKey, error)
	Curr() (*tcell.EventKey, error)
	Clear()
}

func IsDigit(key_event *tcell.EventKey) bool {
	return key_event.Key() == tcell.KeyRune && unicode.IsDigit(key_event.Rune())
}

func ScanInteger(scanner Scanner) (int, error) {
	count := 0
	ev, err := scanner.Curr()
	if err != nil {
		return count, err
	}
	if !IsDigit(ev) {
		return count, ErrNoMatch
	}
	for ; err == nil && IsDigit(ev); ev, err = scanner.Advance() {
		count = count*10 + int(ev.Rune()) - int('0')
	}
	return count, nil
}

type GlobalScanner struct{}

func (self GlobalScanner) Scan(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	if key_event.Key() == tcell.KeyCtrlC {
		return QuitOperation{}, nil
	}
	return nil, ErrNoMatch
}

func (self GlobalScanner) Advance() (*tcell.EventKey, error) {
	return nil, nil
}
func (self InsertScanner) Advance() (*tcell.EventKey, error) {
	return nil, nil
}
func (self VisualScanner) Advance() (*tcell.EventKey, error) {
	return nil, nil
}
func (self TreeScanner) Advance() (*tcell.EventKey, error) {
	return nil, nil
}
func (self GlobalScanner) Curr() (*tcell.EventKey, error) {
	return nil, nil
}
func (self InsertScanner) Curr() (*tcell.EventKey, error) {
	return nil, nil
}
func (self VisualScanner) Curr() (*tcell.EventKey, error) {
	return nil, nil
}
func (self TreeScanner) Curr() (*tcell.EventKey, error) {
	return nil, nil
}
func (self GlobalScanner) Clear() {
}
func (self InsertScanner) Clear() {
}
func (self VisualScanner) Clear() {
}
func (self TreeScanner) Clear() {
}

type InsertScanner struct {
	continuous bool
	change     ReplacementInput
	input      []byte
}

// TODO Make erasing after insert continuous (single modification, single undo)
func (self *InsertScanner) Scan(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	switch key_event.Key() {
	case tcell.KeyRune:
		if !self.continuous {
			self.input = []byte{}
		}
		self.input = append(self.input, string(key_event.Rune())...)
		operation := InsertContent{
			content:              self.input,
			continue_last_insert: self.continuous,
		}
		self.continuous = true
		return operation, nil
	case tcell.KeyEsc:
		self.continuous = false
		return SwitchFromInsertToNormalMode{}, nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		operation := EraseCharInsertMode{continue_last_erase: self.continuous}
		self.continuous = false
		return operation, nil
	case tcell.KeyEnter:
		self.continuous = false
		return InsertNewLine{}, nil
	default:
		return nil, ErrNoMatch

	}
}

type TreeScanner struct {
	prev Operation
}

func (self TreeScanner) Scan(ev tcell.Event) (Operation, error) {
	ek, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
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
		op = ShiftNodeForwardOperation{}
	case ek.Rune() == 'b':
		op = ShiftNodeBackwardOperation{}
	}
	if op != nil {
		return op, nil
	}
	return nil, ErrNoMatch
}

type VisualScanner struct{}

func (self VisualScanner) Scan(ev tcell.Event) (Operation, error) {
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
		'd': EraseSelectionOperation{},

		// Navigation
		'j': NormalCursorDown{},
		'k': NormalCursorUp{},
		'h': NormalCursorLeft{},
		'l': NormalCursorRight{},
	}
	return scanKeysAndRunes(keys, runes, key_event)
}

func scanKeysAndRunes(keys KeyTable, runes RuneTable, ev *tcell.EventKey) (Operation, error) {
	if op, ok := keys[ev.Key()]; ok {
		return op, nil
	}

	if op, ok := runes[ev.Rune()]; ok {
		return op, nil
	}

	return nil, ErrNoMatch
}

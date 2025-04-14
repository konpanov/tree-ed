package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Operation interface {
	Execute(editor *Editor)
}

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor) {
	editor.current_window.cursorDown()
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor) {
	editor.current_window.cursorUp()
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor) {
	editor.current_window.cursorLeft()
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor) {
	editor.current_window.cursorRight()
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor) {
	editor.current_window.switchToInsert()
}

var ErrNotAnEventKey = fmt.Errorf("Accepting only event keys")
var ErrAmbiguous = fmt.Errorf("Ambiguous sequence")
var ErrNoMatch = fmt.Errorf("No match for sequence")

type KeyTable map[tcell.Key]Operation
type RuneTable map[rune]Operation

type Parser interface {
	Prase(ev tcell.Event) (Operation, error)
}

type GlobalParser struct {
	keys  KeyTable
	runes RuneTable
}

func (self GlobalParser) Parse(ev tcell.Event) (Operation, error) {
	if key_event, ok := ev.(*tcell.EventKey); ok {
		return parseKeysAndRunes(self.keys, self.runes, key_event)
	} else {
		return nil, ErrNotAnEventKey
	}

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

type OperationTable struct {
	key_ops  map[tcell.Key]Operation
	rune_ops map[rune]Operation
}

func (self OperationTable) Parse(ev tcell.Event) (Operation, error) {
	key, is_event_key := ev.(*tcell.EventKey)
	if !is_event_key {
		return nil, ErrNotAnEventKey
	}

	op, ok := self.key_ops[key.Key()]
	if ok {
		return op, nil
	}
	op, ok = self.rune_ops[key.Rune()]
	if ok {
		return op, nil
	}

	return nil, ErrNoMatch
}

var global_operations OperationTable = OperationTable{
	key_ops: map[tcell.Key]Operation{
		tcell.KeyCtrlC: QuitOperation{},
	},
	rune_ops: map[rune]Operation{},
}

var normal_operations OperationTable = OperationTable{
	key_ops: map[tcell.Key]Operation{},
	rune_ops: map[rune]Operation{
		'j': NormalCursorDown{},
		'k': NormalCursorUp{},
		'h': NormalCursorLeft{},
		'l': NormalCursorRight{},
		'i': SwitchToInsertMode{},
	},
}

var insert_operations OperationTable = OperationTable{
	key_ops: map[tcell.Key]Operation{},
	rune_ops: map[rune]Operation{
		'j': NormalCursorDown{},
		'k': NormalCursorUp{},
		'h': NormalCursorLeft{},
		'l': NormalCursorRight{},
		'i': SwitchToInsertMode{},
	},
}

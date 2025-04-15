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

type GlobalParser struct {
	keys  KeyTable
	runes RuneTable
}

func (self GlobalParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	return parseKeysAndRunes(self.keys, self.runes, key_event)
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

package main

import (
	"github.com/gdamore/tcell/v2"
)

type Scanner struct {
	state *ScannerState
	mode  WindowMode
}

func NewScanner() *Scanner {
	state := &ScannerState{}
	return &Scanner{
		state: state,
		mode:  NormalMode,
	}
}

func (self *Scanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *Scanner) Scan() (Operation, ScanResult) {
	op, res := OperationGroupGlobal{}.Match(self.state)
	switch self.mode {
	case NormalMode:
		if res == ScanNone {
			op, res = OperationGroupCursorMovement{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupNormal{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self.state)
		}
	case InsertMode:
		if res == ScanNone {
			op, res = OperationGroupInsert{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupTextInsert{}.Match(self.state)
		}
	case VisualMode:
		if res == ScanNone {
			op, res = OperationGroupCursorMovement{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupVisual{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self.state)
		}
	case TreeMode:
		if res == ScanNone {
			op, res = OperationGroupTree{}.Match(self.state)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self.state)
		}
	}

	switch res {
	case ScanFull:
		self.state.Clear()
	case ScanNone:
		if !self.state.IsEnd() {
			self.state.Advance()
		}
		self.state.Clear()
	case ScanStop:
		self.state.Reset()
	}
	return op, res
}

type OperationGroup interface {
	Match(state *ScannerState) (Operation, ScanResult)
}

type OperationGroupGlobal struct {
}

func (self OperationGroupGlobal) Match(state *ScannerState) (Operation, ScanResult) {
	switch {
	case state.IsEnd():
		return nil, ScanStop
	case state.ScanKey(tcell.KeyCtrlC) == ScanFull:
		return OpQuit{}, ScanFull
	default:
		return nil, ScanNone
	}
}

type OperationGroupCount struct {
}

func (self OperationGroupCount) Match(state *ScannerState) (Operation, ScanResult) {
	switch {
	case state.IsEnd():
		return nil, ScanStop
	case state.ScanDigit() == ScanFull:
		res := state.ScanZeroOrMore(state.ScanDigit)
		integer := EventKeysToInteger(state.Scanned())
		return OpCount{count: integer, op: nil}, res
	default:
		return nil, ScanNone
	}
}

type OperationGroupCursorMovement struct {
}

func (self OperationGroupCursorMovement) Match(state *ScannerState) (Operation, ScanResult) {
	runeOperations := map[rune]Operation{
		'j': OpCursorDown{},
		'k': OpCursorUp{},
		'h': OpCursorLeft{},
		'l': OpCursorRight{},
		'w': OpWordStartForward{},
		'b': OpWordStartBackward{},
		'e': OpWordEndForward{},
		'E': OpWordEndBackward{},
		'g': OpMoveToLineNumber{},
		'G': OpMoveToLastLine{},
		'$': OpLineEnd{},
		'0': OpLineStart{},
		'_': OpLineTextStart{},
		'z': OpCenterFrame{},
	}
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyCtrlD: OpMoveHalfFrameDown{},
		tcell.KeyCtrlU: OpMoveHalfFrameUp{},
		tcell.KeyCtrlE: OpMoveFrameByLineDown{},
		tcell.KeyCtrlY: OpMoveFrameByLineUp{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

type OperationGroupNormal struct {
}

func (self OperationGroupNormal) Match(state *ScannerState) (Operation, ScanResult) {
	runeOperations := map[rune]Operation{
		'd': OpEraseCursorLine{},
		'y': OpCopyCursorLine{},
		'x': OpEraseRune{},
		'a': OpInsertModeAfterCursor{},
		'A': OpInsertModeAfterLine{},
		'i': OpInsertModeBeforeCursor{},
		'I': OpInsertModeBeforeLine{},
		'v': OpVisualMode{},
		't': OpTreeMode{},
		'p': OpPasteClipboard{},
		'u': OpUndoChange{},
		's': OpEraseSelectionAndInsert{},
		'o': OpStartNewLine{},
		'O': OpStartNewLineAbove{},
	}
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyCtrlR: OpRedoChange{},
		tcell.KeyCtrlS: OpSaveFile{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

type OperationGroupInsert struct {
}

// TODO: Make erasing after insert continuous (single modification, single undo)
func (self OperationGroupInsert) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:        OpNormalMode{},
		tcell.KeyBackspace2: OpEraseRunePrev{},
		tcell.KeyBackspace:  OpEraseRunePrev{},
		tcell.KeyCtrlW:      OpEraseWordBack{},
		tcell.KeyDelete:     OpEraseRuneNext{},
	}
	return MatchKeyMap(state, keyOperations)
}

type OperationGroupTextInsert struct{}

func (self OperationGroupTextInsert) Match(state *ScannerState) (Operation, ScanResult) {
	if state.ScanTextInput() == ScanFull {
		state.ScanZeroOrMore(state.ScanTextInput)
		return OpInsertInput{content: state.Scanned()}, ScanFull
	}
	return nil, ScanNone
}

type OperationGroupVisual struct {
}

func (self OperationGroupVisual) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc: OpNormalMode{},
	}
	runeOperations := map[rune]Operation{
		'i': OpInsertModeBeforeCursor{},
		'a': OpInsertModeAfterCursor{},
		'v': OpNormalMode{},
		'd': OpEraseSelection{},
		't': OpTreeModeFormVisual{},
		'y': OpSaveClipbaord{},
		's': OpEraseSelectionAndInsert{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

type OperationGroupTree struct {
}

func (self OperationGroupTree) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:   OpNormalMode{},
		tcell.KeyCtrlR: OpRedoChange{},
		tcell.KeyCtrlK: OpDepthAnchorUp{},
	}
	runeOperations := map[rune]Operation{
		't': OpNormalMode{},
		'T': OpNormalModeAsAnchor{},
		'v': OpVisualMode{},
		'V': OpVisualModeAsAnchor{},
		'k': OpNodeUp{},
		'j': OpNodeDown{},
		'H': OpNodePrevSibling{},
		'L': OpNodeNextSibling{},
		'h': OpNodePrevSiblingOrCousin{},
		'l': OpNodeNextSiblingOrCousin{},
		'd': OpEraseSelection{},
		'f': OpSwapNodeNext{},
		'b': OpSwapNodePrev{},
		'$': OpNodeLastSibling{},
		'_': OpNodeFirstSibling{},
		'u': OpUndoChange{},
		's': OpEraseSelectionAndInsert{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

func MatchRuneMap(state *ScannerState, m map[rune]Operation) (Operation, ScanResult) {
	for r, operation := range m {
		res := state.ScanRune(r)
		if res == ScanFull {
			return operation, res
		}
	}
	return nil, ScanNone
}

func MatchKeyMap(state *ScannerState, m map[tcell.Key]Operation) (Operation, ScanResult) {
	for key, operation := range m {
		res := state.ScanKey(key)
		if res == ScanFull {
			return operation, res
		}
	}
	return nil, ScanNone
}

func MatchRuneOrKeysMap(
	state *ScannerState,
	rune_map map[rune]Operation,
	key_map map[tcell.Key]Operation,
) (Operation, ScanResult) {
	op, result := MatchKeyMap(state, key_map)
	if result == ScanNone {
		op, result = MatchRuneMap(state, rune_map)
	}
	return op, result
}

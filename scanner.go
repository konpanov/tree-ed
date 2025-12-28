package main

import (
	"github.com/gdamore/tcell/v2"
)

type Scanner interface {
	Scan() (Operation, error)
	Push(ev tcell.Event) error
}

type EditorScaner struct {
	state *ScannerState
	mode  WindowMode
}

func NewEditorScanner() *EditorScaner {
	state := &ScannerState{}
	return &EditorScaner{
		state: state,
		mode:  NormalMode,
	}
}

func (self *EditorScaner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *EditorScaner) Scan() (Operation, ScanResult) {
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
		return QuitOperation{}, ScanFull
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
		return CountOperation{count: integer, op: nil}, res
	default:
		return nil, ScanNone
	}
}

type OperationGroupCursorMovement struct {
}

func (self OperationGroupCursorMovement) Match(state *ScannerState) (Operation, ScanResult) {
	ops := map[rune]Operation{
		'j': NormalCursorDown{},
		'k': NormalCursorUp{},
		'h': NormalCursorLeft{},
		'l': NormalCursorRight{},
		'w': WordStartForwardOperation{},
		'b': WordBackwardOperation{},
		'e': WordEndForwardOperation{},
		'E': WordEndBackwardOperation{},
		'g': GoOperation{},
		'G': GoEndOperation{},
		'$': LineEndOperation{},
		'0': LineStartOperation{},
		'_': LineTextStartOperation{},
	}
	return MatchRuneMap(state, ops)
}

type OperationGroupNormal struct {
}

func (self OperationGroupNormal) Match(state *ScannerState) (Operation, ScanResult) {
	runeOperations := map[rune]Operation{
		'd': EraseLineAtCursor{},
		'y': CopyLineAtCursor{},
		'x': EraseCharNormalMode{},
		'a': SwitchToInsertModeAsAppend{},
		'A': AppendAtLineEnd{},
		'i': SwitchToInsertMode{},
		'I': InsertAtLineStart{},
		'v': SwitchToVisualmode{},
		't': SwitchToTreeMode{},
		'p': PasteClipboardOperation{},
		'u': UndoChangeOperation{},
		's': DeleteSelectionAndInsert{},
		'z': OperationCenterFrame{},
		'o': OperationStartNewLine{},
	}
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyCtrlR: RedoChangeOperation{},
		tcell.KeyCtrlD: OperationHalfFrameDown{},
		tcell.KeyCtrlU: OperationHalfFrameUp{},
		tcell.KeyCtrlE: OperationFrameLineDown{},
		tcell.KeyCtrlY: OperationFrameLineUp{},
		tcell.KeyCtrlS: OperationSaveFile{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

type OperationGroupInsert struct {
}

// TODO: Make erasing after insert continuous (single modification, single undo)
func (self OperationGroupInsert) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:        SwitchFromInsertToNormalMode{},
		tcell.KeyBackspace2: EraseCharInsertMode{},
		tcell.KeyBackspace:  EraseCharInsertMode{},
		tcell.KeyCtrlW:      DeleteToPreviousWordStart{},
		tcell.KeyDelete:     DeleteCharForward{},
	}
	return MatchKeyMap(state, keyOperations)
}

type OperationGroupTextInsert struct{}

func (self OperationGroupTextInsert) Match(state *ScannerState) (Operation, ScanResult) {
	if state.ScanTextInput() == ScanFull {
		state.ScanZeroOrMore(state.ScanTextInput)
		return InsertContentOperation{content: state.Scanned()}, ScanFull
	}
	return nil, ScanNone
}

type OperationGroupVisual struct {
}

func (self OperationGroupVisual) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc: SwitchToNormalMode{},
	}
	runeOperations := map[rune]Operation{
		'i': SwitchToInsertMode{},
		'a': SwitchToInsertModeAsAppend{},
		'v': SwitchToNormalMode{},
		'd': EraseSelectionOperation{},
		't': SwitchFromVisualToTreeMode{},
		'y': CopyToClipboardOperation{},
		's': DeleteSelectionAndInsert{},
	}
	return MatchRuneOrKeysMap(state, runeOperations, keyOperations)
}

type OperationGroupTree struct {
}

func (self OperationGroupTree) Match(state *ScannerState) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:   SwitchToNormalMode{},
		tcell.KeyCtrlR: RedoChangeOperation{},
		tcell.KeyCtrlK: OperationMoveDepthAnchorUp{},
	}
	runeOperations := map[rune]Operation{
		't': SwitchToNormalMode{},
		'T': OperationSwitchToNormalModeAsSecondCursor{},
		'v': SwitchToVisualmode{},
		'V': SwitchToVisualmodeAsSecondCursor{},
		'k': NodeUpOperation{},
		'j': NodeDownOperation{},
		'H': NodePrevSiblingOperation{},
		'L': NodeNextSiblingOperation{},
		'h': NodePrevSiblingOrCousinOperation{},
		'l': NodeNextSiblingOrCousinOperation{},
		'd': EraseSelectionOperation{},
		'f': SwapNodeForwardEndOperation{},
		'b': SwapNodeBackwardEndOperation{},
		'$': NodeLastSiblingOperation{},
		'_': NodeFirstSiblingOperation{},
		'u': UndoChangeOperation{},
		's': DeleteSelectionAndInsert{},
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

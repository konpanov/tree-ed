package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
)

type Scanner struct {
	mode  WindowMode
	keys  []*tcell.EventKey
	curr  int
	start int
}

func (self *Scanner) Push(ev tcell.Event) {
	switch value := ev.(type) {
	case *tcell.EventKey:
		self.keys = append(self.keys, value)
	default:
		log.Printf("Scanner: ignoring non key event %+v\n", ev)
	}
}

func (self *Scanner) Scan() (Operation, ScanResult) {
	op, res := OperationGroupGlobal{}.Match(self)
	switch self.mode {
	case NormalMode:
		if res == ScanNone {
			op, res = OperationGroupCursorMovement{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupNormal{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self)
		}
	case InsertMode:
		if res == ScanNone {
			op, res = OperationGroupInsert{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupTextInsert{}.Match(self)
		}
	case VisualMode:
		if res == ScanNone {
			op, res = OperationGroupCursorMovement{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupVisual{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self)
		}
	case TreeMode:
		if res == ScanNone {
			op, res = OperationGroupTree{}.Match(self)
		}
		if res == ScanNone {
			op, res = OperationGroupCount{}.Match(self)
		}
	}
	return op, res
}

func (self *Scanner) Update(res ScanResult) {
	switch res {
	case ScanFull:
		self.Clear()
	case ScanNone:
		if !self.IsEnd() {
			self.Advance()
		}
		self.Clear()
	case ScanStop:
		self.Reset()
	}
}

type OperationGroup interface {
	Match(scanner *Scanner) (Operation, ScanResult)
}

type OperationGroupGlobal struct {
}

func (self OperationGroupGlobal) Match(scanner *Scanner) (Operation, ScanResult) {
	switch {
	case scanner.IsEnd():
		return nil, ScanStop
	case scanner.ScanKey(tcell.KeyCtrlC) == ScanFull:
		return OpQuit{}, ScanFull
	default:
		return nil, ScanNone
	}
}

type OperationGroupCount struct {
}

func (self OperationGroupCount) Match(scanner *Scanner) (Operation, ScanResult) {
	switch {
	case scanner.IsEnd():
		return nil, ScanStop
	case scanner.ScanDigit() != ScanFull:
		return nil, ScanNone
	}
	res := scanner.ScanZeroOrMore(scanner.ScanDigit)
	integer := EventKeysToInteger(scanner.Scanned())
	if res == ScanFull {
		scanner.start = scanner.curr
		op, sub_res := scanner.Scan()
		if sub_res == ScanFull {
			return OpCount{count: integer, op: op}, sub_res
		}
		return nil, ScanNone
	}
	return nil, res
}

type OperationGroupCursorMovement struct {
}

func (self OperationGroupCursorMovement) Match(scanner *Scanner) (Operation, ScanResult) {
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
	return MatchRuneOrKeysMap(scanner, runeOperations, keyOperations)
}

type OperationGroupNormal struct {
}

func (self OperationGroupNormal) Match(scanner *Scanner) (Operation, ScanResult) {
	runeOperations := map[rune]Operation{
		'd': OpEraseCursorLine{},
		'y': OpCopyCursorLine{},
		'x': OpEraseRune{},
		'a': OpInsertAfterCursor{},
		'A': OpInsertAfterLine{},
		'i': OpInsertBeforeCursor{},
		'I': OpInsertBeforeLine{},
		'v': OpVisual{},
		't': OpTree{},
		'p': OpPasteClipboard{},
		'u': OpUndoChange{},
		's': OpReplaceSelection{},
		'o': OpStartNewLineBelow{},
		'O': OpStartNewLineAbove{},
	}
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyCtrlR: OpRedoChange{},
		tcell.KeyCtrlS: OpSaveFile{},
	}
	return MatchRuneOrKeysMap(scanner, runeOperations, keyOperations)
}

type OperationGroupInsert struct {
}

// TODO: Make erasing after insert continuous (single modification, single undo)
func (self OperationGroupInsert) Match(scanner *Scanner) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:        OpNormal{},
		tcell.KeyBackspace2: OpEraseRunePrev{},
		tcell.KeyBackspace:  OpEraseRunePrev{},
		tcell.KeyCtrlW:      OpEraseWordBack{},
		tcell.KeyDelete:     OpEraseRuneNext{},
	}
	return MatchKeyMap(scanner, keyOperations)
}

type OperationGroupTextInsert struct{}

func (self OperationGroupTextInsert) Match(scanner *Scanner) (Operation, ScanResult) {
	if scanner.ScanTextInput() == ScanFull {
		scanner.ScanZeroOrMore(scanner.ScanTextInput)
		return OpInsertInput{content: scanner.Scanned()}, ScanFull
	}
	return nil, ScanNone
}

type OperationGroupVisual struct {
}

func (self OperationGroupVisual) Match(scanner *Scanner) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc: OpNormal{},
	}
	runeOperations := map[rune]Operation{
		'i': OpInsertBeforeCursor{},
		'a': OpInsertAfterCursor{},
		'v': OpNormal{},
		'd': OpEraseSelection{},
		't': OpTree{},
		'y': OpSaveClipbaord{},
		's': OpReplaceSelection{},
	}
	return MatchRuneOrKeysMap(scanner, runeOperations, keyOperations)
}

type OperationGroupTree struct {
}

func (self OperationGroupTree) Match(scanner *Scanner) (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:   OpNormal{},
		tcell.KeyCtrlR: OpRedoChange{},
		tcell.KeyCtrlK: OpDepthUp{},
		tcell.KeyCtrlJ: OpDepthDown{},
	}
	runeOperations := map[rune]Operation{
		't': OpNormal{},
		'v': OpVisual{},
		'T': OpNormalAsAnchor{},
		'V': OpVisualAsAnchor{},
		'k': OpNodeUp{},
		'j': OpNodeDown{},
		'H': OpNodePrevSibling{},
		'L': OpNodeNextSibling{},
		'h': OpNodePrevDepth{},
		'l': OpNodeNextDepth{},
		'd': OpEraseSelection{},
		'f': OpSwapNodeNext{},
		'b': OpSwapNodePrev{},
		'$': OpNodeLastSibling{},
		'_': OpNodeFirstSibling{},
		'u': OpUndoChange{},
		's': OpReplaceSelection{},
		'y': OpSaveClipbaord{},
	}
	return MatchRuneOrKeysMap(scanner, runeOperations, keyOperations)
}

func MatchRuneMap(scanner *Scanner, m map[rune]Operation) (Operation, ScanResult) {
	for r, operation := range m {
		res := scanner.ScanRune(r)
		if res == ScanFull {
			return operation, res
		}
	}
	return nil, ScanNone
}

func MatchKeyMap(scanner *Scanner, m map[tcell.Key]Operation) (Operation, ScanResult) {
	for key, operation := range m {
		res := scanner.ScanKey(key)
		if res == ScanFull {
			return operation, res
		}
	}
	return nil, ScanNone
}

func MatchRuneOrKeysMap(
	scanner *Scanner,
	rune_map map[rune]Operation,
	key_map map[tcell.Key]Operation,
) (Operation, ScanResult) {
	op, result := MatchKeyMap(scanner, key_map)
	if result == ScanNone {
		op, result = MatchRuneMap(scanner, rune_map)
	}
	return op, result
}

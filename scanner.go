package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
)

type ScanResult int

const (
	ScanFull ScanResult = iota
	ScanNone
	ScanStop
)

type ScanFunc func() ScanResult
type ScanOpFunc func() (Operation, ScanResult)

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

func (self *Scanner) Peek() *tcell.EventKey {
	return self.keys[self.curr]
}

func (self *Scanner) Advance() *tcell.EventKey {
	curr := self.Peek()
	self.curr++
	return curr
}

func (self *Scanner) IsEnd() bool {
	return self.curr >= len(self.keys)
}

func (self *Scanner) Input() []*tcell.EventKey {
	return self.keys[self.start:]
}

func (self *Scanner) Scanned() []*tcell.EventKey {
	return self.keys[self.start:self.curr]
}

func (self *Scanner) Clear() {
	self.keys = self.keys[self.curr:]
	self.start = 0
	self.curr = 0
}

func (self *Scanner) Reset() {
	self.curr = 0
	self.start = 0
}

func (self *Scanner) Scan() (Operation, ScanResult) {
	if self.IsEnd() {
		return nil, ScanStop
	}
	switch self.mode {
	case NormalMode:
		return self.ScanOperationGroup([]ScanOpFunc{
			self.ScanGlobalOperations,
			self.ScanCursorOperation,
			self.ScanNormalOperation,
			self.ScanCountOperation,
		})
	case InsertMode:
		return self.ScanOperationGroup([]ScanOpFunc{
			self.ScanGlobalOperations,
			self.ScanInsertOperation,
			self.ScanTextInsertOperation,
		})
	case VisualMode:
		return self.ScanOperationGroup([]ScanOpFunc{
			self.ScanGlobalOperations,
			self.ScanCursorOperation,
			self.ScanVisualOperation,
			self.ScanCountOperation,
		})
	case TreeMode:
		return self.ScanOperationGroup([]ScanOpFunc{
			self.ScanGlobalOperations,
			self.ScanTreeOperation,
			self.ScanCountOperation,
		})
	default:
		return self.ScanGlobalOperations()
	}
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

func (self *Scanner) ScanOperationGroup(group []ScanOpFunc) (Operation, ScanResult) {
	for _, scan := range group {
		op, res := scan()
		if res != ScanNone {
			return op, res
		}
	}
	return nil, ScanNone
}

func (self *Scanner) ScanGlobalOperations() (Operation, ScanResult) {
	switch {
	case self.ScanKey(tcell.KeyCtrlC) == ScanFull:
		return OpQuit{}, ScanFull
	default:
		return nil, ScanNone
	}
}

func (self *Scanner) ScanCountOperation() (Operation, ScanResult) {
	if self.ScanRune('0') == ScanFull {
		return nil, ScanNone
	}
	if res := self.ScanOneOrMore(self.ScanDigit); res != ScanFull {
		return nil, res
	}
	integer := EventKeysToInteger(self.Scanned())
	self.start = self.curr
	op, res := self.Scan()
	if res == ScanFull {
		return OpCount{count: integer, op: op}, ScanFull
	}
	return nil, res
}

func (self *Scanner) ScanCursorOperation() (Operation, ScanResult) {
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
	return MatchRuneOrKeysMap(self, runeOperations, keyOperations)
}

func (self *Scanner) ScanNormalOperation() (Operation, ScanResult) {
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
	return MatchRuneOrKeysMap(self, runeOperations, keyOperations)
}

// TODO: Make erasing after insert continuous (single modification, single undo)
func (self *Scanner) ScanInsertOperation() (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:        OpNormal{},
		tcell.KeyBackspace2: OpEraseRunePrev{},
		tcell.KeyBackspace:  OpEraseRunePrev{},
		tcell.KeyCtrlW:      OpEraseWordBack{},
		tcell.KeyDelete:     OpEraseRuneNext{},
	}
	return MatchKeyMap(self, keyOperations)
}

func (self *Scanner) ScanTextInsertOperation() (Operation, ScanResult) {
	if self.ScanTextInput() == ScanFull {
		self.ScanZeroOrMore(self.ScanTextInput)
		return OpInsertInput{content: self.Scanned()}, ScanFull
	}
	return nil, ScanNone
}

func (self *Scanner) ScanVisualOperation() (Operation, ScanResult) {
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
	return MatchRuneOrKeysMap(self, runeOperations, keyOperations)
}

func (self *Scanner) ScanTreeOperation() (Operation, ScanResult) {
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
	return MatchRuneOrKeysMap(self, runeOperations, keyOperations)
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

func (self *Scanner) ScanWithCondition(cond func() bool) ScanResult {
	if self.IsEnd() {
		return ScanStop
	} else if cond() {
		self.Advance()
		return ScanFull
	} else {
		return ScanNone
	}
}

func (self *Scanner) ScanZeroOrMore(scan ScanFunc) ScanResult {
	for {
		switch scan() {
		case ScanStop:
			return ScanStop
		case ScanNone:
			return ScanFull
		}
	}
}

func (self *Scanner) ScanOneOrMore(scan ScanFunc) ScanResult {
	if scan() != ScanFull {
		return ScanNone
	}
	for {
		switch scan() {
		case ScanStop:
			return ScanStop
		case ScanNone:
			return ScanFull
		}
	}
}

func (self *Scanner) ScanKey(key tcell.Key) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Key() == key
	})
}

func (self *Scanner) ScanRune(r rune) ScanResult {
	return self.ScanWithCondition(func() bool {
		return self.Peek().Rune() == r
	})
}

func (self *Scanner) ScanDigit() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsDigitKey(self.Peek())
	})
}

func (self *Scanner) ScanTextInput() ScanResult {
	return self.ScanWithCondition(func() bool {
		return IsTextInputKey(self.Peek())
	})
}

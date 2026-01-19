package main

import (
	"github.com/gdamore/tcell/v2"
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

func (self *Scanner) UpdateMode(mode WindowMode) {
	self.mode = mode
}

func (self *Scanner) Push(ev tcell.Event) {
	switch value := ev.(type) {
	case *tcell.EventKey:
		self.keys = append(self.keys, value)
	default:
		debug_logf("Scanner: ignoring non key event %+v\n", ev)
	}
}

func (self *Scanner) Scan() (Operation, ScanResult) {
	if self.isEnd() {
		return nil, ScanStop
	}
	switch self.mode {
	case NormalMode:
		return self.scanOperationGroup([]ScanOpFunc{
			self.scanGlobalOperations,
			self.scanCursorOperation,
			self.scanNormalOperation,
			self.scanCountOperation,
		})
	case InsertMode:
		return self.scanOperationGroup([]ScanOpFunc{
			self.scanGlobalOperations,
			self.scanInsertOperation,
			self.scanTextInsertOperation,
		})
	case VisualMode:
		return self.scanOperationGroup([]ScanOpFunc{
			self.scanGlobalOperations,
			self.scanCursorOperation,
			self.scanVisualOperation,
			self.scanCountOperation,
		})
	case TreeMode:
		return self.scanOperationGroup([]ScanOpFunc{
			self.scanGlobalOperations,
			self.scanTreeOperation,
			self.scanCountOperation,
		})
	default:
		return self.scanGlobalOperations()
	}
}

func (self *Scanner) Update(res ScanResult) {
	switch res {
	case ScanFull:
		self.clear()
	case ScanNone:
		if !self.isEnd() {
			self.advance()
		}
		self.clear()
	case ScanStop:
		self.reset()
	}
}

func (self *Scanner) scanOperationGroup(group []ScanOpFunc) (Operation, ScanResult) {
	for _, scan := range group {
		op, res := scan()
		if res != ScanNone {
			return op, res
		}
	}
	return nil, ScanNone
}

func (self *Scanner) scanGlobalOperations() (Operation, ScanResult) {
	switch {
	case self.scanKey(tcell.KeyCtrlC) == ScanFull:
		return OpQuit{}, ScanFull
	default:
		return nil, ScanNone
	}
}

func (self *Scanner) scanCountOperation() (Operation, ScanResult) {
	if self.scanRune('0') == ScanFull {
		return nil, ScanNone
	}
	if res := self.scanOneOrMore(self.scanDigit); res != ScanFull {
		return nil, res
	}
	integer := EventKeysToInteger(self.scanned())
	self.start = self.curr
	op, res := self.Scan()
	if res == ScanFull {
		return OpCount{count: integer, op: op}, ScanFull
	}
	return nil, res
}

func (self *Scanner) scanCursorOperation() (Operation, ScanResult) {
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

func (self *Scanner) scanNormalOperation() (Operation, ScanResult) {
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
func (self *Scanner) scanInsertOperation() (Operation, ScanResult) {
	keyOperations := map[tcell.Key]Operation{
		tcell.KeyEsc:        OpNormal{},
		tcell.KeyBackspace2: OpEraseRunePrev{},
		tcell.KeyBackspace:  OpEraseRunePrev{},
		tcell.KeyCtrlW:      OpEraseWordBack{},
		tcell.KeyDelete:     OpEraseRuneNext{},
	}
	return MatchKeyMap(self, keyOperations)
}

func (self *Scanner) scanTextInsertOperation() (Operation, ScanResult) {
	if res := self.scanOneOrMore(self.scanTextInput); res == ScanNone {
		return nil, res
	}
	keys := self.scanned()
	lines := [][]byte{}
	content := []byte{}
	for _, ek := range keys {
		if ek.Key() == tcell.KeyRune {
			content = append(content, []byte(string(ek.Rune()))...)
		} else if ek.Key() == tcell.KeyTab {
			content = append(content, '\t')
		} else if ek.Key() == tcell.KeyCR {
			lines = append(lines, content)
			content = []byte{}
		} else if ek.Key() == tcell.KeyLF {
			content = append(content, '\n')
		} else {
			break
		}
	}
	lines = append(lines, content)
	return OpInsertInput{lines: lines}, ScanFull
}

func (self *Scanner) scanVisualOperation() (Operation, ScanResult) {
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

func (self *Scanner) scanTreeOperation() (Operation, ScanResult) {
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
		res := scanner.scanRune(r)
		if res == ScanFull {
			return operation, res
		}
	}
	return nil, ScanNone
}

func MatchKeyMap(scanner *Scanner, m map[tcell.Key]Operation) (Operation, ScanResult) {
	for key, operation := range m {
		res := scanner.scanKey(key)
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

func (self *Scanner) scanWithCondition(cond func() bool) ScanResult {
	if self.isEnd() {
		return ScanStop
	} else if cond() {
		self.advance()
		return ScanFull
	} else {
		return ScanNone
	}
}

func (self *Scanner) scanZeroOrMore(scan ScanFunc) ScanResult {
	for {
		switch scan() {
		case ScanStop:
			return ScanStop
		case ScanNone:
			return ScanFull
		}
	}
}

func (self *Scanner) scanOneOrMore(scan ScanFunc) ScanResult {
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

func (self *Scanner) scanKey(key tcell.Key) ScanResult {
	return self.scanWithCondition(func() bool {
		return self.peek().Key() == key
	})
}

func (self *Scanner) scanRune(r rune) ScanResult {
	return self.scanWithCondition(func() bool {
		return self.peek().Rune() == r
	})
}

func (self *Scanner) scanDigit() ScanResult {
	return self.scanWithCondition(func() bool {
		return IsDigitKey(self.peek())
	})
}

func (self *Scanner) scanTextInput() ScanResult {
	return self.scanWithCondition(func() bool {
		return IsTextInputKey(self.peek())
	})
}

func (self *Scanner) peek() *tcell.EventKey {
	return self.keys[self.curr]
}

func (self *Scanner) advance() *tcell.EventKey {
	curr := self.peek()
	self.curr++
	return curr
}

func (self *Scanner) isEnd() bool {
	return self.curr >= len(self.keys)
}

func (self *Scanner) Input() []*tcell.EventKey {
	return self.keys
}

func (self *Scanner) scanned() []*tcell.EventKey {
	return self.keys[self.start:self.curr]
}

func (self *Scanner) clear() {
	self.keys = self.keys[self.curr:]
	self.start = 0
	self.curr = 0
}

func (self *Scanner) reset() {
	self.curr = 0
	self.start = 0
}

package main

import (
	"os"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

type Operation interface {
	Execute(editor *Editor, count int)
}

type OpNone struct{}

func (self OpNone) Execute(editor *Editor, count int) {
}

type OpQuit struct{}

func (self OpQuit) Execute(editor *Editor, count int) {
	editor.is_quiting = true
}

type OpCursorDown struct{}

func (self OpCursorDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorDown(count)
}

type OpCursorUp struct{}

func (self OpCursorUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorUp(count)
}

type OpCursorLeft struct{}

func (self OpCursorLeft) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorLeft(count)
}

type OpCursorRight struct{}

func (self OpCursorRight) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorRight(count)
}

type OpInsertModeBeforeCursor struct{}

func (self OpInsertModeBeforeCursor) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToInsert()
}

type OpInsertModeAfterCursor struct{}

func (self OpInsertModeAfterCursor) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight(1)
}

type OpInsertModeAfterLine struct{}

func (self OpInsertModeAfterLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpLineEnd{}.Execute(editor, 1)
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight(1)
}

type OpInsertModeBeforeLine struct{}

func (self OpInsertModeBeforeLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpLineStart{}.Execute(editor, 1)
	editor.curwin.switchToInsert()
}

type OpVisualMode struct{}

func (self OpVisualMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToVisual()
}

type OpVisualModeAsAnchor struct{}

func (self OpVisualModeAsAnchor) Execute(editor *Editor, count int) {
	OpSwapCursorWithAnchor{}.Execute(editor, count)
	OpVisualMode{}.Execute(editor, count)
}

type OpNormalMode struct{}

func (self OpNormalMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.continuousInsert = false
	editor.curwin.switchToNormal()
}

type OpNormalModeAsAnchor struct{}

func (self OpNormalModeAsAnchor) Execute(editor *Editor, count int) {
	OpSwapCursorWithAnchor{}.Execute(editor, count)
	OpNormalMode{}.Execute(editor, count)
}

type OpSwapCursorWithAnchor struct{}

func (self OpSwapCursorWithAnchor) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	cursor := editor.curwin.cursor
	anchor := editor.curwin.anchor
	editor.curwin.setCursor(anchor, true)
	editor.curwin.setAnchor(cursor)
}

type OpTreeMode struct{}

func (self OpTreeMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		index := editor.curwin.cursor.Index()
		node := NodeLeaf(editor.curwin.buffer.Tree().RootNode(), index)
		editor.curwin.setNode(node, true)
		editor.curwin.switchToTree()
	}
}

type OpTreeModeFormVisual struct{}

func (self OpTreeModeFormVisual) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		start, end := editor.curwin.getSelection()
		node := MinimalNode(editor.curwin.buffer.Tree().RootNode(), start, end)
		editor.curwin.setNode(node, true)
		editor.curwin.switchToTree()
	}
}

type OpEraseCursorLine struct{}

func (self OpEraseCursorLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.eraseLineAtCursor(count)
}

type OpCopyCursorLine struct{}

func (self OpCopyCursorLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if clipboard.Unsupported {
		return
	}
	win := editor.curwin
	line := win.buffer.Lines()[win.cursor.Row()]
	text := win.buffer.Content()[line.start:line.next_start]
	clipboard.WriteAll(string(text))
}

type OpEraseRuneNormalMode struct{}

func (self OpEraseRuneNormalMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if debug {
		assert(count != 0, "Count is not expected to be 0")
	}
	win := editor.curwin
	composite := CompositeChange{}
	for range count {
		if win.cursor.IsLineBreak() {
			break
		}
		change := NewEraseRuneChange(win, win.cursor.Index())
		change.Apply(win)
		composite.changes = append(composite.changes, change)
	}
	win.history.Push(HistoryState{change: composite})
}

type OpEraseRuneInsertMode struct {
}

// TODO add composite modification?
func (self OpEraseRuneInsertMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.eraseContent(editor.curwin.continuousInsert)
	editor.curwin.continuousInsert = true
}

type OpInsertInput struct {
	content []*tcell.EventKey
}

func (self OpInsertInput) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	content := []byte{}
	for _, ek := range self.content {
		if ek.Key() == tcell.KeyRune {
			content = append(content, []byte(string(ek.Rune()))...)
		} else if ek.Key() == tcell.KeyTab {
			content = append(content, '\t')
		} else if ek.Key() == tcell.KeyCR {
			content = append(content, editor.curwin.buffer.LineBreak()...)
		} else if ek.Key() == tcell.KeyLF {
			content = append(content, '\n')
		} else {
			break
		}
	}
	editor.curwin.insertContent(editor.curwin.continuousInsert, content)
	editor.curwin.continuousInsert = true
}

type OpNodeUp struct{}

func (self OpNodeUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeUp()
		}
	}
}

type OpNodeDown struct{}

func (self OpNodeDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeDown()
		}
	}
}

type OpNodeNextSibling struct{}

func (self OpNodeNextSibling) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSibling()
		}
	}
}

type OpNodeNextSiblingOrCousin struct{}

func (self OpNodeNextSiblingOrCousin) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSiblingOrCousin()
		}
	}
}

type OpNodePrevSibling struct{}

func (self OpNodePrevSibling) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSibling()
		}
	}
}

type OpNodePrevSiblingOrCousin struct{}

func (self OpNodePrevSiblingOrCousin) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSiblingOrCousin()
		}
	}
}

type OpNodeFirstSibling struct{}

func (self OpNodeFirstSibling) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToFirstSibling()
	}
}

type OpNodeLastSibling struct{}

func (self OpNodeLastSibling) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToLastSibling()
	}
}

type OpEraseSelection struct{}

func (self OpEraseSelection) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	start, end := win.cursor.Index(), win.anchor.Index()
	start, end = min(start, end), max(start, end)
	change := NewEraseChange(win, start, end+1)
	change.Apply(win)
	win.history.Push(HistoryState{change: change})
	win.switchToNormal()
}

type OpUndoChange struct{}

func (self OpUndoChange) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	for range count {
		if mod := win.history.Back(); mod != nil {
			mod.Reverse().Apply(win)
		}
	}
}

type OpRedoChange struct{}

func (self OpRedoChange) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	for range count {
		if mod := win.history.Forward(); mod != nil {
			mod.Apply(win)
		}
	}
}

type OpWordStartForward struct{}

func (self OpWordStartForward) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartNext(), true)
	}
}

type OpWordEndForward struct{}

func (self OpWordEndForward) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndNext(), true)
	}
}

type OpWordEndBackward struct{}

func (self OpWordEndBackward) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndPrev(), true)
	}
}

type OpWordStartBackward struct{}

func (self OpWordStartBackward) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartPrev(), true)
	}
}

type OpLineEnd struct{}

func (self OpLineEnd) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToLineEnd(), true)

}

type OpLineStart struct{}

func (self OpLineStart) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToLineStart(), true)

}

type OpLineTextStart struct{}

func (self OpLineTextStart) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToLineTextStart(), true)

}

type OpCount struct {
	count int
	op    Operation
}

func (self OpCount) Execute(editor *Editor, count int) {
	if self.op != nil {
		self.op.Execute(editor, self.count)
	}
}

type OpMoveToLineNumer struct{}

func (self OpMoveToLineNumer) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	pos := editor.curwin.cursor.Pos()
	pos.col = editor.curwin.originColumn
	pos.row = max(0, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type OpMoveToLastLine struct{}

func (self OpMoveToLastLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	pos := editor.curwin.cursor.Pos()
	pos.col = editor.curwin.originColumn
	pos.row = max(0, len(editor.curwin.buffer.Lines())-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type OpSwapNodeNext struct{}

func (self OpSwapNodeNext) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		win := editor.curwin

		swapee := win.getNode()
		for range count {
			if swapee = NextSiblingOrCousinDepth(swapee, win.originDepth); swapee == nil {
				return
			}
		}

		startB, endB := int(swapee.StartByte()), int(swapee.EndByte())
		startA, endA := order(win.cursor.Index(), win.anchor.Index())
		change := NewSwapChange(win, startA, endA+1, startB, endB)
		change.Apply(win)

		win.setCursor(win.cursor.ToIndex(-endA+startA+endB-1), true)
		win.setAnchor(win.anchor.ToIndex(endB - 1))
		win.history.Push(HistoryState{change: change})
	}
}

type OpSwapNodePrev struct{}

func (self OpSwapNodePrev) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		win := editor.curwin
		swapee := win.getNode()
		for range count {
			if swapee = PrevSiblingOrCousinDepth(swapee, win.originDepth); swapee == nil {
				return
			}
		}
		startA, endA := int(swapee.StartByte()), int(swapee.EndByte())
		startB, endB := order(win.cursor.Index(), win.anchor.Index())
		change := NewSwapChange(win, startA, endA, startB, endB+1)
		change.Apply(win)

		win.setCursor(win.cursor.ToIndex(startA), true)
		win.setAnchor(win.anchor.ToIndex(startA + endB - startB))
		win.history.Push(HistoryState{change: change})
	}
}

type OpPasteClipboard struct{}

func (self OpPasteClipboard) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if clipboard.Unsupported {
		return
	}
	text, err := clipboard.ReadAll()
	panic_if_error(err)
	editor.curwin.insertContent(false, []byte(text))
}

type OpSaveClipbaord struct{}

func (self OpSaveClipbaord) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if clipboard.Unsupported {
		return
	}
	win := editor.curwin
	start, end := win.getSelection()
	text := win.buffer.Content()[start:end]
	clipboard.WriteAll(string(text))
}

type OpDepthAnchorUp struct{}

func (self OpDepthAnchorUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.originDepth = min(editor.curwin.originDepth - 1)
}

type OpEraseToPreviousWordStart struct{}

func (self OpEraseToPreviousWordStart) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	end := editor.curwin.cursor
	start := end.WordStartPrev()
	change := NewEraseChange(editor.curwin, start.Index(), end.Index())
	change.Apply(editor.curwin)
	editor.curwin.history.Push(HistoryState{change: change})
	editor.curwin.continuousInsert = false
}

// TODO: Make continuous with inserts
type OpEraseCharNext struct{}

func (self OpEraseCharNext) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	start := editor.curwin.cursor
	end := start.RuneNext()
	change := NewEraseChange(editor.curwin, start.Index(), end.Index())
	change.Apply(editor.curwin)
	editor.curwin.history.Push(HistoryState{change: change})
	editor.curwin.continuousInsert = false
}

type OpEraseSelectionAndInsert struct{}

func (self OpEraseSelectionAndInsert) Execute(editor *Editor, count int) {
	OpEraseSelection{}.Execute(editor, count)
	OpInsertModeBeforeCursor{}.Execute(editor, count)
}

type OpMoveHalfFrameDown struct{}

func (self OpMoveHalfFrameDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	rows := count * frame.Height() / 2
	OpCursorDown{}.Execute(editor, rows)
	pos := frame.TopLeft()
	content_height := len(editor.curwin.buffer.Lines())
	pos.row += max(min(rows, content_height-frame.bot), 0)
	editor.curwin.frame = frame.Shift(pos)
}

type OpMoveHalfFrameUp struct{}

func (self OpMoveHalfFrameUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	rows := count * frame.Height() / 2
	OpCursorUp{}.Execute(editor, rows)
	pos := frame.TopLeft()
	pos.row = max(pos.row-rows, 0)
	editor.curwin.frame = frame.Shift(pos)
}

type OpMoveFrameByLineUp struct{}

func (self OpMoveFrameByLineUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	OpCursorUp{}.Execute(editor, count)
	pos := Pos{
		row: min(max(frame.top-count, 0), len(editor.curwin.buffer.Lines())-frame.Height()),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OpMoveFrameByLineDown struct{}

func (self OpMoveFrameByLineDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	OpCursorDown{}.Execute(editor, count)
	pos := Pos{
		row: min(max(frame.top+count, 0), len(editor.curwin.buffer.Lines())-frame.Height()),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OpCenterFrame struct{}

func (self OpCenterFrame) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	cursor := editor.curwin.cursor
	pos := Pos{
		row: max(cursor.Row()-frame.Height()/2, 0),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OpSaveFile struct{}

func (self OpSaveFile) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer == nil {
		return
	}
	if editor.curwin.buffer.Filename() == "" {
		return
	}
	filename := editor.curwin.buffer.Filename()
	info, err := os.Stat(filename)
	panic_if_error(err)
	os.WriteFile(filename, editor.curwin.buffer.Content(), info.Mode())
}

type OpStartNewLine struct{}

func (self OpStartNewLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	OpInsertModeAfterLine{}.Execute(editor, count)
	editor.curwin.insertContent(false, editor.curwin.buffer.LineBreak())
}

type OpStartNewLineAbove struct{}

func (self OpStartNewLineAbove) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}

	OpInsertModeBeforeLine{}.Execute(editor, count)
	editor.curwin.insertContent(false, editor.curwin.buffer.LineBreak())
	OpCursorUp{}.Execute(editor, count)
}

package main

import (
	"os"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

type Operation interface {
	Execute(editor *Editor, count int)
}

type NoOperation struct{}

func (self NoOperation) Execute(editor *Editor, count int) {
}

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor, count int) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorDown(count)
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorUp(count)
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorLeft(count)
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.cursorRight(count)
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToInsert()
}

type SwitchToInsertModeAsAppend struct{}

func (self SwitchToInsertModeAsAppend) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight(1)
}

type AppendAtLineEnd struct{}

func (self AppendAtLineEnd) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	LineEndOperation{}.Execute(editor, 1)
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight(1)
}

type InsertAtLineStart struct{}

func (self InsertAtLineStart) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	LineStartOperation{}.Execute(editor, 1)
	editor.curwin.switchToInsert()
}

type SwitchToVisualmode struct{}

func (self SwitchToVisualmode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToVisual()
}

type SwitchToVisualmodeAsSecondCursor struct{}

func (self SwitchToVisualmodeAsSecondCursor) Execute(editor *Editor, count int) {
	OperationSwapCursors{}.Execute(editor, count)
	SwitchToVisualmode{}.Execute(editor, count)
}

type SwitchToNormalMode struct{}

func (self SwitchToNormalMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.switchToNormal()
}

type OperationSwitchToNormalModeAsSecondCursor struct{}

func (self OperationSwitchToNormalModeAsSecondCursor) Execute(editor *Editor, count int) {
	OperationSwapCursors{}.Execute(editor, count)
	SwitchToNormalMode{}.Execute(editor, count)
}

type OperationSwapCursors struct{}

func (self OperationSwapCursors) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	cursor := editor.curwin.cursor
	anchor := editor.curwin.anchor
	editor.curwin.setCursor(anchor, true)
	editor.curwin.setAnchor(cursor)
}

type SwitchFromInsertToNormalMode struct{}

func (self SwitchFromInsertToNormalMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.continuousInsert = false
	editor.curwin.switchToNormal()
}

type SwitchToTreeMode struct{}

func (self SwitchToTreeMode) Execute(editor *Editor, count int) {
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

type SwitchFromVisualToTreeMode struct{}

func (self SwitchFromVisualToTreeMode) Execute(editor *Editor, count int) {
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

type EraseLineAtCursor struct{}

func (self EraseLineAtCursor) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.eraseLineAtCursor(count)
}

type CopyLineAtCursor struct{}

func (self CopyLineAtCursor) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if clipboard.Unsupported {
		return
	}
	win := editor.curwin
	row := win.cursor.Row()
	line, err := win.buffer.Line(row)
	panic_if_error(err)
	text := win.buffer.Content()[line.start:line.next_start]
	clipboard.WriteAll(string(text))
}

type EraseCharNormalMode struct{}

func (self EraseCharNormalMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if debug {
		assert(count != 0, "Count is not expected to be 0")
	}
	win := editor.curwin
	composite := CompositeChange{}
	for range count {
		if win.cursor.IsNewLine() {
			break
		}
		change := NewEraseRuneChange(win, win.cursor.Index())
		change.Apply(win)
		composite.changes = append(composite.changes, change)
	}
	win.undotree.Push(UndoState{change: composite}, true)
}

type EraseCharInsertMode struct {
}

// TODO add composite modification?
func (self EraseCharInsertMode) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.eraseContent(editor.curwin.continuousInsert)
	editor.curwin.continuousInsert = true
}

type InsertContentOperation struct {
	content []*tcell.EventKey
}

func (self InsertContentOperation) Execute(editor *Editor, count int) {
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
			content = append(content, editor.curwin.buffer.Nl_seq()...)
		} else if ek.Key() == tcell.KeyLF {
			content = append(content, '\n')
		} else {
			break
		}
	}
	editor.curwin.insertContent(editor.curwin.continuousInsert, content)
	editor.curwin.continuousInsert = true
}

type NodeUpOperation struct{}

func (self NodeUpOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeUp()
		}
	}
}

type NodeDownOperation struct{}

func (self NodeDownOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeDown()
		}
	}
}

type NodeNextSiblingOperation struct{}

func (self NodeNextSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSibling()
		}
	}
}

type NodeNextSiblingOrCousinOperation struct{}

func (self NodeNextSiblingOrCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSiblingOrCousin()
		}
	}
}

type NodePrevSiblingOperation struct{}

func (self NodePrevSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSibling()
		}
	}
}

type NodePrevSiblingOrCousinOperation struct{}

func (self NodePrevSiblingOrCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSiblingOrCousin()
		}
	}
}

type NodeFirstSiblingOperation struct{}

func (self NodeFirstSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToFirstSibling()
	}
}

type NodeLastSiblingOperation struct{}

func (self NodeLastSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToLastSibling()
	}
}

type EraseSelectionOperation struct{}

func (self EraseSelectionOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	start, end := win.cursor.Index(), win.anchor.Index()
	start, end = min(start, end), max(start, end)
	change := NewEraseChange(win, start, end+1)
	change.Apply(win)
	win.undotree.Push(UndoState{change: change}, true)
	win.switchToNormal()
}

type UndoChangeOperation struct{}

func (self UndoChangeOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	for range count {
		if mod := win.undotree.Back(); mod != nil {
			mod.Reverse().Apply(win)
		}
	}
}

type RedoChangeOperation struct{}

func (self RedoChangeOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	win := editor.curwin
	for range count {
		if mod := win.undotree.Forward(); mod != nil {
			mod.Apply(win)
		}
	}
}

type WordStartForwardOperation struct{}

func (self WordStartForwardOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartNext(), true)
	}
}

type WordEndForwardOperation struct{}

func (self WordEndForwardOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndNext(), true)
	}
}

type WordEndBackwardOperation struct{}

func (self WordEndBackwardOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndPrev(), true)
	}
}

type WordBackwardOperation struct{}

func (self WordBackwardOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartPrev(), true)
	}
}

type LineEndOperation struct{}

func (self LineEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowEnd(), true)

}

type LineStartOperation struct{}

func (self LineStartOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowStart(), true)

}

type LineTextStartOperation struct{}

func (self LineTextStartOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowTextStart(), true)

}

type CountOperation struct {
	count int
	op    Operation
}

func (self CountOperation) Execute(editor *Editor, count int) {
	if self.op != nil {
		self.op.Execute(editor, self.count)
	}
}

type GoOperation struct{}

func (self GoOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.originColumn
	pos.row = max(0, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type GoEndOperation struct{}

func (self GoEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.originColumn
	pos.row = max(0, len(editor.curwin.buffer.Lines())-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type SwapNodeForwardEndOperation struct{}

func (self SwapNodeForwardEndOperation) Execute(editor *Editor, count int) {
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
		win.undotree.Push(UndoState{change: change}, true)
	}
}

type SwapNodeBackwardEndOperation struct{}

func (self SwapNodeBackwardEndOperation) Execute(editor *Editor, count int) {
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
		win.undotree.Push(UndoState{change: change}, true)
	}
}

type PasteClipboardOperation struct{}

func (self PasteClipboardOperation) Execute(editor *Editor, count int) {
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

type CopyToClipboardOperation struct{}

func (self CopyToClipboardOperation) Execute(editor *Editor, count int) {
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

type OperationMoveDepthAnchorUp struct{}

func (self OperationMoveDepthAnchorUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	editor.curwin.originDepth = min(editor.curwin.originDepth - 1)
}

type DeleteToPreviousWordStart struct{}

func (self DeleteToPreviousWordStart) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	end := editor.curwin.cursor
	start := end.WordStartPrev()
	change := NewEraseChange(editor.curwin, start.Index(), end.Index())
	change.Apply(editor.curwin)
	editor.curwin.undotree.Push(UndoState{change: change}, true)
	editor.curwin.continuousInsert = false
}

// TODO: Make continuous with inserts
type DeleteCharForward struct{}

func (self DeleteCharForward) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	start := editor.curwin.cursor
	end := start.RuneNext()
	change := NewEraseChange(editor.curwin, start.Index(), end.Index())
	change.Apply(editor.curwin)
	editor.curwin.undotree.Push(UndoState{change: change}, true)
	editor.curwin.continuousInsert = false
}

type DeleteSelectionAndInsert struct{}

func (self DeleteSelectionAndInsert) Execute(editor *Editor, count int) {
	EraseSelectionOperation{}.Execute(editor, count)
	SwitchToInsertMode{}.Execute(editor, count)
}

type OperationHalfFrameDown struct{}

func (self OperationHalfFrameDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	rows := count * frame.Height() / 2
	NormalCursorDown{}.Execute(editor, rows)
	pos := frame.TopLeft()
	content_height := len(editor.curwin.buffer.Lines())
	pos.row += max(min(rows, content_height-frame.bot), 0)
	editor.curwin.frame = frame.Shift(pos)
}

type OperationHalfFrameUp struct{}

func (self OperationHalfFrameUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	rows := count * frame.Height() / 2
	NormalCursorUp{}.Execute(editor, rows)
	pos := frame.TopLeft()
	pos.row = max(pos.row-rows, 0)
	editor.curwin.frame = frame.Shift(pos)
}

type OperationFrameLineUp struct{}

func (self OperationFrameLineUp) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	NormalCursorUp{}.Execute(editor, count)
	pos := Point{
		row: min(max(frame.top-count, 0), len(editor.curwin.buffer.Lines())-frame.Height()),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OperationFrameLineDown struct{}

func (self OperationFrameLineDown) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	NormalCursorDown{}.Execute(editor, count)
	pos := Point{
		row: min(max(frame.top+count, 0), len(editor.curwin.buffer.Lines())-frame.Height()),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OperationCenterFrame struct{}

func (self OperationCenterFrame) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	frame := editor.curwin.frame
	cursor := editor.curwin.cursor
	pos := Point{
		row: max(cursor.Row()-frame.Height()/2, 0),
		col: frame.left,
	}
	editor.curwin.frame = frame.Shift(pos)
}

type OperationSaveFile struct{}

func (self OperationSaveFile) Execute(editor *Editor, count int) {
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

type OperationStartNewLine struct{}

func (self OperationStartNewLine) Execute(editor *Editor, count int) {
	if editor.curwin == nil {
		return
	}
	AppendAtLineEnd{}.Execute(editor, count)
	editor.curwin.insertContent(false, editor.curwin.buffer.Nl_seq())
}

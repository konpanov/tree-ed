package main

import (
	"github.com/atotto/clipboard"
)

// import "github.com/gdamore/tcell/v2"

type Operation interface {
	Execute(editor *Editor, count int)
}

// type ClipboardOperation interface {
// }

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor, count int) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor, count int) {
	editor.curwin.cursorDown(count)
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor, count int) {
	editor.curwin.cursorUp(count)
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor, count int) {
	editor.curwin.cursorLeft(count)
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor, count int) {
	editor.curwin.cursorRight(count)
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor, count int) {
	editor.curwin.switchToInsert()
}

type SwitchToInsertModeAsAppend struct{}

func (self SwitchToInsertModeAsAppend) Execute(editor *Editor, count int) {
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight(1)
}

type SwitchToVisualmode struct{}

func (self SwitchToVisualmode) Execute(editor *Editor, count int) {
	editor.curwin.switchToVisual()
}

type SwitchToNormalMode struct{}

func (self SwitchToNormalMode) Execute(editor *Editor, count int) {
	editor.curwin.switchToNormal()
}

type SwitchFromInsertToNormalMode struct{}

func (self SwitchFromInsertToNormalMode) Execute(editor *Editor, count int) {
	editor.curwin.switchToNormal()
}

type SwitchToTreeMode struct{}

func (self SwitchToTreeMode) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		index := editor.curwin.cursor.Index()
		node := NodeLeaf(editor.curwin.buffer.Tree().RootNode(), index)
		editor.curwin.setNode(node, true)
		editor.curwin.switchToTree()
	}
}

type SwitchFromVisualToTreeMode struct{}

func (self SwitchFromVisualToTreeMode) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		start, end := editor.curwin.getSelection()
		node := MinimalNode(editor.curwin.buffer.Tree().RootNode(), start, end)
		editor.curwin.setNode(node, true)
		editor.curwin.switchToTree()
	}
}

type EraseLineAtCursor struct{}

func (self EraseLineAtCursor) Execute(editor *Editor, count int) {
	editor.curwin.eraseLineAtCursor(count)
}

type EraseCharNormalMode struct{}

func (self EraseCharNormalMode) Execute(editor *Editor, count int) {
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
	continue_last_erase bool
}

// TODO add composite modification?
func (self EraseCharInsertMode) Execute(editor *Editor, count int) {
	editor.curwin.eraseContent(self.continue_last_erase)
}

type InsertContentOperation struct {
	content              []byte
	continue_last_insert bool
}

func (self InsertContentOperation) Execute(editor *Editor, count int) {
	editor.curwin.insertContent(self.continue_last_insert, self.content)
}

type NodeUpOperation struct{}

func (self NodeUpOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeUp()
		}
	}
}

type NodeDownOperation struct{}

func (self NodeDownOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeDown()
		}
	}
}

type NodeNextSiblingOperation struct{}

func (self NodeNextSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSibling()
		}
	}
}

type NodeNextSiblingOrCousinOperation struct{}

func (self NodeNextSiblingOrCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextSiblingOrCousin()
		}
	}
}

type NodePrevSiblingOperation struct{}

func (self NodePrevSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSibling()
		}
	}
}

type NodePrevSiblingOrCousinOperation struct{}

func (self NodePrevSiblingOrCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevSiblingOrCousin()
		}
	}
}

type NodeFirstSiblingOperation struct{}

func (self NodeFirstSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToFirstSibling()
	}
}

type NodeLastSiblingOperation struct{}

func (self NodeLastSiblingOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		editor.curwin.nodeToLastSibling()
	}
}

type EraseSelectionOperation struct{}

func (self EraseSelectionOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	start, end := win.cursor.Index(), win.secondCursor.Index()
	start, end = min(start, end), max(start, end)
	change := NewEraseChange(win, start, end+1)
	// mod.cursorBefore = win.cursor.Index() // TMPCHANGE
	// mod.cursorAfter = win.cursor.Index() // TMPCHANGE
	change.Apply(win)
	win.undotree.Push(UndoState{change: change}, true)
	win.switchToNormal()
}

type UndoChangeOperation struct{}

func (self UndoChangeOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	for range count {
		if mod := win.undotree.Back(); mod != nil {
			mod.Reverse().Apply(win)
		}
	}
}

type RedoChangeOperation struct{}

func (self RedoChangeOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	for range count {
		if mod := win.undotree.Forward(); mod != nil {
			mod.Apply(win)
		}
	}
}

type WordStartForwardOperation struct{}

func (self WordStartForwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartNext(), true)
	}
}

type WordEndForwardOperation struct{}

func (self WordEndForwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndNext(), true)
	}
}

type WordEndBackwardOperation struct{}

func (self WordEndBackwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordEndPrev(), true)
	}
}

type WordBackwardOperation struct{}

func (self WordBackwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.setCursor(editor.curwin.cursor.WordStartPrev(), true)
	}
}

type LineEndOperation struct{}

func (self LineEndOperation) Execute(editor *Editor, count int) {
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowEnd(), true)

}

type LineStartOperation struct{}

func (self LineStartOperation) Execute(editor *Editor, count int) {
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowStart(), true)

}

type LineTextStartOperation struct{}

func (self LineTextStartOperation) Execute(editor *Editor, count int) {
	NormalCursorDown{}.Execute(editor, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.ToRowTextStart(), true)

}

type CountOperation struct {
	count int
	op    Operation
}

func (self CountOperation) Execute(editor *Editor, count int) {
	self.op.Execute(editor, self.count)
}

type GoOperation struct{}

func (self GoOperation) Execute(editor *Editor, count int) {
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.cursorAnchor
	pos.row = max(0, count-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type GoEndOperation struct{}

func (self GoEndOperation) Execute(editor *Editor, count int) {
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.cursorAnchor
	pos.row = max(0, len(editor.curwin.buffer.Lines())-1)
	editor.curwin.setCursor(editor.curwin.cursor.MoveToRunePos(pos), false)
}

type SwapNodeForwardEndOperation struct{}

func (self SwapNodeForwardEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		win := editor.curwin

		node := win.getNode()
		swapee := node
		for range count {
			if swapee = NextSiblingOrCousinDepth(swapee, win.anchorDepth); swapee == nil {
				return
			}
		}

		startA, endA := int(node.StartByte()), int(node.EndByte())
		startB, endB := int(swapee.StartByte()), int(swapee.EndByte())
		change := NewSwapChange(win, startA, endA, startB, endB)
		change.Apply(win)

		win.setCursor(win.cursor.ToIndex(-endA+startA+endB), true)
		win.secondCursor = win.secondCursor.ToIndex(endB - 1)
		win.undotree.Push(UndoState{change: change}, true)
	}
}

type SwapNodeBackwardEndOperation struct{}

func (self SwapNodeBackwardEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		win := editor.curwin

		node := win.getNode()
		swapee := node
		for range count {
			if swapee = PrevSiblingOrCousinDepth(swapee, win.anchorDepth); swapee == nil {
				return
			}
		}

		startA, endA := int(swapee.StartByte()), int(swapee.EndByte())
		startB, endB := int(node.StartByte()), int(node.EndByte())
		change := NewSwapChange(win, startA, endA, startB, endB)
		change.Apply(win)

		win.setCursor(win.cursor.ToIndex(startA), true)
		win.secondCursor = win.secondCursor.ToIndex(startA + endB - startB - 1)
		win.undotree.Push(UndoState{change: change}, true) // TMPCHANGE
	}
}

type PasteClipboardOperation struct{}

func (self PasteClipboardOperation) Execute(editor *Editor, count int) {
	if clipboard.Unsupported {
		return
	}
	text, err := clipboard.ReadAll()
	panic_if_error(err)
	editor.curwin.insertContent(false, []byte(text))
}

type CopyToClipboardOperation struct{}

func (self CopyToClipboardOperation) Execute(editor *Editor, count int) {
	if clipboard.Unsupported {
		return
	}
	win := editor.curwin
	start, end := win.getSelection()
	text := win.buffer.Content()[start:end]
	clipboard.WriteAll(string(text))
}

type SlurpNodeOperation struct{}

func (self SlurpNodeOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	if win.buffer.Tree() == nil {
		return
	}
	node := win.getNode()
	parent := node
	if parent == nil {
		return
	}
	if parent.ChildCount() < 2 {
		return
	}
	last_siblling := parent.Child(parent.ChildCount() - 1)
	next := NextSiblingOrCousinDepth(parent, win.anchorDepth)
	if next == nil {
		return
	}

	startA, endA := int(last_siblling.StartByte()), int(last_siblling.EndByte())
	startB, endB := int(next.StartByte()), int(next.EndByte())
	change := NewSwapChange(win, startA, endA, startB, endB)
	change.Apply(win)
	win.undotree.Push(UndoState{change: change}, true) //TMPCHANGE
	node = win.getNode()
	win.setNode(node, true)
}

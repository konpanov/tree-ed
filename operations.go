package main

import ()

type Operation interface {
	Execute(editor *Editor, count int)
}

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
	new_cursor := editor.curwin.cursor.RunePrev()
	if new_cursor.RunePosition().row == editor.curwin.cursor.RunePosition().row {
		editor.curwin.cursor = new_cursor
	}
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
	win := editor.curwin
	composite := CompositeChange{}
	for range count {
		if win.cursor.IsNewLine() {
			break
		}
		mod := NewEraseRuneModification(win, win.cursor.Index())
		mod.cursorBefore = win.cursor.Index()
		mod.cursorAfter = win.cursor.Index()
		mod.Apply(win)
		composite.changes = append(composite.changes, mod)
	}
	win.undotree.Push(composite)

}

type EraseCharInsertMode struct {
	continue_last_erase bool
}

// TODO add composite modification?
func (self EraseCharInsertMode) Execute(editor *Editor, count int) {
	var err error
	win := editor.curwin
	for range count {
		if !win.cursor.IsBegining() {
			win.cursor = win.cursor.RunePrev()
			panic_if_error(err)
			mod := NewEraseRuneModification(win, win.cursor.Index())
			mod.cursorBefore = win.cursor.Index()
			mod.cursorAfter = win.cursor.Index()
			mod.Apply(win)
			win.undotree.Push(mod)
		}
	}
}

type InsertContent struct {
	content              []byte
	continue_last_insert bool
}

func (self InsertContent) Execute(editor *Editor, count int) {
	editor.curwin.insertContent(self.continue_last_insert, self.content)
}

type InsertNewLine struct{}

func (self InsertNewLine) Execute(editor *Editor, count int) {
	win := editor.curwin
	for range count {
		mod := NewReplacementModification(win.cursor.Index(), []byte{}, win.buffer.Nl_seq())
		mod.cursorBefore = win.cursor.Index()
		mod.cursorAfter = win.cursor.Index() + len(mod.after)
		mod.Apply(win)
		win.undotree.Push(mod)
	}
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

type NodeNextCousinOperation struct{}

func (self NodeNextCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodeNextCousin()
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

type NodePrevCousinOperation struct{}

func (self NodePrevCousinOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		for range count {
			editor.curwin.nodePrevCousin()
		}
	}
}

type EraseSelectionOperation struct{}

func (self EraseSelectionOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	start, end := win.cursor.Index(), win.secondCursor.Index()
	start, end = min(start, end), max(start, end)
	mod := NewEraseModification(win, start, end+1)
	mod.cursorBefore = win.cursor.Index()
	mod.cursorAfter = win.cursor.Index()
	mod.Apply(win)
	win.undotree.Push(mod)
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
		editor.curwin.cursor = editor.curwin.cursor.WordStartNext()
	}
}

type WordEndForwardOperation struct{}

func (self WordEndForwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.cursor = editor.curwin.cursor.WordEndNext()
	}
}

type WordEndBackwardOperation struct{}

func (self WordEndBackwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.cursor = editor.curwin.cursor.WordEndPrev()
	}
}

type WordBackwardOperation struct{}

func (self WordBackwardOperation) Execute(editor *Editor, count int) {
	for range count {
		editor.curwin.cursor = editor.curwin.cursor.WordStartPrev()
	}
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
	var err error
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.cursorAnchor
	pos.row = max(0, count-1)
	editor.curwin.cursor, err = editor.curwin.cursor.MoveToRunePos(pos)
	panic_if_error(err)
}

type ShiftNodeForwardEndOperation struct{}

func (self ShiftNodeForwardEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		var err error
		win := editor.curwin

		node := win.getNode()
		cousin := node
		for range count {
			if cousin = NextCousinDepth(cousin, win.anchorDepth); cousin == nil {
				return
			}
		}

		startA, endA := int(node.StartByte()), int(node.EndByte())
		startB, endB := int(cousin.StartByte()), int(cousin.EndByte())
		change := NewSwapChange(win, startA, endA, startB, endB)
		change.Apply(win)

		win.cursor, err = win.cursor.ToIndex(-endA + startA + endB)
		panic_if_error(err)
		win.secondCursor, err = win.secondCursor.ToIndex(endB - 1)
		panic_if_error(err)
		win.undotree.Push(change)
	}
}

type ShiftNodeBackwardEndOperation struct{}

func (self ShiftNodeBackwardEndOperation) Execute(editor *Editor, count int) {
	if editor.curwin.buffer.Tree() != nil {
		var err error
		win := editor.curwin

		node := win.getNode()
		cousin := node
		for range count {
			if cousin = PrevCousinDepth(cousin, win.anchorDepth); cousin == nil {
				return
			}
		}

		startA, endA := int(cousin.StartByte()), int(cousin.EndByte())
		startB, endB := int(node.StartByte()), int(node.EndByte())
		change := NewSwapChange(win, startA, endA, startB, endB)
		change.Apply(win)

		win.cursor, err = win.cursor.ToIndex(startA)
		panic_if_error(err)
		win.secondCursor, err = win.secondCursor.ToIndex(startA + endB - startB - 1)
		panic_if_error(err)
		win.undotree.Push(change)
	}
}

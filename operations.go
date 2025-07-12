package main

import (
	"log"
	// sitter "github.com/smacker/go-tree-sitter"
)

type Operation interface {
	Execute(editor *Editor, count int)
}

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor, count int) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor, count int) {
	var err error
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.cursorAnchor
	pos.row += count
	editor.curwin.cursor, err = editor.curwin.cursor.MoveToRunePos(pos)
	panic_if_error(err)
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.cursorUp()
	}
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.cursorLeft()
	}
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.cursorRight()
	}
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor, count int) {
	editor.curwin.switchToInsert()
}

type SwitchToInsertModeAsAppend struct{}

func (self SwitchToInsertModeAsAppend) Execute(editor *Editor, count int) {
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight()
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
	new_cursor, _ := editor.curwin.cursor.RunesBackward(1)
	if new_cursor.RunePosition().row == editor.curwin.cursor.RunePosition().row {
		editor.curwin.cursor = new_cursor
	}
}

type SwitchToTreeMode struct{}

func (self SwitchToTreeMode) Execute(editor *Editor, count int) {
	editor.curwin.switchToTree()
}

type EraseLineAtCursor struct{}

func (self EraseLineAtCursor) Execute(editor *Editor, count int) {
	win := editor.curwin
	composite := CompositeModification{}
	pos := win.cursor.RunePosition()
	for i := 0; i < count; i++ {
		mod := NewEraseLineModification(win, win.cursor.BytePosition().row)
		mod.cursorBefore = win.cursor.Index()
		mod.Apply(win)
		win.cursor, _ = win.cursor.MoveToRunePos(Point{pos.row, win.cursorAnchor})
		mod.cursorAfter = win.cursor.Index()
		composite.modifications = append(composite.modifications, mod)
	}
	win.undotree.Push(composite)
}

type EraseCharNormalMode struct{}

func (self EraseCharNormalMode) Execute(editor *Editor, count int) {
	win := editor.curwin
	composite := CompositeModification{}
	for i := 0; i < count && !win.cursor.IsNewLine(); i++ {
		mod := NewEraseRuneModification(win, win.cursor.Index())
		mod.cursorBefore = win.cursor.Index()
		mod.cursorAfter = win.cursor.Index()
		mod.Apply(win)
		composite.modifications = append(composite.modifications, mod)
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
	for i := 0; i < count; i++ {
		if !win.cursor.IsBegining() {
			win.cursor, err = win.cursor.RunesBackward(1)
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
	win := editor.curwin
	if self.continue_last_insert {
		if mod := win.undotree.Back(); mod != nil {
			mod.Reverse().Apply(win)
		}
	}
	mod := NewReplacementModification(win.cursor.Index(), []byte{}, self.content)
	mod.cursorBefore = win.cursor.Index()
	mod.cursorAfter = win.cursor.Index() + len(mod.after)
	mod.Apply(win)
	win.undotree.Push(mod)
}

type InsertNewLine struct{}

func (self InsertNewLine) Execute(editor *Editor, count int) {
	win := editor.curwin
	for i := 0; i < count; i++ {
		mod := NewReplacementModification(win.cursor.Index(), []byte{}, win.buffer.Nl_seq())
		mod.cursorBefore = win.cursor.Index()
		mod.cursorAfter = win.cursor.Index() + len(mod.after)
		mod.Apply(win)
		win.undotree.Push(mod)
	}
}

type NodeUpOperation struct{}

func (self NodeUpOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodeUp()
	}
}

type NodeDownOperation struct{}

func (self NodeDownOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodeDown()
	}
}

type NodeNextSiblingOperation struct{}

func (self NodeNextSiblingOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodeNextSibling()
	}
}

type NodeNextCousinOperation struct{}

func (self NodeNextCousinOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodeNextCousin()
	}
}

type NodePrevSiblingOperation struct{}

func (self NodePrevSiblingOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodePrevSibling()
	}
}

type NodePrevCousinOperation struct{}

func (self NodePrevCousinOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.nodePrevCousin()
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
	for i := 0; i < count; i++ {
		if mod := win.undotree.Back(); mod != nil {
			mod.Reverse().Apply(win)
		}
	}
}

type RedoChangeOperation struct{}

func (self RedoChangeOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	for i := 0; i < count; i++ {
		if mod := win.undotree.Forward(); mod != nil {
			mod.Apply(win)
		}
	}
}

type WordStartForwardOperation struct{}

func (self WordStartForwardOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Word start forward")
		if cursor, err := editor.curwin.cursor.WordStartForward(); err == nil || err == ErrReachBufferEnd {
			editor.curwin.cursor = cursor
		}
	}
}

type WordEndForwardOperation struct{}

func (self WordEndForwardOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Word end forward")
		if cursor, err := editor.curwin.cursor.WordEndForward(); err == nil || err == ErrReachBufferEnd {
			editor.curwin.cursor = cursor
		}
	}
}

type WordBackwardOperation struct{}

func (self WordBackwardOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Word backward")
		if cursor, err := editor.curwin.cursor.WordStartBackward(); err == nil || err == ErrReachBufferBeginning {
			editor.curwin.cursor = cursor
		}
	}
}

type CountOperation struct {
	count int
	op    Operation
}

func (self CountOperation) Execute(editor *Editor, count int) {
	self.op.Execute(editor, self.count)
}

type GoOperation struct {
}

func (self GoOperation) Execute(editor *Editor, count int) {
	var err error
	pos := editor.curwin.cursor.RunePosition()
	pos.col = editor.curwin.cursorAnchor
	pos.row = max(0, count-1)
	editor.curwin.cursor, err = editor.curwin.cursor.MoveToRunePos(pos)
	panic_if_error(err)
}

type ShiftNodeForwardOperation struct {
	continuous bool
}

func (self ShiftNodeForwardOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	node := win.shift_node
	cousin := NextCousin(node)
	if cousin == nil {
		return
	}

	// Erase selected text
	start, end := order(win.cursor.Index(), win.secondCursor.Index())
	mod1 := NewEraseModification(win, start, end+1)
	mod1.cursorBefore = win.cursor.Index()
	mod1.cursorAfter = win.cursor.Index()
	mod1.Apply(win)

	// Paste erased text after the sibling
	mod2 := ReplacementModification{
		at:           int(cousin.EndByte()) - len(mod1.before),
		before:       []byte{},
		after:        mod1.before,
		cursorBefore: win.cursor.Index(),
		cursorAfter:  win.cursor.Index(),
	}
	mod2.Apply(win)

	// Move selection to newly pasted text
	var err error
	offset := int(cousin.EndByte()) - end - 1
	win.cursor, err = win.cursor.ToIndex(win.cursor.Index() + offset)
	win.secondCursor, err = win.secondCursor.ToIndex(win.secondCursor.Index() + offset)
	panic_if_error(err)

	win.shift_node = cousin
	comp_mod := CompositeModification{[]Modification{mod1, mod2}}
	win.undotree.Push(comp_mod)

	// mod.cursorBefore = win.cursor.Index()
	// mod.cursorAfter = win.cursor.Index()
	// mod.Apply(win)
	// win.undotree.Push(mod)
	// win.switchToNormal()
}

type ShiftNodeBackwardOperation struct {
	continuous bool
}

func (self ShiftNodeBackwardOperation) Execute(editor *Editor, count int) {
	win := editor.curwin
	node := win.shift_node
	cousin := PrevCousin(node)
	if cousin == nil {
		return
	}

	// Erase selected text
	start, end := order(win.cursor.Index(), win.secondCursor.Index())
	mod1 := NewEraseModification(win, start, end+1)
	mod1.cursorBefore = win.cursor.Index()
	mod1.cursorAfter = win.cursor.Index()
	mod1.Apply(win)

	// Paste erased text after the sibling
	mod2 := ReplacementModification{
		at:           int(cousin.EndByte()) - len(mod1.before),
		before:       []byte{},
		after:        mod1.before,
		cursorBefore: win.cursor.Index(),
		cursorAfter:  win.cursor.Index(),
	}
	mod2.Apply(win)

	// Move selection to newly pasted text
	var err error
	offset := int(cousin.EndByte()) - end - 1
	win.cursor, err = win.cursor.ToIndex(win.cursor.Index() + offset)
	win.secondCursor, err = win.secondCursor.ToIndex(win.secondCursor.Index() + offset)
	panic_if_error(err)

	win.shift_node = cousin
	comp_mod := CompositeModification{[]Modification{mod1, mod2}}
	win.undotree.Push(comp_mod)

	// mod.cursorBefore = win.cursor.Index()
	// mod.cursorAfter = win.cursor.Index()
	// mod.Apply(win)
	// win.undotree.Push(mod)
	// win.switchToNormal()
}

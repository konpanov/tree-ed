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
	for i := 0; i < count; i++{
		editor.curwin.cursorDown()
	}
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++{
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
	for i := 0; i < count; i++ {
		row := editor.curwin.cursor.BytePosition().row
		editor.curwin.cursor.buffer.EraseLine(row)
	}
}

type EraseCharNormalMode struct{}

func (self EraseCharNormalMode) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		cur := editor.curwin.cursor
		buf := editor.curwin.buffer

		if cur.IsBegining() && cur.IsEnd() {
			log.Printf("Cannot remove from empty buffer\n")
			return
		}

		if cur.IsNewLine() {
			log.Printf("Should not erase new lines in normal mode\n")
			return
		}

		err := buf.EraseRune(cur.Index())
		panic_if_error(err)

		if cur.IsNewLine() && !cur.IsLineStart() {
			editor.curwin.cursor, _ = cur.RunesBackward(1)
		}

		log.Println("Erased succesfully")
	}

}

type EraseCharInsertMode struct {
	continue_last_erase bool
}

func (self EraseCharInsertMode) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		if !editor.curwin.cursor.IsBegining() {
			prev, err := editor.curwin.cursor.RunesBackward(1)
			panic_if_error(err)
			err = editor.curwin.buffer.EraseRune(prev.Index())
			panic_if_error(err)
			editor.curwin.cursor = prev
		}
	}
}

type InsertContent struct {
	content              []byte
	start                BufferCursor
	continue_last_insert bool
}

func (self InsertContent) Execute(editor *Editor, count int) {
	curwin := editor.curwin
	editInput := ReplacementInput{
		start:       curwin.cursor.Index(),
		end:         curwin.cursor.Index(),
		replacement: self.content,
	}

	err := editor.curwin.buffer.Edit(editInput)
	panic_if_error(err)
	changes := editor.curwin.buffer.Changes()
	last_change := changes[len(changes)-1]
	editor.curwin.cursor, _ = editor.curwin.cursor.UpdateToChange(last_change)
}

type InsertNewLine struct {
	char rune
}

func (self InsertNewLine) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		editor.curwin.insert(editor.curwin.buffer.Nl_seq())
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

type DeleteSelectionOperation struct{}

func (self DeleteSelectionOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Deleteing node")
		win := editor.curwin
		region := NewRegion(win.cursor.index, win.secondCursor.index)
		region.end++
		win.deleteRange(region)
		win.cursor.ToIndex(region.start)
		win.secondCursor.ToIndex(region.start)
		win.switchToNormal()
	}
}

type UndoChangeOperation struct{}

func (self UndoChangeOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Undoing a change")
		buffer := editor.curwin.buffer
		if buffer.Undo() == nil {
			change := buffer.Changes()[buffer.ChangeIndex()]
			editor.curwin.cursor, _ = editor.curwin.cursor.ToIndex(change.old_end_index - 1)
		}
	}
}

type RedoChangeOperation struct{}

func (self RedoChangeOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Redoing a change")
		buffer := editor.curwin.buffer
		if buffer.Redo() == nil {
			change := buffer.Changes()[buffer.ChangeIndex()-1]
			editor.curwin.cursor, _ = editor.curwin.cursor.ToIndex(change.old_end_index)
		}
	}
}

type WordForwardOperation struct{}

func (self WordForwardOperation) Execute(editor *Editor, count int) {
	for i := 0; i < count; i++ {
		log.Println("Word forward")
		if cursor, err := editor.curwin.cursor.WordStartForward(); err == nil || err == ErrReachBufferEnd {
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

type CountOperation struct{
	count int
	op Operation
}

func (self CountOperation) Execute(editor *Editor, count int){
	self.op.Execute(editor, self.count)
}

type GoOperation struct{
}

func (self GoOperation) Execute(editor *Editor, count int){
	row := editor.curwin.cursor.RunePosition().row
	for row < min(count-1, len(editor.curwin.buffer.Lines())){
		editor.curwin.cursorDown()
		row = editor.curwin.cursor.RunePosition().row
	}
	for row > max(count-1, 0){
		editor.curwin.cursorUp()
		row = editor.curwin.cursor.RunePosition().row
	}
}

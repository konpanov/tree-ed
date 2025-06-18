package main

import (
	"log"
)

type Operation interface {
	Execute(editor *Editor)
}

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor) {
	editor.curwin.cursorDown()
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor) {
	editor.curwin.cursorUp()
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor) {
	editor.curwin.cursorLeft()
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor) {
	editor.curwin.cursorRight()
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor) {
	editor.curwin.switchToInsert()
}

type SwitchToInsertModeAsAppend struct{}

func (self SwitchToInsertModeAsAppend) Execute(editor *Editor) {
	editor.curwin.switchToInsert()
	editor.curwin.cursorRight()
}

type SwitchToVisualmode struct{}

func (self SwitchToVisualmode) Execute(editor *Editor) {
	editor.curwin.switchToVisual()
}

type SwitchToNormalMode struct{}

func (self SwitchToNormalMode) Execute(editor *Editor) {
	editor.curwin.switchToNormal()
}

type SwitchToTreeMode struct{}

func (self SwitchToTreeMode) Execute(editor *Editor) {
	editor.curwin.switchToTree()
}

type EraseLineAtCursor struct{}

func (self EraseLineAtCursor) Execute(editor *Editor) {
	row := editor.curwin.cursor.BytePosition().row
	editor.curwin.cursor.buffer.EraseLine(row)
}

type EraseCharNormalMode struct{}

func (self EraseCharNormalMode) Execute(editor *Editor) {
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

type EraseCharInsertMode struct {
	continue_last_erase bool
}

func (self EraseCharInsertMode) Execute(editor *Editor) {
	if !editor.curwin.cursor.IsBegining() {
		prev, err := editor.curwin.cursor.RunesBackward(1)
		panic_if_error(err)
		err = editor.curwin.buffer.EraseRune(prev.Index())
		panic_if_error(err)
		editor.curwin.cursor = prev

		if self.continue_last_erase {
			err := editor.curwin.buffer.MergeLastChanges()
			panic_if_error(err)
		}
	}
}

type InsertContent struct {
	content              []byte
	start                BufferCursor
	continue_last_insert bool
}

func (self InsertContent) Execute(editor *Editor) {
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

	if self.continue_last_insert {
		err := editor.curwin.buffer.MergeLastChanges()
		panic_if_error(err)
	}
}

type InsertNewLine struct {
	char rune
}

func (self InsertNewLine) Execute(editor *Editor) {
	editor.curwin.insert(editor.curwin.buffer.Nl_seq())
}

type NodeUpOperation struct{}

func (self NodeUpOperation) Execute(editor *Editor) {
	editor.curwin.nodeUp()
}

type NodeDownOperation struct{}

func (self NodeDownOperation) Execute(editor *Editor) {
	editor.curwin.nodeDown()
}

type NodeRightOperation struct{}

func (self NodeRightOperation) Execute(editor *Editor) {
	editor.curwin.nodeRight()
}

type NodeLeftOperation struct{}

func (self NodeLeftOperation) Execute(editor *Editor) {
	editor.curwin.nodeLeft()
}

type DeleteSelectionOperation struct{}

func (self DeleteSelectionOperation) Execute(editor *Editor) {
	log.Println("Deleteing node")
	win := editor.curwin
	region := NewRegion(win.cursor.index, win.secondCursor.index)
	region.end++
	win.deleteRange(region)
	win.cursor.ToIndex(region.start)
	win.secondCursor.ToIndex(region.start)
	win.switchToNormal()
}

type UndoChangeOperation struct{}

func (self UndoChangeOperation) Execute(editor *Editor) {
	log.Println("Undoing a change")
	buffer := editor.curwin.buffer
	changes := buffer.Changes()
	change_index := buffer.ChangeIndex()
	if len(changes) == 0 || change_index == 0 {
		log.Println("No changes to undo")
		return
	}
	last_change := changes[change_index-1]
	log.Println(string(last_change.before))
	editor.curwin.cursor, _ = editor.curwin.cursor.UpdateToChange(last_change.Reverse())
	editor.curwin.buffer.Undo()
}

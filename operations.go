package main

import "log"

type Operation interface {
	Execute(editor *Editor)
}

type QuitOperation struct{}

func (self QuitOperation) Execute(editor *Editor) {
	editor.is_quiting = true
}

type NormalCursorDown struct{}

func (self NormalCursorDown) Execute(editor *Editor) {
	editor.current_window.cursorDown()
}

type NormalCursorUp struct{}

func (self NormalCursorUp) Execute(editor *Editor) {
	editor.current_window.cursorUp()
}

type NormalCursorLeft struct{}

func (self NormalCursorLeft) Execute(editor *Editor) {
	editor.current_window.cursorLeft()
}

type NormalCursorRight struct{}

func (self NormalCursorRight) Execute(editor *Editor) {
	editor.current_window.cursorRight()
}

type SwitchToInsertMode struct{}

func (self SwitchToInsertMode) Execute(editor *Editor) {
	editor.current_window.switchToInsert()
}

type SwitchToInsertModeAsAppend struct{}

func (self SwitchToInsertModeAsAppend) Execute(editor *Editor) {
	editor.current_window.switchToInsert()
	editor.current_window.cursorRight()
}

type SwitchToVisualmode struct{}

func (self SwitchToVisualmode) Execute(editor *Editor) {
	editor.current_window.switchToVisual()
}

type SwitchToNormalMode struct{}

func (self SwitchToNormalMode) Execute(editor *Editor) {
	editor.current_window.switchToNormal()
}

type SwitchToTreeMode struct{}

func (self SwitchToTreeMode) Execute(editor *Editor) {
	editor.current_window.switchToTree()
}

type EraseLineAtCursor struct{}

func (self EraseLineAtCursor) Execute(editor *Editor) {
	row := editor.current_window.cursor.BytePosition().row
	editor.current_window.cursor.buffer.EraseLine(row)
}

type EraseCharAtCursor struct{}

func (self EraseCharAtCursor) Execute(editor *Editor) {
	editor.current_window.remove()
}

type ReplaceSequence struct {
	region   Region
	sequence []byte
}

func (self ReplaceSequence) Execute(editor *Editor) {
	editor.current_window.remove()
}

type InsertChar struct {
	char rune
}

func (self InsertChar) Execute(editor *Editor) {
	editor.current_window.insert([]byte(string(self.char)))
}

type InsertNewLine struct {
	char rune
}

func (self InsertNewLine) Execute(editor *Editor) {
	editor.current_window.insert(editor.current_window.buffer.Nl_seq())
}

type NodeUpOperation struct{}

func (self NodeUpOperation) Execute(editor *Editor) {
	editor.current_window.nodeUp()
}

type NodeDownOperation struct{}

func (self NodeDownOperation) Execute(editor *Editor) {
	editor.current_window.nodeDown()
}

type NodeRightOperation struct{}

func (self NodeRightOperation) Execute(editor *Editor) {
	editor.current_window.nodeRight()
}

type NodeLeftOperation struct{}

func (self NodeLeftOperation) Execute(editor *Editor) {
	editor.current_window.nodeLeft()
}

type DeleteSelectionOperation struct{}

func (self DeleteSelectionOperation) Execute(editor *Editor) {
	log.Println("Deleteing node")
	win := editor.current_window
	region := NewRegion(win.cursor.index, win.secondCursor.index)
	region.end++
	win.deleteRange(region)
	win.cursor.ToIndex(region.start)
	win.secondCursor.ToIndex(region.start)
	win.switchToNormal()
}

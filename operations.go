package main

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

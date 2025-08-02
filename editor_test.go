package main

import (
	"testing"
)

func TestEditorInsertNewLine(t *testing.T) {
	nl := NewLineUnix
	buffer := mkTestBuffer(t, "", nl)
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1         ",
		"          ",
		"----------",
		"file: , li",
	})
	assertPointsEqual(t, editor.curwin.cursor.RunePosition(), Point{0, 0})
	InsertContentOperation{content: []byte("\r"), continue_last_insert: false}.Execute(editor, 1)
	assertScreenRunes(t, editor.screen, []string{
		"1         ",
		"2         ",
		"----------",
		"file: , li",
	})
	assertPointsEqual(t, editor.curwin.cursor.RunePosition(), Point{1, 0})
}

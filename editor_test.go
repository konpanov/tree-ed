package main

import (
	"testing"
	// "github.com/gdamore/tcell/v2"
)

// TODO: TEST STUFF

func TestEditorInsertNewLine(t *testing.T) {
	nl := LineBreakPosix
	buffer := mkTestBuffer(t, "ą", string(nl))
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1 ą       ",
		"          ",
		"[N1:1 100%",
		"          ",
	})
	OpSaveClipbaord{}.Execute(editor, 1)
	OpPasteClipboard{}.Execute(editor, 1)
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1 ąą      ",
		"          ",
		"[N1:2 100%",
		"          ",
	})
}

// func TestEditorSelectNodeOverScreen(t *testing.T) {
// 	nl := NewLineUnix
// 	buffer := mkTestBuffer(t, "//abcdeghklmnop", nl)
// 	screen := mkTestScreen(t, "")
// 	screen.SetSize(10, 4)
// 	editor := NewEditor(screen)
// 	editor.OpenBuffer(buffer)
// 	editor.Redraw()
// 	assertScreenRunes(t, editor.screen, []string{
// 		"1 //abcdegh",
// 		"          ",
// 		"----------",
// 		"file: , li",
// 	})
// 	assertPointsEqual(t, editor.curwin.cursor.RunePosition(), Point{0, 0})
// }

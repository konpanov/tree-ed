package main

// import (
// 	"testing"
//
// 	"github.com/gdamore/tcell/v2"
// )

// TODO: TEST STUFF

// func TestEditorInsertNewLine(t *testing.T) {
// 	nl := NewLineUnix
// 	buffer := mkTestBuffer(t, "", nl)
// 	screen := mkTestScreen(t, "")
// 	screen.SetSize(10, 4)
// 	editor := NewEditor(screen)
// 	editor.OpenBuffer(buffer)
// 	editor.Redraw()
// 	assertScreenRunes(t, editor.screen, []string{
// 		"1         ",
// 		"          ",
// 		"----------",
// 		"file: , li",
// 	})
// 	assertPointsEqual(t, editor.curwin.cursor.RunePosition(), Point{0, 0})
//
// 	eventEnter := *tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone)
// 	InsertContentOperation{
// 		content:              []tcell.EventKey{eventEnter},
// 		continue_last_insert: false,
// 	}.Execute(editor, 1)
// 	assertScreenRunes(t, editor.screen, []string{
// 		"1         ",
// 		"2         ",
// 		"----------",
// 		"file: , li",
// 	})
// 	assertPointsEqual(t, editor.curwin.cursor.RunePosition(), Point{1, 0})
// }

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

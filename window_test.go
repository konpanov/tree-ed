package main

import (
	"slices"
	"testing"
)

var helloworld = as_content([]string{
	" package main",
	"",
	"func main() {",
	"	print(\"Hello, World!\")",
	"}",
}, string(LineBreakUnix))

func TestWindowEraseAtCursor(t *testing.T) {
	buffer := mkTestBuffer(t, string(helloworld), string(LineBreakUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(LineBreakUnix))

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.LineBreak(), LineBreakUnix)

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Pos{0, 0}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowEraseAtCursorWithAnchor(t *testing.T) {
	buffer := mkTestBuffer(t, string(helloworld), string(LineBreakUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.cursorRight(9)
	window.eraseLineAtCursor(1)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(LineBreakUnix))

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.LineBreak(), LineBreakUnix)
	assertPositionsEqual(t, window.cursor.Pos(), Pos{0, 9})
	assertIntEqualMsg(t, window.originColumn, 9, "Unexpected cursor anchor: ")

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Pos{0, 9}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowUndoUnicodeInsert(t *testing.T) {
	buffer := mkTestBuffer(t, "", string(LineBreakUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent(false, []byte("П"))
	window.insertContent(false, []byte("р"))

	window.history.Back().Reverse().Apply(window)

	expected := as_content([]string{"П"}, string(LineBreakUnix))
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

func TestWindowContinuouUnicodeInsertWithErase(t *testing.T) {
	buffer := mkTestBuffer(t, "", string(LineBreakUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent(false, []byte("Пр"))
	window.eraseContent(true)
	expected := as_content([]string{"П"}, string(LineBreakUnix))
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

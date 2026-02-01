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
}, string(LF))

func TestWindowEraseAtCursor(t *testing.T) {
	buffer := mkTestBuffer(t, string(helloworld), string(LF))
	window := windowFromBuffer(buffer, 10, 10)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(LF))

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.LineBreak(), LF)

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Pos{0, 0}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowEraseAtCursorWithAnchor(t *testing.T) {
	buffer := mkTestBuffer(t, string(helloworld), string(LF))
	window := windowFromBuffer(buffer, 10, 10)
	window.cursorRight(9)
	window.eraseLineAtCursor(1)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(LF))

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.LineBreak(), LF)
	assertPositionsEqual(t, window.cursor.Pos(), Pos{0, 0})
	assertIntEqualMsg(t, window.originColumn, 0, "Unexpected cursor origin column: ")

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Pos{0, 0}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowUndoUnicodeInsert(t *testing.T) {
	buffer := mkTestBuffer(t, "", string(LF))
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent([]byte("П"))
	window.insertContent([]byte("р"))

	window.history.Back().Reverse().Apply(window)

	expected := as_content([]string{"П"}, string(LF))
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

func TestWindowContinuouUnicodeInsertWithErase(t *testing.T) {
	buffer := mkTestBuffer(t, "", string(LF))
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent([]byte("Пр"))
	window.eraseContent()
	expected := as_content([]string{"П"}, string(LF))
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

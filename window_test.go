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
}, NewLineUnix)

func TestWindowEraseAtCursor(t *testing.T) {
	nl_seq := []byte(NewLineUnix)
	buffer := mkTestBuffer(t, string(helloworld), NewLineUnix)
	window := windowFromBuffer(buffer, 10, 10)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, NewLineUnix)

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.Nl_seq(), nl_seq)

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Point{0, 0}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowEraseAtCursorWithAnchor(t *testing.T) {
	nl_seq := []byte(NewLineUnix)
	buffer := mkTestBuffer(t, string(helloworld), NewLineUnix)
	window := windowFromBuffer(buffer, 10, 10)
	window.cursorRight(9)
	window.eraseLineAtCursor(1)
	window.eraseLineAtCursor(1)
	expected := as_content([]string{
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, NewLineUnix)

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.Nl_seq(), nl_seq)
	assertPointsEqual(t, window.cursor.RunePosition(), Point{0, 9})
	assertIntEqualMsg(t, window.originColumn, 9, "Unexpected cursor anchor: ")

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Point{0, 9}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

func TestWindowUndoUnicodeInsert(t *testing.T) {
	buffer := mkTestBuffer(t, "", NewLineUnix)
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent(false, []byte("П"))
	window.insertContent(false, []byte("р"))

	window.undotree.Back().Reverse().Apply(window)

	expected := as_content([]string{"П"}, NewLineUnix)
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

func TestWindowContinuouUnicodeInsertWithErase(t *testing.T) {
	buffer := mkTestBuffer(t, "", NewLineUnix)
	window := windowFromBuffer(buffer, 10, 10)
	window.switchToInsert()
	window.insertContent(false, []byte("Пр"))
	window.eraseContent(true)
	expected := as_content([]string{"П"}, NewLineUnix)
	if slices.Compare(expected, buffer.Content()) != 0 {
		t.Errorf("\"%+v\" != \"%+v\"\n", buffer.Content(), expected)
	}
	assertBytesEqual(t, buffer.Content(), []byte(expected))
}

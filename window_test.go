package main

import (
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
	window := windowFromBuffer(buffer)
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
	window := windowFromBuffer(buffer)
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
	assertIntEqualMsg(t, window.cursorAnchor, 9, "Unexpected cursor anchor: ")

	cursor_pos := window.cursor.Index()
	expected_pos := window.cursor.MoveToRunePos(Point{0, 9}).Index()
	if cursor_pos != expected_pos {
		t.Errorf("Cursor is in unexpected position. Expected %+v, got %+v", expected_pos, cursor_pos)
	}
}

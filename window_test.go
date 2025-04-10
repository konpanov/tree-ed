package main

import (
	"testing"
)

func TestInsertEmptyContent(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.Content(), value)
}

func TestDeleteLinesAndInsertEmptyContent(t *testing.T) {
	content := []byte("line\nline\n")
	buffer, _ := bufferFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	lines := window.buffer.Lines()
	window.deleteRange(Region{start: lines[1].start, end: lines[1].end + len(buffer.nl_seq)})
	window.deleteRange(Region{start: lines[0].start, end: lines[0].end + len(buffer.nl_seq)})
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.Content(), value)
}

func TestWindowMoveRightCursorOnNonasciiCharacters(t *testing.T) {
	// ą ć ż
	// 012345
	content := []byte("ąćż")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.moveCursor(Right)
	assertIntEqualMsg(t, window.cursor.index, 2, "Expected cursor to move to move from 'ą' to 'ć': ")
}

func TestWindowMoveLeftCursorOnNonasciiCharacters(t *testing.T) {
	// ą ć ż
	// 012345
	content := []byte("ąćż")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.cursor, _ = window.cursor.ToIndex(4)
	window.moveCursor(Left)
	assertIntEqualMsg(t, window.cursor.index, 2, "Expected cursor to move to move from 'ż' to 'ć': ")
}

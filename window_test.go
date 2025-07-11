package main

import (
	"strings"
	"testing"
)

func TestWindowMoveRightCursorOnNonasciiCharacters(t *testing.T) {
	// ą ć ż
	// 012345
	content := []byte("ąćż")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer)
	window.cursorRight()
	assertIntEqualMsg(t, window.cursor.index, 2, "Expected cursor to move to move from 'ą' to 'ć': ")
}

func TestWindowMoveLeftCursorOnNonasciiCharacters(t *testing.T) {
	// ą ć ż
	// 012345
	content := []byte("ąćż")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer)
	window.cursor, _ = window.cursor.ToIndex(4)
	window.cursorLeft()
	assertIntEqualMsg(t, window.cursor.index, 2, "Expected cursor to move to move from 'ż' to 'ć': ")
}

func TestWindowMoveCursorDownAndSaveAnchor(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1longer",
		"line2",
		"line3longer",
	}
	content := strings.Join(lines, nl)
	buffer, _ := bufferFromContent([]byte(content), []byte(nl))
	window := windowFromBuffer(buffer)
	for i := 0; i < 7; i++ {
		window.cursorRight()
	}
	assertIntEqualMsg(t, window.cursor.Index(), 7, "Expected cursor to be move right before moving down")
	for i := 0; i < 2; i++ {
		window.cursorDown()
	}
	assertIntEqualMsg(t, window.cursor.Index(), 25, "Expected cursor to be under anchor cursor")
}

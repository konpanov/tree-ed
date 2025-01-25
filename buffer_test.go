package main

import (
	"bytes"
	"testing"
)

func assertIntEqual(t *testing.T, a int, b int) {
	if a != b {
		t.Errorf("%d != %d", a, b)
	}
}

func assertBytesEqual(t *testing.T, a []byte, b []byte) {
	if !bytes.Equal(a, b) {
		t.Errorf("%s != %s", string(a), string(b))
	}
}

func TestInsertCharacter(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\n"))
	buffer.insert(2, []byte{'X'})
	assertBytesEqual(t, []byte("abXcdef"), buffer.content)
}

func TestInsertCharacterAtTheStart(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\n"))
	buffer.insert(0, []byte{'X'})
	assertBytesEqual(t, []byte("Xabcdef"), buffer.content)
}

func TestInsertCharacterBeforeTheEnd(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\n"))
	buffer.insert(5, []byte{'X'})
	assertBytesEqual(t, []byte("abcdeXf"), buffer.content)
}

func TestInsertCharacterAtTheEnd(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\n"))
	buffer.insert(6, []byte{'X'})
	assertBytesEqual(t, []byte("abcdefX"), buffer.content)
}

func TestInsertCharacterBeforeNewLine(t *testing.T) {
	buffer := bufferFromContent([]byte("ab\ncdef"), []byte("\n"))
	buffer.insert(2, []byte{'X'})
	assertBytesEqual(t, []byte("abX\ncdef"), buffer.content)
}

func TestInsertCharacterAfterNewLine(t *testing.T) {
	buffer := bufferFromContent([]byte("ab\ncdef"), []byte("\n"))
	buffer.insert(3, []byte{'X'})
	assertBytesEqual(t, []byte("ab\nXcdef"), buffer.content)
}

func TestInsertCharacterAtEmptyLine(t *testing.T) {
	buffer := bufferFromContent([]byte("ab\n\ncdef"), []byte("\n"))
	buffer.insert(3, []byte{'X'})
	assertBytesEqual(t, []byte("ab\nX\ncdef"), buffer.content)
}

func TestInsertMultipleCharacters(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\n"))
	buffer.insert(2, []byte("XY"))
	assertBytesEqual(t, []byte("abXYcdef"), buffer.content)
}

func TestInsertWindowsNewLine(t *testing.T) {
	buffer := bufferFromContent([]byte("abcdef"), []byte("\r\n"))
	buffer.insert(2, []byte("\r\n\r\n"))
	assertBytesEqual(t, []byte("ab\r\n\r\ncdef"), buffer.content)
	assertIntEqual(t, 3, len(buffer.lines))
}

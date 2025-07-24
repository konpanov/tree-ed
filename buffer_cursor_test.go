package main

import (
	"strings"
	"testing"
)

func TestBufferCursorIndexAtTheBeginning(t *testing.T) {
	nl := NewLineUnix
	content := "line1"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}
	assertIntEqualMsg(t, cursor.Index(), 0, "Expected cursor to be at the begining: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{0, 0})
	assertPointsEqual(t, cursor.RunePosition(), Point{0, 0})
}

func TestBufferCursorAfterMovementToTheNextByte(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.BytesForward(1)
	assertIntEqualMsg(t, cursor.Index(), 1, "Expected cursor to be at the second byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 1})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 1})
}

func TestBufferCursorAfterMovementForwardAndBackward(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}
	cursor = cursor.BytesForward(3)
	assertNoErrors(t, err)
	cursor = cursor.BytesBackward(1)
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 2, "Expected cursor to be at the third byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 2})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 2})
}

func TestBufferCursorIsNewLine(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)

	cursor := BufferCursor{buffer: buffer, index: 0}
	if cursor.IsNewLine() {
		t.Errorf("Expected cursor not to be on new line")
	}
	cursor = cursor.AsEdge().BytesForward(5)
	assertNoErrors(t, err)
	if !cursor.IsNewLine() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
}

func TestBufferSearchForward(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)

	cursor, err := BufferCursor{buffer: buffer, index: 0}.AsEdge().SearchForward(buffer.Nl_seq())
	if !cursor.IsNewLine() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
	assertIntEqualMsg(t, cursor.Index(), 5, "Expected cursor to be at the fifth byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 5})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 5})
}

func TestBufferSearchBackward(t *testing.T) {
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)

	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(14)
	assertNoErrors(t, err)
	cursor, err = cursor.AsEdge().SearchBackward(buffer.Nl_seq())
	if !cursor.IsNewLine() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
	assertIntEqualMsg(t, cursor.Index(), 11, "Expected cursor to be at the 10th byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 1, col: 5})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 1, col: 5})
}

func TestBufferCursorRunesForwardOnce(t *testing.T) {
	nl := NewLineUnix
	content := "aąłb"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(1)
	assertNoErrors(t, err)
	cursor = cursor.RuneNext()
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 3, "Expected cursor to be at the third byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 3})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 2})
}

func TestBufferCursorMultipleRunesForward(t *testing.T) {
	nl := NewLineUnix
	content := "aąłbźg"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(1)
	assertNoErrors(t, err)
	for range 3 {
		cursor = cursor.RuneNext()
	}
	assertIntEqualMsg(t, cursor.Index(), 6, "Expected cursor to be at the third byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 6})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 4})
}

func TestBufferCursorMultipleRunesBackward(t *testing.T) {
	nl := NewLineUnix
	content := "aąłbźg"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(8)
	assertNoErrors(t, err)
	for range 3 {
		cursor = cursor.RunePrev()
	}
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 3, "Expected cursor to be at the third byte: ")
	assertPointsEqual(t, cursor.BytePosition(), Point{row: 0, col: 3})
	assertPointsEqual(t, cursor.RunePosition(), Point{row: 0, col: 2})
}

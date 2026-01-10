package main

import (
	"strings"
	"testing"
)

func TestBufferCursorIndexAtTheBeginning(t *testing.T) {
	nl := LineBreakPosix
	content := "line1"
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}
	assertIntEqualMsg(t, cursor.Index(), 0, "Expected cursor to be at the begining: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{0, 0})
}

func TestBufferCursorAfterMovementToTheNextByte(t *testing.T) {
	nl := LineBreakPosix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, string(nl))
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.BytesForward(1)
	assertIntEqualMsg(t, cursor.Index(), 1, "Expected cursor to be at the second byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 1})
}

func TestBufferCursorAfterMovementForwardAndBackward(t *testing.T) {
	nl := LineBreakPosix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, string(nl))
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}
	cursor = cursor.BytesForward(3)
	assertNoErrors(t, err)
	cursor = cursor.BytesBackward(1)
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 2, "Expected cursor to be at the third byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 2})
}

func TestBufferCursorIsNewLine(t *testing.T) {
	nl := LineBreakPosix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, string(nl))
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)

	cursor := BufferCursor{buffer: buffer, index: 0}
	if cursor.IsLineBreak() {
		t.Errorf("Expected cursor not to be on new line")
	}
	cursor = cursor.AsEdge().BytesForward(5)
	assertNoErrors(t, err)
	if !cursor.IsLineBreak() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
}

func TestBufferSearchForward(t *testing.T) {
	nl := LineBreakPosix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, string(nl))
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)

	cursor, err := BufferCursor{buffer: buffer, index: 0}.AsEdge().SearchForward(buffer.LineBreak())
	if !cursor.IsLineBreak() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
	assertIntEqualMsg(t, cursor.Index(), 5, "Expected cursor to be at the fifth byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 5})
}

func TestBufferSearchBackward(t *testing.T) {
	nl := LineBreakPosix
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, string(nl))
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)

	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(14)
	assertNoErrors(t, err)
	cursor, err = cursor.AsEdge().SearchBackward(buffer.LineBreak())
	if !cursor.IsLineBreak() {
		t.Errorf("Expected cursor to be on new line, rune: %+q", buffer.Content()[cursor.Index()])
	}
	assertIntEqualMsg(t, cursor.Index(), 11, "Expected cursor to be at the 10th byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 1, col: 5})
}

func TestBufferCursorRunesForwardOnce(t *testing.T) {
	nl := LineBreakPosix
	content := "aąłb"
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(1)
	assertNoErrors(t, err)
	cursor = cursor.RuneNext()
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 3, "Expected cursor to be at the third byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 2})
}

func TestBufferCursorMultipleRunesForward(t *testing.T) {
	nl := LineBreakPosix
	content := "aąłbźg"
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(1)
	assertNoErrors(t, err)
	for range 3 {
		cursor = cursor.RuneNext()
	}
	assertIntEqualMsg(t, cursor.Index(), 6, "Expected cursor to be at the third byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 4})
}

func TestBufferCursorMultipleRunesBackward(t *testing.T) {
	nl := LineBreakPosix
	content := "aąłbźg"
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.ToIndex(8)
	assertNoErrors(t, err)
	for range 3 {
		cursor = cursor.RunePrev()
	}
	assertNoErrors(t, err)
	assertIntEqualMsg(t, cursor.Index(), 3, "Expected cursor to be at the third byte: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{row: 0, col: 2})
}

func TestBufferCursorPrevRuneOnWindowsLineBreak(t *testing.T) {
	content := "abc\r\nedf"
	buffer, err := bufferFromContent([]byte(content), LineBreakWindows, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 5}.AsEdge()
	cursor = cursor.RunePrev()
	if cursor.Index() != 3 {
		t.Errorf("Cursor should be on index 3, but was on %d\n", cursor.Index())
	}
}

func TestBufferCursorNextRuneOnWindowsLineBreak(t *testing.T) {
	content := "abc\r\nedf"
	buffer, err := bufferFromContent([]byte(content), LineBreakWindows, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 3}.AsEdge()
	cursor = cursor.RuneNext()
	if cursor.Index() != 5 {
		t.Errorf("Cursor should be on index 5, but was on %d\n", cursor.Index())
	}
}

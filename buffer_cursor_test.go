package main

import (
	"strings"
	"testing"
)

func TestBufferCursorIndexAtTheBeginning(t *testing.T) {
	nl := LF
	content := "line1"
	buffer, err := bufferFromContent([]byte(content), nl, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}
	assertIntEqualMsg(t, cursor.Index(), 0, "Expected cursor to be at the begining: ")
	assertPositionsEqual(t, cursor.Pos(), Pos{0, 0})
}

func TestBufferCursorAfterMovementToTheNextByte(t *testing.T) {
	nl := LF
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
	nl := LF
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
	nl := LF
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
	nl := LF
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
	nl := LF
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
	nl := LF
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
	nl := LF
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
	nl := LF
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
	buffer, err := bufferFromContent([]byte(content), CRLF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 5}.AsEdge()
	cursor = cursor.RunePrev()
	if cursor.Index() != 3 {
		t.Errorf("Cursor should be on index 3, but was on %d\n", cursor.Index())
	}
}

func TestBufferCursorNextRuneOnWindowsLineBreak(t *testing.T) {
	content := "abc\r\nedf"
	buffer, err := bufferFromContent([]byte(content), CRLF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 3}.AsEdge()
	cursor = cursor.RuneNext()
	if cursor.Index() != 5 {
		t.Errorf("Cursor should be on index 5, but was on %d\n", cursor.Index())
	}
}

func TestBufferCursorRuneClass(t *testing.T) {
	content := "a.:1 *\t"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar()
	if class, expected := cursor.Class(), RuneClassChar; class != expected {
		t.Errorf("Expected rune class %+v, but got %+v", expected, class)
	}
	cursor = cursor.RuneNext()
	if class, expected := cursor.Class(), RuneClassPunct; class != expected {
		t.Errorf("Expected rune class %+v, but got %+v", expected, class)
	}
	cursor = cursor.RuneNext()
	if class, expected := cursor.Class(), RuneClassPunct; class != expected {
		t.Errorf("Expected rune class %+v, but got %+v", expected, class)
	}
	cursor = cursor.RuneNext()
	if class, expected := cursor.Class(), RuneClassChar; class != expected {
		t.Errorf("Expected rune class %+v, but got %+v", expected, class)
	}
	cursor = cursor.RuneNext()
	if class, expected := cursor.Class(), RuneClassSpace; class != expected {
		t.Errorf("Expected rune class %+v, but got %+v", expected, class)
	}
}

func TestBufferCursorWordStartNext(t *testing.T) {
	content := "abc (123)  !@\n    edf"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar()
	cursor = cursor.WordStartNext()
	if index, expected := cursor.Index(), 4; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartNext()
	if index, expected := cursor.Index(), 5; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartNext()
	if index, expected := cursor.Index(), 8; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartNext()
	if index, expected := cursor.Index(), 11; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartNext()
	if index, expected := cursor.Index(), 18; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
}

func TestBufferCursorWordStartPrev(t *testing.T) {
	content := "abc (123)  !@\n    edf"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar().ToIndex(20)
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 18; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 11; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 8; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 5; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 4; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordStartPrev()
	if index, expected := cursor.Index(), 0; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
}

func TestBufferCursorWordEndNext(t *testing.T) {
	content := "abc (123)  !@\n    edf"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar()
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 2; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 4; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 7; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 8; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 12; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndNext()
	if index, expected := cursor.Index(), 20; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
}

func TestBufferCursorWordEndPrev(t *testing.T) {
	content := "abc (123)  !@\n    edf"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar().ToIndex(20)
	cursor = cursor.WordEndPrev()
	if index, expected := cursor.Index(), 12; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndPrev()
	if index, expected := cursor.Index(), 8; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndPrev()
	if index, expected := cursor.Index(), 7; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndPrev()
	if index, expected := cursor.Index(), 4; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.WordEndPrev()
	if index, expected := cursor.Index(), 2; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
}

func TestBufferCursorToLineEnd(t *testing.T) {
	content := "abc\nedf\nzxc\njkl;"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar()
	cursor = cursor.ToLineEnd()
	if index, expected := cursor.Index(), 2; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(5)
	cursor = cursor.ToLineEnd()
	if index, expected := cursor.Index(), 6; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(10)
	cursor = cursor.ToLineEnd()
	if index, expected := cursor.Index(), 10; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(11)
	cursor = cursor.ToLineEnd()
	if index, expected := cursor.Index(), 10; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(12)
	cursor = cursor.ToLineEnd()
	if index, expected := cursor.Index(), 15; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}

}

func TestBufferCursorToLineStart(t *testing.T) {
	content := "abc\nedf\n  zxc\njkl;"
	buffer, err := bufferFromContent([]byte(content), LF, nil)
	assertNoErrors(t, err)
	cursor := BufferCursor{buffer: buffer, index: 0}.AsChar()
	cursor = cursor.ToLineStart()
	if index, expected := cursor.Index(), 0; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(5)
	cursor = cursor.ToLineStart()
	if index, expected := cursor.Index(), 4; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(10)
	cursor = cursor.ToLineStart()
	if index, expected := cursor.Index(), 8; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
	cursor = cursor.ToIndex(17)
	cursor = cursor.ToLineStart()
	if index, expected := cursor.Index(), 14; index != expected {
		t.Errorf("Expected index %+v, but got %+v", expected, index)
	}
}

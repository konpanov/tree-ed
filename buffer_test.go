package main

import (
	"testing"
)

func TestBufferCreateFromContent(t *testing.T) {
	content := []byte("content\nfor\ntesting\n")
	nl_seq := []byte("\n")
	buffer, err := bufferFromContent(content, nl_seq)
	assertNoErrors(t, err)
	assertBytesEqual(t, buffer.content, content)
	assertBytesEqual(t, buffer.nl_seq, nl_seq)
}

func TestBufferInsertWhenEmpty(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte(""), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("a"))
	assertBytesEqual(t, buffer.content, []byte("a"))

	buffer, err = bufferFromContent([]byte(""), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("ab"))
	assertBytesEqual(t, buffer.content, []byte("ab"))

	buffer, err = bufferFromContent([]byte(""), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("a\nc"))
	assertBytesEqual(t, buffer.content, []byte("a\nc"))
}

func TestBufferInsertAtTheBeginningOfALine(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("a"))
	assertBytesEqual(t, buffer.content, []byte("aoriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("abc"))
	assertBytesEqual(t, buffer.content, []byte("abcoriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("\nqwe\n"))
	assertBytesEqual(t, buffer.content, []byte("\nqwe\noriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\r\n"))
	assertNoErrors(t, err)
	buffer.Insert(0, []byte("\r\nqwe\r\n"))
	assertBytesEqual(t, buffer.content, []byte("\r\nqwe\r\noriginal"))
}

func TestBufferInsertAtTheEndOfALine(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(3, []byte("d"))
	assertBytesEqual(t, buffer.content, []byte("abcd"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(3, []byte("de"))
	assertBytesEqual(t, buffer.content, []byte("abcde"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"))
	assertNoErrors(t, err)
	buffer.Insert(3, []byte("\nde\n"))
	assertBytesEqual(t, buffer.content, []byte("abc\nde\n"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\r\n"))
	assertNoErrors(t, err)
	buffer.Insert(3, []byte("\r\nde\r\n"))
	assertBytesEqual(t, buffer.content, []byte("abc\r\nde\r\n"))
}

func TestBufferFailsOnIndexOutOfBound(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Insert(-1, []byte("test"))
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Insert(4, []byte("test"))
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}
}

func TestBufferErase(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Erase(Range{1, 4})
	if err != nil {
		t.Error(err)
	}
	assertBytesEqual(t, buffer.content, []byte("ae"))
}

func TestBufferEraseOutOfBound(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Erase(Range{-1, 4})
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Erase(Range{6, 8})
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Erase(Range{1, -3})
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"))
	assertNoErrors(t, err)
	err = buffer.Erase(Range{2, 12})
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}
}

func TestBufferFindCoordOnSingleLine(t *testing.T) {
	var err error
	nl := "\n"
	content := ""
	content += "first line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	coord, err := buffer.Coord(5)
	if err != nil {
		t.Error(err)
	}
	expected := Point{row: 0, col: 5}
	if coord != expected {
		t.Errorf("Recieved coordinates do not match expected value %#v != %#v", coord, expected)
	}
}

func TestBufferFindCoordOnSecondLine(t *testing.T) {
	var err error
	nl := "\n"
	content := ""
	content += "first line"
	content += nl
	content += "second line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	coord, err := buffer.Coord(17)
	if err != nil {
		t.Error(err)
	}
	expected := Point{row: 1, col: 6}
	if coord != expected {
		t.Errorf("Recieved coordinates do not match expected value %#v != %#v", coord, expected)
	}
}

func TestBufferFindCoordOnEmptyLine(t *testing.T) {
	var err error
	nl := "\n"
	content := ""
	content += "abcde"
	content += nl
	content += nl
	content += "third line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	coord, err := buffer.Coord(6)
	if err != nil {
		t.Error(err)
	}
	expected := Point{row: 1, col: 0}
	if coord != expected {
		t.Errorf("Recieved coordinates do not match expected value %#v != %#v", coord, expected)
	}
}

func TestBufferFindCoordAfterEmptyLine(t *testing.T) {
	var err error
	nl := "\n"
	content := ""
	content += "abcde"
	content += nl
	content += nl
	content += "third line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	coord, err := buffer.Coord(7)
	if err != nil {
		t.Error(err)
	}
	expected := Point{row: 2, col: 0}
	if coord != expected {
		t.Errorf("Recieved coordinates do not match expected value %#v != %#v", coord, expected)
	}
}

func TestBufferFindCoordOnEmptyLineWithWindowsNewLineSeq(t *testing.T) {
	var err error
	nl := "\r\n"
	content := ""
	content += "abcde"
	content += nl
	content += nl
	content += "third line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)

	coord, err := buffer.Coord(6)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 1, col: 0})

	coord, err = buffer.Coord(7)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 1, col: 0})

	coord, err = buffer.Coord(8)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 2, col: 0})
}

func TestBufferLineInfoOnContentWithoutNewLineAtTheEnd(t *testing.T) {
	nl := "\n"
	content := ""
	content += "abcde"
	content += nl
	content += "nopqr"
	content += nl
	content += "third line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	lines := buffer.Lines()
	expectedLength := 3
	if len(lines) != expectedLength {
		t.Errorf(
			"Expected line info to have length %d, but gut %d",
			expectedLength,
			len(lines),
		)
	}
	assertIntEqual(t, lines[0].start, 0)
	assertIntEqual(t, lines[0].end, 5)
	assertIntEqual(t, lines[1].start, 6)
	assertIntEqual(t, lines[1].end, 11)
	assertIntEqual(t, lines[2].start, 12)
	assertIntEqual(t, lines[2].end, 22)
}

func TestBufferLineInfoOnEmptyContent(t *testing.T) {
	nl := "\n"
	content := ""
	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	lines := buffer.Lines()
	expectedLength := 1
	if len(lines) != expectedLength {
		t.Errorf(
			"Expected line info to have length %d, but got %d",
			expectedLength,
			len(lines),
		)
	}
	assertIntEqual(t, lines[0].start, 0)
	assertIntEqual(t, lines[0].end, 0)
}

func TestBufferLineInfoOnEmptyLine(t *testing.T) {
	nl := "\n"
	content := ""
	content += "first line"
	content += nl
	content += nl
	content += "third line"

	buffer, err := bufferFromContent([]byte(content), []byte(nl))
	assertNoErrors(t, err)
	lines := buffer.Lines()
	expectedLength := 3
	if len(lines) != expectedLength {
		t.Errorf(
			"Expected line info to have length %d, but got %d",
			expectedLength,
			len(lines),
		)
	}
	assertIntEqual(t, lines[0].start, 0)
	assertIntEqual(t, lines[0].end, 10)
	assertIntEqual(t, lines[1].start, 11)
	assertIntEqual(t, lines[1].end, 11)
	assertIntEqual(t, lines[2].start, 12)
	assertIntEqual(t, lines[2].end, 22)
}

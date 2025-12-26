package main

import (
	"fmt"
	"strings"
	"testing"
)

func mkTestBuffer(t *testing.T, content string, nl string) IBuffer {
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	if err != nil {
		t.Fatalf("Failed to create test buffer")
	}
	return buffer
}

func TestBufferCreateFromContent(t *testing.T) {
	content := []byte("content\nfor\ntesting\n")
	nl_seq := []byte("\n")
	buffer, err := bufferFromContent(content, nl_seq, nil)
	assertNoErrors(t, err)
	assertBytesEqual(t, buffer.content, content)
	assertBytesEqual(t, buffer.nl_seq, nl_seq)
}

func TestBufferInsertWhenEmpty(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte(""), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("a")})
	assertBytesEqual(t, buffer.content, []byte("a"))

	buffer, err = bufferFromContent([]byte(""), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("ab")})
	assertBytesEqual(t, buffer.content, []byte("ab"))

	buffer, err = bufferFromContent([]byte(""), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("a\nc")})
	assertBytesEqual(t, buffer.content, []byte("a\nc"))
}

func TestBufferInsertAtTheBeginningOfALine(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("a")})
	assertBytesEqual(t, buffer.content, []byte("aoriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("abc")})
	assertBytesEqual(t, buffer.content, []byte("abcoriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("\nqwe\n")})
	assertBytesEqual(t, buffer.content, []byte("\nqwe\noriginal"))

	buffer, err = bufferFromContent([]byte("original"), []byte("\r\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{0, 0, []byte("\r\nqwe\r\n")})
	assertBytesEqual(t, buffer.content, []byte("\r\nqwe\r\noriginal"))
}

func TestBufferInsertAtTheEndOfALine(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{3, 3, []byte("d")})
	assertBytesEqual(t, buffer.content, []byte("abcd"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{3, 3, []byte("de")})
	assertBytesEqual(t, buffer.content, []byte("abcde"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{3, 3, []byte("\nde\n")})
	assertBytesEqual(t, buffer.content, []byte("abc\nde\n"))

	buffer, err = bufferFromContent([]byte("abc"), []byte("\r\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{3, 3, []byte("\r\nde\r\n")})
	assertBytesEqual(t, buffer.content, []byte("abc\r\nde\r\n"))
}

func TestBufferFailsOnIndexOutOfBound(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{-1, -1, []byte("test")})
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abc"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{4, 4, []byte("test")})
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}
}

func TestBufferErase(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"), nil)
	assertNoErrors(t, err)

	err = buffer.Edit(ReplacementInput{1, 4, []byte{}})
	if err != nil {
		t.Error(err)
	}
	assertBytesEqual(t, buffer.content, []byte("ae"))
}

func TestBufferEraseOutOfBound(t *testing.T) {
	var buffer *Buffer
	var err error

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{-1, 4, []byte{}})
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{6, 8, []byte{}})
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{1, -3, []byte{}})
	if err != ErrIndexLessThanZero {
		t.Error("Expected ErrIndexLessThanZero")
	}

	buffer, err = bufferFromContent([]byte("abcde"), []byte("\n"), nil)
	assertNoErrors(t, err)
	err = buffer.Edit(ReplacementInput{2, 12, []byte{}})
	if err != ErrIndexGreaterThanBufferSize {
		t.Error("Expected ErrIndexGreaterThanBufferSize")
	}
}

func TestBufferFindCoordOnSingleLine(t *testing.T) {
	var err error
	nl := "\n"
	content := ""
	content += "first line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)

	coord, err := buffer.Coord(6)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 0, col: 6})

	coord, err = buffer.Coord(7)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 1, col: 0})

	coord, err = buffer.Coord(8)
	assertNoErrors(t, err)
	assertPointsEqual(t, coord, Point{row: 1, col: 1})
}

func TestBufferLineInfoOnContentWithoutNewLineAtTheEnd(t *testing.T) {
	nl := "\n"
	content := ""
	content += "abcde"
	content += nl
	content += "nopqr"
	content += nl
	content += "third line"
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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

func TestBufferLineInfoOnContentEndingOnNewLine(t *testing.T) {
	nl := "\n"
	content := ""
	content += "abcde"
	content += nl
	content += "nopqr"
	content += nl
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	lines := buffer.Lines()
	expectedLength := 2
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
	assertIntEqualMsg(t, lines[1].end, 11, "Unexpected end of the second line: ")
}

func TestBufferLineInfoOnContentEndingOnNewLineWindowVersion(t *testing.T) {
	nl := NewLineWindows
	content := ""
	content += "abcde"
	content += nl
	content += "nopqr"
	content += nl
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	lines := buffer.Lines()
	expectedLength := 2
	if len(lines) != expectedLength {
		t.Errorf(
			"Expected line info to have length %d, but gut %d",
			expectedLength,
			len(lines),
		)
	}
	assertIntEqual(t, lines[0].start, 0)
	assertIntEqual(t, lines[0].end, 5)
	assertIntEqual(t, lines[1].start, 7)
	assertIntEqualMsg(t, lines[1].end, 12, "Unexpected end of the second line: ")
}

func TestBufferLineInfoOnEmptyContent(t *testing.T) {
	nl := "\n"
	content := ""
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
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

func TestBufferRuneCoordWithoutNonAsciiRunes(t *testing.T) {
	nl := "\n"
	content := ""
	//          0123456789
	content += "first line"
	//         10
	content += nl
	//         11
	content += nl
	//          1214161820
	content += "third line"
	//           1315171921

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	coord, err := buffer.RuneCoord(18)
	assertIntEqualMsg(t, coord.row, 2, "Unexpected rune coord row: ")
	assertIntEqualMsg(t, coord.col, 6, "Unexpected rune coord col: ")
}

func TestBufferRuneCoordWithNonAsciiRunes(t *testing.T) {
	nl := "\n"
	content := ""
	//          0123456789
	content += "first line"
	//         10
	content += nl
	//         11
	content += nl
	//          12141618222527
	content += "third ążćline"
	//           131517202426

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	coord, err := buffer.RuneCoord(20)
	assertIntEqualMsg(t, coord.row, 2, "Unexpected rune coord row: ")
	assertIntEqualMsg(t, coord.col, 7, "Unexpected rune coord col: ")
}

func TestBufferRuneCoordInbetweenNewLines(t *testing.T) {
	nl := "\n"
	content := ""
	//          0123456789
	content += "first line"
	//         10
	content += nl
	//         11
	content += nl
	//          12141618222527
	content += "third ążćline"
	//           131517202426

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	coord, err := buffer.RuneCoord(11)
	assertIntEqualMsg(t, coord.row, 1, "Unexpected rune coord row: ")
	assertIntEqualMsg(t, coord.col, 0, "Unexpected rune coord col: ")
}

func TestBufferRuneCoordFileEndingNewLine(t *testing.T) {
	nl := "\n"
	content := ""
	//          0123456789	  10
	content += "first line" + nl
	//         11
	content += nl
	//          12141618222527   28
	content += "third ążćline" + nl
	//           131517202426

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	coord, err := buffer.RuneCoord(28)
	assertIntEqualMsg(t, coord.row, 2, "Unexpected rune coord row: ")
	assertIntEqualMsg(t, coord.col, 13, "Unexpected rune coord col: ")
}

func TestBufferIndexFromRuneCoord(t *testing.T) {
	nl := "\n"
	lines := []string{
		"line1",
		"line2",
		"line3",
	}
	content := strings.Join(lines, nl)

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	index := buffer.IndexFromRuneCoord(Point{row: 1, col: 2})
	assertIntEqualMsg(t, index, 8, "Unexpected index: ")
}

func TestBufferIndexFromRuneCoordWithUnevenRunes(t *testing.T) {
	nl := "\n"
	lines := []string{
		"ląne1",
		"łońe2",
		"line3",
	}
	content := strings.Join(lines, nl)

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	index := buffer.IndexFromRuneCoord(Point{row: 1, col: 2})
	assertIntEqualMsg(t, index, 10, "Unexpected index: ")
}

func TestBufferIndexFromRuneCoordWithEmptyLine(t *testing.T) {
	nl := "\n"
	lines := []string{
		"ląne1",
		"",
		"line3",
	}
	content := strings.Join(lines, nl)

	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrors(t, err)
	index := buffer.IndexFromRuneCoord(Point{row: 1, col: 0})
	assertIntEqualMsg(t, index, 7, "Unexpected index: ")
}

func TestBufferIndexFromRuneCoordOutsideTheLine(t *testing.T) {
	var err error
	nl := "\n"
	content := strings.Join([]string{"line 1", "line 2", "line 3"}, nl)
	buffer, err := bufferFromContent([]byte(content), []byte(nl), nil)
	assertNoErrorsMsg(t, err, "Could not create buffer from content: ")
	index := buffer.IndexFromRuneCoord(Point{row: 1, col: 20})
	expected := 14
	if index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}

	index = buffer.IndexFromRuneCoord(Point{row: 2, col: 20})
	expected = 20
	if index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestBufferTestEmptyContentLines(t *testing.T) {
	buffer, err := NewEmptyBuffer([]byte("\n"), nil)
	assertNoErrorsMsg(t, err, "Cound not create empty buffer")
	lines := buffer.Lines()
	if len(lines) != 1 && lines[0].start != 0 && lines[0].end != 0 {
		t.Errorf("Unexpected lines. %+v", lines)
	}
}

func BenchmarkBufferReadBigFile(b *testing.B) {
	content := []byte("package main\n\n")
	for range 1000000 {
		content = append(content, []byte("func main() {\n print(\"Hello, World\")\n}\n")...)
	}
	fmt.Printf("Bigfile size: %d MB\n", len(content)/(1024*1024))
	for range b.N {
		_, _ = bufferFromContent(content, []byte("\n"), nil)
	}
}

func TestBufferCalculateLinesInit(t *testing.T) {
	buffer, err := NewEmptyBuffer([]byte("\n"), nil)
	assertNoErrorsMsg(t, err, "Cound not create empty buffer")
	buffer.Edit(ReplacementInput{0, 0, []byte("hello\n")})
	if len(buffer.lines) != 1 {
		t.Errorf("There should be only 1 line, got %d\n", len(buffer.lines))
	}

}

func TestBufferCalculateLinesAfterDeletingEmptySecondLine(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"",
		"whatever",
	}, "\n")
	buffer, _ := bufferFromContent([]byte(content), []byte(NewLineUnix), nil)
	buffer.Edit(ReplacementInput{start: 12, end: 13, replacement: []byte{}})
	lines := buffer.Lines()
	if len(lines) != 2 {
		t.Errorf("Expected lines to have 2 line, but got %d\n", len(lines))
	}
	if lines[0].start != 0 {
		t.Errorf("Expected first line to start from 0, but got %d", lines[0].start)
	}
}

func TestBufferCalculateLinesAfterInsertingNewlinAtTheEnd(t *testing.T) {

	content := strings.Join([]string{
		"01234",
		"",
		"",
	}, "\n")
	buffer, _ := bufferFromContent([]byte(content), []byte(NewLineUnix), nil)
	input := []byte{'x', '\n'}
	buffer.Edit(ReplacementInput{start: 7, end: 7, replacement: input})
}

package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestInsertEmptyContent(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.content, value)
}

func TestDeleteLinesAndInsertEmptyContent(t *testing.T) {
	content := []byte("line\nline\n")
	buffer, _ := bufferFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	lines := window.buffer.Lines()
	window.deleteRange(Range{start: lines[1].start, end: lines[1].end + len(buffer.nl_seq)})
	window.deleteRange(Range{start: lines[0].start, end: lines[0].end + len(buffer.nl_seq)})
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.content, value)
}

func TestDrawEmptyContentInNormalMode(t *testing.T) {
	content := []byte("")
	screen, err := tcell.NewScreen()
	if err != nil {
		t.Fatalf("Could not create screen")
	}
	buffer, _ := bufferFromContent(content, []byte("\n"))
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	window.switchToNormal()
	window.draw(screen)
}

func TestDrawEmptyContentInInsertMode(t *testing.T) {
	content := []byte("")
	screen, err := tcell.NewScreen()
	if err != nil {
		t.Fatalf("Could not create screen")
	}
	buffer, _ := bufferFromContent(content, []byte("\n"))
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	window.switchToInsert()
	window.draw(screen)
}

func TestWindowFitCursorWithoutNeedToShift(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 10, 10)
	cursor := WindowCursor{
		index:               0,
		row:                 5,
		col:                 5,
		originColumn:        5,
		invalidOriginColumn: false,
	}
	window.shift_to_fit_cursor(cursor)
	if window.topLine != 0 || window.leftColumn != 0 {
		t.Errorf("Expected to not shift window")
	}
}

func TestWindowFitCursorVerticalDown(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 10, 10)
	cursor := WindowCursor{
		index:               0,
		row:                 15,
		col:                 5,
		originColumn:        5,
		invalidOriginColumn: false,
	}
	window.shift_to_fit_cursor(cursor)

	expectedTopLine := 6
	expectedLeftColumn := 0
	if window.topLine != expectedTopLine {
		t.Errorf("Expected to shift window to topLine = %d, but topLine = %d", expectedTopLine, window.topLine)
	}
	if window.leftColumn != expectedLeftColumn {
		t.Errorf("Expected to shift window to leftColumn = %d, but leftColumn = %d", expectedLeftColumn, window.leftColumn)
	}
}

func TestWindowFitCursorVerticalBackUp(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 10, 10)
	window.shift_to_fit_cursor(WindowCursor{
		index:               0,
		row:                 30,
		col:                 5,
		originColumn:        5,
		invalidOriginColumn: false,
	})
	window.shift_to_fit_cursor(WindowCursor{
		index:               0,
		row:                 15,
		col:                 5,
		originColumn:        5,
		invalidOriginColumn: false,
	})

	expectedTopLine := 15
	expectedLeftColumn := 0
	if window.topLine != expectedTopLine {
		t.Errorf("Expected to shift window to topLine = %d, but topLine = %d", expectedTopLine, window.topLine)
	}
	if window.leftColumn != expectedLeftColumn {
		t.Errorf("Expected to shift window to leftColumn = %d, but leftColumn = %d", expectedLeftColumn, window.leftColumn)
	}
}

func TestWindowFitCursorHorizontalRight(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferFromContent(content, []byte(NewLineUnix))
	window := windowFromBuffer(buffer, 12, 10)
	cursor := WindowCursor{
		index:               0,
		row:                 5,
		col:                 15,
		originColumn:        5,
		invalidOriginColumn: false,
	}
	window.shift_to_fit_cursor(cursor)

	expectedTopLine := 0
	expectedLeftColumn := 6
	if window.topLine != expectedTopLine {
		t.Errorf("Expected to shift window to topLine = %d, but topLine = %d", expectedTopLine, window.topLine)
	}
	if window.leftColumn != expectedLeftColumn {
		t.Errorf("Expected to shift window to leftColumn = %d, but leftColumn = %d", expectedLeftColumn, window.leftColumn)
	}
}

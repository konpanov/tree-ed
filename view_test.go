package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func mkTestScreen(t *testing.T, charset string) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen(charset)
	if s == nil {
		t.Fatalf("Failed to get simulation screen")
	}
	if e := s.Init(); e != nil {
		t.Fatalf("Failed to initialize screen: %v", e)
	}
	s.SetSize(20, 20)
	return s
}

func assertCells(t *testing.T, cells []tcell.SimCell, text []rune, msg string) {
	if len(cells) != len(text) {
		format := "%sExpected text length (%d) does not match cell slice length (%d)"
		t.Errorf(format, msg, len(text), len(cells))
	} else {
		for i, value := range text {
			if slices.Compare(cells[i].Runes, []rune{value}) != 0 {
				t.Errorf("%sUnexpected cell %#U, expected %#U\n", msg, cells[i].Runes, value)
			}
		}
	}
}

func assertScreenText(t *testing.T, screen tcell.SimulationScreen, pos Point, text []rune, msg string) {
	cells, width, _ := screen.GetContents()
	start := width*pos.row + pos.col
	assertCells(t, cells[start:start+len(text)], text, msg)
}

func assertCellsEmpty(t *testing.T, cells []tcell.SimCell) {
	for _, cell := range cells {
		if slices.Compare(cell.Runes, []rune(" ")) != 0 {
			t.Errorf("Unexpected nonempty cell %#U\n", cell.Runes)
		}
	}
}

// TEXT VIEW

func TestDrawSingleLineTextView(t *testing.T) {
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	text := []rune("hello")
	var view View2
	view = NewTextView2(screen, Rect{Point{0, 0}, Point{10, 10}}, [][]rune{text})
	view.Draw()

	screen.Show()
	cells, _, _ := screen.GetContents()
	assertCells(t, cells[:5], text, "")
	assertCellsEmpty(t, cells[5:])
}

func TestDrawDoubleLineTextView(t *testing.T) {
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	line1 := []rune("hello")
	line2 := []rune("world")
	var view View2
	view = NewTextView2(screen, Rect{Point{0, 0}, Point{10, 10}}, [][]rune{line1, line2})
	view.Draw()

	screen.Show()
	cells, width, _ := screen.GetContents()
	assertCells(t, cells[:len(line1)], line1, "")
	assertCells(t, cells[width:width+len(line2)], line2, "")

	assertCellsEmpty(t, cells[len(line1):width])
	assertCellsEmpty(t, cells[width+len(line2):])
}

func TestDrawDoubleLineTextViewWithOffsetRoi(t *testing.T) {
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	line1 := []rune("hello")
	line2 := []rune("world")
	var view View2
	view = NewTextView2(screen, Rect{Point{3, 4}, Point{10, 10}}, [][]rune{line1, line2})
	view.Draw()

	screen.Show()
	cells, width, _ := screen.GetContents()
	start := 3*width + 4
	assertCells(t, cells[start:start+len(line1)], line1, "")
	assertCells(t, cells[start+width:start+width+len(line2)], line2, "")

	assertCellsEmpty(t, cells[start+len(line1):start+width])
	assertCellsEmpty(t, cells[start+width+len(line2):])
}

func TestDrawSkippedLineTextView(t *testing.T) {
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	line1 := []rune("hello")
	line2 := []rune("")
	line3 := []rune("world")

	var view View2
	view = NewTextView2(screen, Rect{Point{0, 0}, Point{10, 10}}, [][]rune{line1, line2, line3})
	view.Draw()

	screen.Show()
	cells, width, _ := screen.GetContents()
	assertCells(t, cells[:len(line1)], line1, "")
	assertCellsEmpty(t, cells[len(line1):width])
	assertCellsEmpty(t, cells[width:width*2])
	assertCells(t, cells[width*2:width*2+len(line3)], line3, "")
	assertCellsEmpty(t, cells[width*2+len(line3):])
}

// WINDOW VIEW

func TestDrawWindowViewWithSingleLine(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	content := "hello"
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	roi := Rect{Point{0, 0}, Point{10, 10}}
	window := NewWindowView(screen, roi, buffer, NewBufferCursor(buffer), NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	cells, _, _ := screen.GetContents()
	assertCells(t, cells[:5], []rune(content), "")
	assertCellsEmpty(t, cells[5:])
}

func TestDrawWindowViewWithOverflowHeightLine(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 3)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()
	roi := Rect{Point{0, 0}, Point{h, w}}
	window := NewWindowView(screen, roi, buffer, NewBufferCursor(buffer), NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertIntEqual(t, w, 10)
	assertIntEqual(t, h, 3)
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune(lines[0]), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune(lines[1]), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune(lines[2]), "")
}

func TestDrawWindowViewWithNonAsciiCharacters(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"ląćź2",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()
	roi := Rect{Point{0, 0}, Point{h, w}}
	var window View2
	window = NewWindowView(screen, roi, buffer, NewBufferCursor(buffer), NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune(lines[0]), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune(lines[1]), "")
}

func TestDrawWindowViewWithVerticalTextOffset(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(5, 2)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)
	cursor, _ := NewBufferCursor(buffer).ToIndex(20)

	// Setup window
	w, h := screen.Size()
	assertIntEqualMsg(t, w, 5, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, cursor.Index(), 20, "")
	roi := Rect{Point{0, 0}, Point{h, w}}
	var window View2
	window = NewWindowView(screen, roi, buffer, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("line3"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("line4"), "")
}

func TestDrawWindowViewWithVerticalTextOffsetAndReturn(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(5, 2)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)
	cursor, _ := NewBufferCursor(buffer).ToIndex(20)

	// Setup window
	w, h := screen.Size()
	assertIntEqualMsg(t, w, 5, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, cursor.Index(), 20, "")
	roi := Rect{Point{0, 0}, Point{h, w}}
	window := NewWindowView(screen, roi, buffer, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	cursor, _ = cursor.ToIndex(14)
	assertIntEqualMsg(t, cursor.Index(), 14, "")
	window.Update(roi, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()
	screen.Show()

	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("line3"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("line4"), "")

	cursor, _ = cursor.ToIndex(8)
	assertIntEqualMsg(t, cursor.Index(), 8, "")
	window.Update(roi, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()
	screen.Show()

	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("line2"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("line3"), "")
}

func TestDrawWindowViewWithHorizontalTextOffset(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(2, 5)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"ląne3",
		"liźe4",
		"line5",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)
	cursor, _ := NewBufferCursor(buffer).RunesForward(4)

	// Setup window
	w, h := screen.Size()
	assertIntEqualMsg(t, w, 2, "")
	assertIntEqualMsg(t, h, 5, "")
	assertIntEqualMsg(t, cursor.Index(), 4, "")
	roi := Rect{Point{0, 0}, Point{h, w}}
	window := NewWindowView(screen, roi, buffer, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertPointsEqual(t, window.text_offset, Point{col: 3, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("e1"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("e2"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("e3"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("e4"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("e5"), "")
}

func TestDrawWindowViewWithHorizontalTextOffsetAndReturn(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(2, 5)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
		"ląne3",
		"liźe4",
		"line5",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)
	cursor, _ := NewBufferCursor(buffer).RunesForward(4)

	// Setup window
	w, h := screen.Size()
	assertIntEqualMsg(t, w, 2, "")
	assertIntEqualMsg(t, h, 5, "")
	assertIntEqualMsg(t, cursor.Index(), 4, "")
	roi := Rect{Point{0, 0}, Point{h, w}}
	window := NewWindowView(screen, roi, buffer, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	cursor, _ = cursor.RunesBackward(1)
	window.Update(roi, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertPointsEqual(t, window.text_offset, Point{col: 3, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("e1"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("e2"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("e3"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("e4"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("e5"), "")

	cursor, _ = cursor.RunesBackward(1)
	window.Update(roi, cursor, NewBufferCursor(buffer), NormalMode)
	window.Draw()

	screen.Show()
	assertPointsEqual(t, window.text_offset, Point{col: 2, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("ne"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("ne"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("ne"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("źe"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("ne"), "")
}

// CHARACTER CURSOR

func TestDrawCharacterCursor(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"ląćź2",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup cursor
	cursor := NewBufferCursor(buffer)

	// Setup window
	w, h := screen.Size()
	roi := Rect{Point{0, 0}, Point{h, w}}
	var cursorView View2
	cursorView = NewCharacterViewCursor2(screen, roi, buffer, cursor, Point{0, 0})
	cursorView.Draw()

	screen.Show()
	x, y, visible := screen.GetCursor()
	assertIntEqual(t, x, 0)
	assertIntEqual(t, y, 0)
	if !visible {
		t.Errorf("Expected cursor to be visible")
	}
}

func TestDrawCharacterCursorAfterMovement(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"line1",
		"line2",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()

	// Setup cursor
	window := windowFromBuffer(buffer, w, h)
	window.moveCursor(Right)
	window.moveCursor(Down)

	roi := Rect{Point{0, 0}, Point{h, w}}
	var cursorView View2
	cursorView = NewCharacterViewCursor2(screen, roi, buffer, window.cursor, Point{})
	cursorView.Draw()

	screen.Show()
	x, y, visible := screen.GetCursor()
	assertIntEqual(t, x, 1)
	assertIntEqual(t, y, 1)
	if !visible {
		t.Errorf("Expected cursor to be visible")
	}
}

func TestDrawCharacterCursorAfterMovementOnNonAscii(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"ąźne1",
		"ćłne2",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()

	// Setup cursor
	window := windowFromBuffer(buffer, w, h)
	window.moveCursor(Right)
	window.moveCursor(Down)

	roi := Rect{Point{0, 0}, Point{h, w}}
	var cursorView View2
	cursorView = NewCharacterViewCursor2(screen, roi, buffer, window.cursor, Point{0, 0})
	cursorView.Draw()

	screen.Show()
	x, y, visible := screen.GetCursor()
	assertIntEqual(t, x, 1)
	assertIntEqual(t, y, 1)
	if !visible {
		t.Errorf("Expected cursor to be visible")
	}
}

func TestDrawIndexCursorAfterMovementOnNonAscii(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	lines := []string{
		"ąźne1",
		"ćłne2",
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()

	// Setup cursor
	window := windowFromBuffer(buffer, w, h)
	window.moveCursor(Right)
	window.moveCursor(Down)

	roi := Rect{Point{0, 0}, Point{h, w}}
	var cursorView View2
	cursorView = &IndexViewCursor{screen, roi, buffer, window.cursor, Point{0, 0}}
	cursorView.Draw()

	screen.Show()
	x, y, visible := screen.GetCursor()
	assertIntEqual(t, x, 1)
	assertIntEqual(t, y, 1)
	if !visible {
		t.Errorf("Expected cursor to be visible")
	}
}

// TODO: Add test to selection cursor view

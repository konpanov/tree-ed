package main

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

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
	var view View
	view = NewTextView(screen, Rect{left: 0, top: 0, right: 10, bot: 10}, [][]rune{text})
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
	var view View
	view = NewTextView(screen, Rect{left: 0, top: 0, right: 10, bot: 10}, [][]rune{line1, line2})
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
	var view View
	view = NewTextView(screen, Rect{top: 3, left: 4, bot: 10, right: 10}, [][]rune{line1, line2})
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

	var view View
	view = NewTextView(screen, Rect{top: 0, left: 0, bot: 10, right: 10}, [][]rune{line1, line2, line3})
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
	w, h := screen.Size()
	window := windowFromBuffer(buffer)
	roi := Rect{top: 0, left: 0, bot: w, right: h}
	window_view := NewWindowView(screen, roi, window)
	window_view.Draw()

	screen.Show()
	cells, _, _ := screen.GetContents()
	assertCells(t, cells[2:7], []rune(content), "")
	assertCellsEmpty(t, cells[7:w*(h-2)])
}

func TestDrawWindowViewWithOverflowHeightLine(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 5)
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
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window := windowFromBuffer(buffer)
	window_view := NewWindowView(screen, roi, window)
	window_view.Draw()

	screen.Show()
	assertIntEqual(t, w, 10)
	assertIntEqual(t, h, 5)
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 "+lines[0]), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 "+lines[1]), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("3 "+lines[2]), "")
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
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window := windowFromBuffer(buffer)
	var window_view View
	window_view = NewWindowView(screen, roi, window)
	window_view.Draw()

	screen.Show()
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 "+lines[0]), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 "+lines[1]), "")
}

func TestDrawWindowViewWithVerticalTextOffset(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(8, 2)
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
	w, h := screen.Size()
	window := windowFromBuffer(buffer)
	window.setCursor(window.cursor.ToIndex(20), true)
	assertIntEqualMsg(t, w, 8, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, window.cursor.Index(), 20, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}

	// Setup window
	var window_view View
	window_view = NewWindowView(screen, roi, window)
	window_view.Draw()

	screen.Show()
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("3 line3"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("4 line4"), "")
}

func TestDrawWindowViewWithVerticalTextOffsetAndReturn(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(8, 2)
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
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window := windowFromBuffer(buffer)
	window.setCursor(window.cursor.ToIndex(20), true)
	window_view := NewWindowView(screen, roi, window)
	assertIntEqualMsg(t, w, 8, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, window.cursor.Index(), 20, "")
	window_view.Draw()

	window.setCursor(window.cursor.ToIndex(14), true)
	assertIntEqualMsg(t, window.cursor.Index(), 14, "")
	window_view.SetRoi(roi)
	window_view.Draw()
	screen.Show()

	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("3 line3"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("4 line4"), "")

	window.setCursor(window.cursor.ToIndex(8), true)
	assertIntEqualMsg(t, window.cursor.Index(), 8, "")
	window_view.SetRoi(roi)
	window_view.Draw()
	screen.Show()

	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("2 line2"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("3 line3"), "")
}

func TestDrawWindowViewWithHorizontalTextOffset(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(4, 8)
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

	// Setup window
	w, h := screen.Size()
	window := windowFromBuffer(buffer)
	for range 4 {
		window.setCursor(window.cursor.RuneNext(), true)
	}
	assertIntEqualMsg(t, w, 4, "")
	assertIntEqualMsg(t, h, 8, "")
	assertIntEqualMsg(t, window.cursor.Index(), 4, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := NewWindowView(screen, roi, window)
	window_view.Draw()

	screen.Show()
	assertPointsEqual(t, window_view.text_offset, Point{col: 3, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 e1"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 e2"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("3 e3"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("4 e4"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("5 e5"), "")
}

func TestDrawWindowViewWithHorizontalTextOffsetAndReturn(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(4, 8)
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

	// Setup window
	w, h := screen.Size()
	window := windowFromBuffer(buffer)
	for range 4 {
		window.setCursor(window.cursor.RuneNext(), true)
	}
	assertIntEqualMsg(t, w, 4, "")
	assertIntEqualMsg(t, h, 8, "")
	assertIntEqualMsg(t, window.cursor.Index(), 4, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := NewWindowView(screen, roi, window)
	window_view.Draw()

	window.setCursor(window.cursor.RunePrev(), true)
	window_view.SetRoi(roi)
	window_view.Draw()

	screen.Show()
	assertPointsEqual(t, window_view.text_offset, Point{col: 3, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 e1"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 e2"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("3 e3"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("4 e4"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("5 e5"), "")

	window.setCursor(window.cursor.RunePrev(), true)
	window_view.SetRoi(roi)
	window_view.Draw()

	screen.Show()
	assertPointsEqual(t, window_view.text_offset, Point{col: 2, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 ne"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 ne"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("3 ne"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("4 źe"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("5 ne"), "")
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
	cursor := BufferCursor{buffer: buffer, index: 0}

	// Setup window
	w, h := screen.Size()
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = NewCharacterViewCursor(screen, roi, buffer, cursor, Point{0, 0})
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
	window := windowFromBuffer(buffer)
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = NewCharacterViewCursor(screen, roi, buffer, window.cursor, Point{})
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
	window := windowFromBuffer(buffer)
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = NewCharacterViewCursor(screen, roi, buffer, window.cursor, Point{0, 0})
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
	window := windowFromBuffer(buffer)
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
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

func TestDrawSelectionCursorOnWholePage(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 5)
	defer screen.Fini()
	w, h := screen.Size()
	assertIntEqualMsg(t, w, 10, "Unexpected screen width: ")
	assertIntEqualMsg(t, h, 5, "Unexpected screen width: ")

	// Setup buffer
	nl := NewLineUnix
	lines := []string{}
	for i := 0; i < h+10; i++ {
		lines = append(lines, "line"+strconv.Itoa(i+1))
	}
	content := strings.Join(lines, nl)
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup cursor
	window := windowFromBuffer(buffer)
	window_view := NewWindowView(
		screen,
		Rect{left: 0, top: 0, right: w, bot: h},
		window,
	)

	for range 5 {
		window.cursorDown(1)
		window_view.SetRoi(Rect{left: 0, top: 0, right: w, bot: h})
		window_view.Draw()
	}
	window.switchToVisual()
	window_view.SetRoi(Rect{left: 0, top: 0, right: w, bot: h})
	window_view.Draw()

	for i := 5; i < len(lines)-5; i++ {
		window.cursorDown(1)
		window_view.SetRoi(Rect{left: 0, top: 0, right: w, bot: h})
		window_view.Draw()
	}

	screen.Show()
}

func TestDrawWindowEraseAtCursor(t *testing.T) {
	var err error
	content := []byte(strings.Join([]string{
		"package main",
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(NewLineUnix)))
	nl_seq := []byte(NewLineUnix)
	buffer, err := bufferFromContent(content, nl_seq, nil)
	assertNoErrors(t, err)
	window := windowFromBuffer(buffer)
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window_view := NewWindowView(
		screen,
		Rect{left: 0, top: 0, right: w, bot: h},
		window,
	)
	window_view.Draw()
	window.eraseLineAtCursor(1)
	window_view.Draw()
}

func TestDrawWindowInsertCursor(t *testing.T) {
	var err error
	content := []byte(strings.Join([]string{
		"package main",
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(NewLineUnix)))
	nl_seq := []byte(NewLineUnix)
	buffer, err := bufferFromContent(content, nl_seq, nil)
	assertNoErrors(t, err)
	window := windowFromBuffer(buffer)
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window_view := NewWindowView(
		screen,
		Rect{left: 0, top: 0, right: w, bot: h},
		window,
	)
	window_view.Draw()
	x, y, visible := screen.GetCursor()
	if !visible || x != 2 || y != 0 {
		t.Errorf(
			"Cursor in an unexpected state: %+v, %+v, %+v",
			visible,
			x,
			y,
		)
	}
	window.switchToInsert()
	x, y, visible = screen.GetCursor()
	if !visible || x != 2 || y != 0 {
		t.Errorf(
			"Cursor in an unexpected state: %+v, %+v, %+v",
			visible,
			x,
			y,
		)
	}
}

func TestDrawWindowInsertCursorOnEmptyContent(t *testing.T) {
	var err error
	content := []byte{}
	nl_seq := []byte(NewLineUnix)
	buffer, err := bufferFromContent(content, nl_seq, nil)
	assertNoErrors(t, err)
	window := windowFromBuffer(buffer)
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window_view := NewWindowView(
		screen,
		Rect{left: 0, top: 0, right: w, bot: h},
		window,
	)
	window.switchToInsert()
	window.insertContent(false, []byte("a"))
	window_view.Draw()
	x, y, visible := screen.GetCursor()
	if !visible || x != 3 || y != 0 {
		t.Errorf(
			"Cursor in an unexpected state: %+v, %+v, %+v",
			visible,
			x,
			y,
		)
	}
}

func TestDrawWindowAppendMode(t *testing.T) {
	var err error
	content := []byte("a")
	nl_seq := []byte(NewLineUnix)
	buffer, err := bufferFromContent(content, nl_seq, nil)
	assertNoErrors(t, err)
	window := windowFromBuffer(buffer)
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	// w, h := screen.Size()
	// window_view := NewWindowView(
	// 	screen,
	// 	Rect{left: 0, top: 0, right: w, bot: h},
	// 	window,
	// )
	window.switchToInsert()
	window.cursorRight(1)
	if window.cursor.index != 1 {
		t.Errorf("Unexpected cursor index %+v", window.cursor.index)
	}
	// window_view.Draw()
	// x, y, visible := screen.GetCursor()
	// if !visible || x != 3 || y != 0 {
	// 	t.Errorf(
	// 		"Cursor in an unexpected state: %+v, %+v, %+v",
	// 		visible,
	// 		x,
	// 		y,
	// 	)
	// }
}

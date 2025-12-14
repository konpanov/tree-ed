package main

import (
	"fmt"
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
	w, h := 10, 4
	screen.SetSize(w, h)
	defer screen.Fini()

	nl := NewLineUnix
	content := []byte(strings.Join([]string{
		"hello",
	}, nl))
	buf, err := bufferFromContent(content, []byte(nl), nil)
	panic_if_error(err)
	win := windowFromBuffer(buf, w, h)
	ctx := DrawContext{screen: screen, roi: Rect{top: 0, left: 0, right: w, bot: h}, theme: default_theme}
	WindowView{window: win}.DrawFrameText(ctx)

	screen.Show()
	assertScreenRunes(t, screen, []string{
		"hello     ",
		"          ",
		"          ",
		"          ",
	})
}

func TestDrawDoubleLineTextView(t *testing.T) {
	screen := mkTestScreen(t, "")
	w, h := 10, 4
	screen.SetSize(w, h)
	defer screen.Fini()

	nl := NewLineUnix
	content := []byte(strings.Join([]string{
		"hello",
		"world",
	}, nl))
	buf, err := bufferFromContent(content, []byte(nl), nil)
	panic_if_error(err)
	win := windowFromBuffer(buf, w, h)
	ctx := DrawContext{screen: screen, roi: Rect{top: 0, left: 0, right: w, bot: h}, theme: default_theme}
	WindowView{window: win}.DrawFrameText(ctx)

	screen.Show()
	assertScreenRunes(t, screen, []string{
		"hello     ",
		"world     ",
		"          ",
		"          ",
	})
}

func TestDrawDoubleLineTextViewWithOffsetRoi(t *testing.T) {
	screen := mkTestScreen(t, "")
	w, h := 10, 4
	screen.SetSize(w, h)
	defer screen.Fini()

	nl := NewLineUnix
	content := []byte(strings.Join([]string{
		"hello",
		"world",
	}, nl))
	buf, err := bufferFromContent(content, []byte(nl), nil)
	panic_if_error(err)
	win := windowFromBuffer(buf, w, h)
	ctx := DrawContext{screen: screen, roi: Rect{top: 1, left: 2, right: w, bot: h}, theme: default_theme}
	WindowView{window: win}.DrawFrameText(ctx)

	screen.Show()
	assertScreenRunes(t, screen, []string{
		"          ",
		"  hello   ",
		"  world   ",
		"          ",
	})
}

func TestDrawSkippedLineTextView(t *testing.T) {
	screen := mkTestScreen(t, "")
	w, h := 10, 4
	screen.SetSize(w, h)
	defer screen.Fini()

	nl := NewLineUnix
	content := []byte(strings.Join([]string{
		"hello",
		"",
		"world",
	}, nl))
	buf, err := bufferFromContent(content, []byte(nl), nil)
	panic_if_error(err)
	win := windowFromBuffer(buf, w, h)
	ctx := DrawContext{screen: screen, roi: Rect{top: 0, left: 0, right: w, bot: h}, theme: default_theme}
	WindowView{window: win}.DrawFrameText(ctx)

	screen.Show()
	assertScreenRunes(t, screen, []string{
		"hello     ",
		"          ",
		"world     ",
		"          ",
	})
}

// WINDOW VIEW

func TestDrawWindowViewWithSingleLine(t *testing.T) {
	// Setup screen
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	defer screen.Fini()

	// Setup buffer
	nl := NewLineUnix
	content := "hello"
	buffer := mkTestBuffer(t, content+nl, nl)

	// Setup window
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	screen.Show()
	cells, _, _ := screen.GetContents()
	assertScreenRunes(t, screen, []string{
		"1 hello   ",
		"          ",
		"          ",
		"          ",
	})
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
	window := windowFromBuffer(buffer, w, h)
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window := windowFromBuffer(buffer, w, h)
	var window_view View
	window_view = WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window := windowFromBuffer(buffer, w, h)
	window.setCursor(window.cursor.ToIndex(20), true)
	assertIntEqualMsg(t, w, 8, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, window.cursor.Index(), 20, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}

	// Setup window
	var window_view View
	window_view = WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	screen.Show()
	actual := window.cursor.RunePosition()
	expected := Point{row: 3, col: 2}
	if actual != expected {
		t.Errorf("Unexpected cursor position %+v, expected %+v.", actual, expected)
	}
	assertScreenRunes(t, screen, []string{
		"3 line3 ",
		"4 line4 ",
	})
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
	window := windowFromBuffer(buffer, w, h)
	window.setCursor(window.cursor.ToIndex(20), true)
	window_view := WindowView{window: window}
	assertIntEqualMsg(t, w, 8, "")
	assertIntEqualMsg(t, h, 2, "")
	assertIntEqualMsg(t, window.cursor.Index(), 20, "")
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	window.setCursor(window.cursor.ToIndex(14), true)
	assertIntEqualMsg(t, window.cursor.Index(), 14, "")
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
	screen.Show()

	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("3 line3"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("4 line4"), "")

	window.setCursor(window.cursor.ToIndex(8), true)
	assertIntEqualMsg(t, window.cursor.Index(), 8, "")
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
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
	window := windowFromBuffer(buffer, w, h)
	for range 4 {
		window.setCursor(window.cursor.RuneNext(), true)
	}
	assertIntEqualMsg(t, w, 4, "")
	assertIntEqualMsg(t, h, 8, "")
	assertIntEqualMsg(t, window.cursor.Index(), 4, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	screen.Show()
	assertPointsEqual(t, window.frame.TopLeft(), Point{col: 3, row: 0})
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
	window := windowFromBuffer(buffer, w, h)
	for range 4 {
		window.setCursor(window.cursor.RuneNext(), true)
	}
	assertIntEqualMsg(t, w, 4, "")
	assertIntEqualMsg(t, h, 8, "")
	assertIntEqualMsg(t, window.cursor.Index(), 4, "")
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	window.setCursor(window.cursor.RunePrev(), true)
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	screen.Show()
	assertPointsEqual(t, window_view.window.frame.TopLeft(), Point{col: 3, row: 0})
	assertScreenText(t, screen, Point{row: 0, col: 0}, []rune("1 e1"), "")
	assertScreenText(t, screen, Point{row: 1, col: 0}, []rune("2 e2"), "")
	assertScreenText(t, screen, Point{row: 2, col: 0}, []rune("3 e3"), "")
	assertScreenText(t, screen, Point{row: 3, col: 0}, []rune("4 e4"), "")
	assertScreenText(t, screen, Point{row: 4, col: 0}, []rune("5 e5"), "")

	window.setCursor(window.cursor.RunePrev(), true)
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	screen.Show()
	assertPointsEqual(t, window_view.window.frame.TopLeft(), Point{col: 2, row: 0})
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
	window := windowFromBuffer(buffer, w, h)
	window.cursor = cursor

	var cursorView View
	cursorView = &CharacterViewCursor{window: window}
	cursorView.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = &CharacterViewCursor{window: window}
	cursorView.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = &CharacterViewCursor{window: window}
	cursorView.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window.cursorRight(1)
	window.cursorDown(1)

	roi := Rect{top: 0, left: 0, bot: h, right: w}
	var cursorView View
	cursorView = &IndexViewCursor{window: window}
	cursorView.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

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
	window := windowFromBuffer(buffer, w, h)
	window_view := WindowView{window: window}

	for range 5 {
		window.cursorDown(1)
		roi := Rect{left: 0, top: 0, right: w, bot: h}
		window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
	}
	window.switchToVisual()
	roi := Rect{left: 0, top: 0, right: w, bot: h}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	for i := 5; i < len(lines)-5; i++ {
		window.cursorDown(1)
		window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
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
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	roi := Rect{left: 0, top: 0, right: w, bot: h}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
	window.eraseLineAtCursor(1)
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
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
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	roi := Rect{left: 0, top: 0, right: w, bot: h}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
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
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	roi := Rect{left: 0, top: 0, right: w, bot: h}
	window.switchToInsert()
	window.insertContent(false, []byte("a"))
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
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
	screen := mkTestScreen(t, "")
	defer screen.Fini()
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	window.switchToInsert()
	window.cursorRight(1)
	if window.cursor.index != 1 {
		t.Errorf("Unexpected cursor index %+v", window.cursor.index)
	}
}

func TestMoveBelowFrameAndUp(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 3)
	defer screen.Fini()

	nl := NewLineUnix
	lines := []string{}
	for i := range 100 {
		lines = append(lines, fmt.Sprintf("line%d", i+1))
	}
	content := as_content(lines, nl)
	buffer := mkTestBuffer(t, string(content), nl)

	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)

	screen.Clear()
	roi := Rect{top: 0, left: 0, bot: h, right: w}
	window_view := WindowView{window: window}
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})

	expected := Rect{top: 0, left: 0, bot: 3, right: 6}
	if window.frame != expected {
		t.Errorf("Unexpected window frame %+v, expected %+v.", window.frame, expected)
	}
	assertScreenRunes(t, screen, []string{
		"1   line1 ",
		"2   line2 ",
		"3   line3 ",
	})

	window.setCursor(window.cursor.MoveToRunePos(Point{row: 24, col: 2}), true)
	screen.Clear()
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
	assertScreenRunes(t, screen, []string{
		"23  line23",
		"24  line24",
		"25  line25",
	})

	window.setCursor(window.cursor.MoveToRunePos(Point{row: 8, col: 3}), true)
	screen.Clear()
	window_view.Draw(DrawContext{screen: screen, roi: roi, theme: default_theme})
	assertScreenRunes(t, screen, []string{
		"9   line9 ",
		"10  line10",
		"11  line11",
	})
}

func TestDrawPreview(t *testing.T) {
	editor := mkTestEditor(t, Point{col: 80, row: 20})
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                       mm       ",
		"       @@@**@@**@@@                                                  *@@@       ",
		"       @*   @@   *@                                                    @@       ",
		"            @@     *@@@m@@@   mm@*@@   mm@*@@           mm@*@@    m@**@@@       ",
		"            !@       @@* **  m@*   @@ m@*   @@         m@*   @@ m@@    @@       ",
		"            !@       @!      !@****** !@******  @@@@@  !@****** @!@    @!       ",
		"            !@       @!      !@m    m !@m    m         !@m    m *!@    @!       ",
		"            !@       !!      !!****** !!******         !!****** !!!    !!       ",
		"            !!       !:      :!!      :!!              :!!      *:!    !:       ",
		"          : :!::   : :::      : : ::   : : ::           : : ::   : : : ! :      ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"input:                                                                          ",
	})
}

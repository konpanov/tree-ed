package main

import (
	"log"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type View2 interface {
	Draw()
	SetRoi(roi Rect)
	GetRoi() Rect
}

type View interface {
	Draw(screen tcell.Screen, roi Rect)
}

type BufferView struct {
	roi           Rect
	buffer        *Buffer
	cursor        WindowCursor
	number_column LineNumberColumn
	text          *TextView
	status_line   *StatusLine
	debug_view    DebugView
}

func (view *BufferView) Draw(screen tcell.Screen) {
	roi := view.roi
	line_number_column_width := view.number_column.GetWidth(view)
	status_line_height := view.status_line.GetHeight()
	debug_width := 20

	sl_roi := view.roi.SetTop(roi.Bot() - status_line_height)
	column_roi := roi.SetRight(line_number_column_width).SetBot(sl_roi.Top())
	debug_roi := roi.SetLeft(roi.Right() - debug_width).SetBot(sl_roi.Top())
	text_roi := roi.SetLeft(column_roi.Right()).SetRight(debug_roi.Left()).SetBot(sl_roi.Top())

	view.status_line.Draw(screen, sl_roi)
	view.number_column.Draw(screen, view, column_roi)
	view.text.Draw(screen, text_roi)
	view.debug_view.Draw(screen, debug_roi)
}

/******************************************************/
/*                   Debug view                       */
/******************************************************/

type DebugView struct {
	buffer *Buffer
	window *Window
}

func (self DebugView) Draw(screen tcell.Screen, roi Rect) {
	for row := roi.Top(); row < roi.Bot(); row++ {
		set_rune(screen, Point{row: row, col: roi.Left()}, '|')
	}
	set_line(screen, roi, Point{col: 1, row: 0}, "Debug view")
	set_line(screen, roi, Point{col: 1, row: 1}, "Buffer length:")
	set_line(screen, roi, Point{col: 1, row: 2}, strconv.Itoa(len(self.buffer.content)))
}

/******************************************************/
/*                   Text view                        */
/******************************************************/

type TextView struct {
	buffer      *Buffer
	style       tcell.Style
	text_offset Point
	shifter     ViewTextShifter
	cursor      ViewCursor
	text        [][]rune
}

func (v *TextView) Draw(screen tcell.Screen, roi Rect) {
	v.shifter.Shift(v, roi)
	v.text = get_text_to_draw(roi, v.text_offset, v.buffer)
	// log.Println(string(v.buffer.content[to_draw[2].start:to_draw[2].end]))

	v.draw_background(screen, roi)
	// TODO Make cursor draw use slice of slices of runes or bring back slice of region or other
	v.cursor.Draw(screen, roi, v)
	draw_text(screen, roi, v.text)

}

func view_pos_to_text_pos(pos Point, text_offset Point) Point {
	return Point{
		col: pos.col + text_offset.col,
		row: pos.row + text_offset.row,
	}
}

func view_pos_to_screen_pos(pos Point, roi Rect) Point {
	col := pos.col + roi.top_left.col
	row := pos.row + roi.top_left.row
	screen_pos := Point{col: col, row: row}
	if row >= roi.bot_right.row || row < roi.top_left.row || col >= roi.bot_right.col || col < roi.top_left.col {
		log.Panicf("View should not draw outside it's roi. View roi: %+v. Screen position: %+v. View position: %+v.\n", roi, screen_pos, pos)
	}
	return screen_pos
}

func text_pos_to_view_pos(pos Point, text_offset Point, roi Rect) Point {
	width := roi.bot_right.col - roi.top_left.col
	height := roi.bot_right.row - roi.top_left.row
	view_pos := Point{
		col: min(pos.col-text_offset.col, width-1),
		row: min(pos.row-text_offset.row, height-1),
	}
	if view_pos.row < 0 || view_pos.col < 0 {
		log.Panicf("View position should not be negative. View pos: %+v, Text offset: %+v, Text pos: %+v", view_pos, text_offset, pos)
	}
	return view_pos
}

func (v TextView) view_pos_to_text_pos(pos Point) Point {
	return view_pos_to_text_pos(pos, v.text_offset)
}

func (v TextView) text_pos_to_view_pos(pos Point, roi Rect) Point {
	return text_pos_to_view_pos(pos, v.text_offset, roi)
}

func (v *TextView) draw_background(screen tcell.Screen, roi Rect) {
	for row := roi.top_left.row; row < roi.bot_right.row; row++ {
		for col := roi.top_left.col; col < roi.bot_right.col; col++ {
			set_style(screen, Point{col: col, row: row}, v.style)
		}
	}
}

func draw_text(screen tcell.Screen, roi Rect, to_draw [][]rune) {
	log.Println("Drawing text to sreen.")
	if len(to_draw) > roi.Height() {
		log.Panicf("Lines in text to draw do not fit in the given roi. Roi: %s, to_draw length: %d\n", roi.StringInfo(), len(to_draw))
	}
	width := roi.Width()
	for row, line := range to_draw {
		if len(line) > width {
			log.Panicf("A line in text to draw does not fit in the given roi. Roi: %s, line id: %d, line length: %d\n", roi.StringInfo(), row, len(line))
		}

		for col, value := range line {
			screen_pos := view_pos_to_screen_pos(Point{col: col, row: row}, roi)
			set_rune(screen, screen_pos, value)
		}
	}
}

func get_text_to_draw(roi Rect, text_offset Point, buffer *Buffer) [][]rune {
	to_draw := []Region{}
	lines := buffer.Lines()
	view_width := roi.Width()
	view_height := roi.Height()

	view_start_pos := Point{0, 0}
	view_end_pos := Point{row: min(view_height, len(lines)), col: view_width}

	text_start_pos := view_pos_to_text_pos(view_start_pos, text_offset)
	text_end_pos := view_pos_to_text_pos(view_end_pos, text_offset)

	for y, line := range lines[text_start_pos.row:text_end_pos.row] {
		view_line_start_pos := Point{row: y, col: 0}
		view_line_end_pos := Point{row: y, col: view_width}

		text_line_start_pos := view_pos_to_text_pos(view_line_start_pos, text_offset)
		text_line_end_pos := view_pos_to_text_pos(view_line_end_pos, text_offset)

		buffer_line_start := min(text_line_start_pos.col+line.start, line.end)
		buffer_line_end := min(text_line_end_pos.col+line.start, line.end)

		to_draw = append(to_draw, Region{buffer_line_start, buffer_line_end})
	}

	text_to_draw := [][]rune{}
	for _, region := range to_draw {
		text_to_draw = append(text_to_draw, []rune(string(buffer.content[region.start:region.end])))
	}

	return text_to_draw
}

/******************************************************/
/*                   Status line                      */
/******************************************************/

type StatusLine struct {
	filename string
	cursor   WindowCursor
	buffer   *Buffer
	mode     string
}

func (sl *StatusLine) GetHeight() int {
	return 2
}

func (sl *StatusLine) Draw(screen tcell.Screen, roi Rect) {
	log.Println("Drawing status line")
	info := ""
	info += "file: " + sl.filename
	info += ", "
	info += "line: " + strconv.Itoa(sl.cursor.row)
	info += ", "
	info += "byte: " + strconv.Itoa(sl.cursor.col)
	info += ", "
	info += "mode: " + sl.mode

	for col := roi.Left(); col < roi.Right(); col++ {
		set_rune(screen, Point{col: col, row: roi.Top()}, '-')
	}

	for col, value := range info {
		pos := view_pos_to_screen_pos(Point{row: 1, col: col}, roi)
		set_rune(screen, pos, value)
	}
}

/******************************************************/
/*                   Line number column               */
/******************************************************/
type LineNumberColumn interface {
	GetWidth(view *BufferView) int
	Draw(screen tcell.Screen, view *BufferView, roi Rect)
}

func default_buffer_line_number_max_width(buffer IBuffer) int {
	return int(math.Log10(float64(len(buffer.Lines())))) + 2
}

type RelativeNumberColumnView struct{}

func (nc RelativeNumberColumnView) GetWidth(view *BufferView) int {
	return default_buffer_line_number_max_width(view.buffer)
}

func (nc RelativeNumberColumnView) Draw(screen tcell.Screen, view *BufferView, roi Rect) {
	lines := view.buffer.Lines()
	width := roi.Width()
	height := roi.Height()
	start_line := view.text.text_offset.row

	for y := range lines[start_line : start_line+height] {
		relative := start_line + y - view.cursor.row
		if relative < 0 {
			relative = -relative
		}
		if relative == 0 {
			relative = start_line + y + 1
		}
		line_num := strconv.Itoa(relative)
		for i, r := range line_num {
			screen_pos := view_pos_to_screen_pos(
				Point{col: width - 1 - len(line_num) + i, row: y},
				roi,
			)
			set_rune(screen, screen_pos, r)
		}
		set_rune(screen, view_pos_to_screen_pos(Point{row: y, col: width - 1}, roi), '|')
	}
}

type AbsoluteNumberColumnView struct {
}

func (nc AbsoluteNumberColumnView) GetWidth(view *BufferView) int {
	return default_buffer_line_number_max_width(view.buffer)
}

func (nc AbsoluteNumberColumnView) Draw(screen tcell.Screen, view *BufferView, roi Rect) {
	lines := view.buffer.Lines()
	width := roi.Width()
	height := roi.Height()
	start_line := view.text.text_offset.row

	for y := range lines[start_line : start_line+height] {
		line_num := strconv.Itoa(start_line + y + 1)
		for i, r := range line_num {
			screen_pos := view_pos_to_screen_pos(
				Point{col: width - 1 - len(line_num) + i, row: y},
				roi,
			)
			set_rune(screen, screen_pos, r)
		}
	}
}

/******************************************************/
/*                   View shifters                    */
/******************************************************/

type ViewTextShifter interface {
	Shift(view *TextView, roi Rect)
}

type CursorViewShifter struct {
	cursors []*WindowCursor
}

func (shifter CursorViewShifter) Shift(view *TextView, roi Rect) {
	view_width := roi.bot_right.col - roi.top_left.col
	view_height := roi.bot_right.row - roi.top_left.row
	for _, cursor := range shifter.cursors {
		view.text_offset = Point{
			col: max(min(view.text_offset.col, cursor.col), cursor.col-view_width+1),
			row: max(min(view.text_offset.row, cursor.row), cursor.row-view_height+1),
		}
	}
}

/******************************************************/
/*                     Cursors                        */
/******************************************************/

type ViewCursor interface {
	Draw(screen tcell.Screen, roi Rect, view *TextView)
}

type CharacterViewCursor struct {
	position_in_buffer int
}

type BetweenCharactersViewCursor struct {
	position_in_buffer int
}

type SelectoionViewCursor struct {
	selection Region
	style     tcell.Style
}

func (cursor CharacterViewCursor) Draw(screen tcell.Screen, roi Rect, view *TextView) {
	screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	coord, err := view.buffer.RuneCoord(cursor.position_in_buffer)
	if err != nil {
		log.Fatalln("Could not find cursor index: ", err)
	}
	view_pos := view.text_pos_to_view_pos(coord, roi)
	screen_pos := view_pos_to_screen_pos(view_pos, roi)
	screen.ShowCursor(screen_pos.col, screen_pos.row)
}

func (cursor SelectoionViewCursor) Draw(screen tcell.Screen, roi Rect, view *TextView) {
	screen.SetCursorStyle(tcell.CursorStyleDefault)
	screen.ShowCursor(-1, -1)

	start := cursor.selection.Start()
	start_coord, _ := view.buffer.RuneCoord(start)
	start_view_coord := view.text_pos_to_view_pos(start_coord, roi)

	end := cursor.selection.End()
	end_coord, _ := view.buffer.RuneCoord(end)
	end_view_coord := view.text_pos_to_view_pos(end_coord, roi)

	for row, line := range view.text {
		line_selection := Region{0, len(line) - 1}
		if row < start_view_coord.row || row > end_view_coord.row {
			continue
		}
		if row == start_view_coord.row {
			line_selection.start = start_view_coord.col
		}
		if row == end_view_coord.row {
			line_selection.end = end_view_coord.col
		}
		for col := line_selection.start; col <= line_selection.end; col++ {
			view_pos := Point{row: row, col: col}
			screen_pos := view_pos_to_screen_pos(view_pos, roi)
			set_style(screen, screen_pos, cursor.style)
		}

	}
}

func (cursor BetweenCharactersViewCursor) Draw(screen tcell.Screen, view *TextView, to_draw []Region, roi Rect) {
	screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
	coord, err := view.buffer.Coord(cursor.position_in_buffer)
	if err == nil {
		pos := view.text_pos_to_view_pos(coord, roi)
		screen.ShowCursor(roi.top_left.col+pos.col, roi.top_left.row+pos.row)
	}
}

/******************************************************/
/*                     Utils                          */
/******************************************************/

func set_line(screen tcell.Screen, roi Rect, view_pos Point, text string) {
	for col, value := range text {
		pos := view_pos_to_screen_pos(Point{row: view_pos.row, col: view_pos.col + col}, roi)
		set_rune(screen, pos, value)
	}
}

func set_rune(screen tcell.Screen, pos Point, value rune) {
	_, _, style, _ := screen.GetContent(pos.col, pos.row)
	if value == '\r' || value == '\n' {
		value = ' '
	}
	screen.SetContent(pos.col, pos.row, value, nil, style)
}

func set_style(screen tcell.Screen, pos Point, style tcell.Style) {
	value, _, _, _ := screen.GetContent(pos.col, pos.row)
	screen.SetContent(pos.col, pos.row, value, nil, style)
}

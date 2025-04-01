package main

import (
	"log"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type View interface {
	Draw(screen tcell.Screen, roi Rect)
}

type StatusLine struct {
	filename string
	cursor   WindowCursor
	buffer   *Buffer
}

func (sl *StatusLine) Draw(screen tcell.Screen, roi Rect) {
	info := ""
	info += "file: " + sl.filename
	info += ", "
	info += "line: " + strconv.Itoa(sl.cursor.row)
	info += ", "
	info += "column: " + strconv.Itoa(sl.cursor.col)

	for col, value := range info {
		pos := view_pos_to_screen_pos(Point{row: 0, col: col}, roi)
		set_rune(screen, pos, byte(value))
	}
}

type NumberColumnView struct {
	buffer *Buffer
}

func (nc *NumberColumnView) Draw(screen tcell.Screen, roi Rect, row_offset int) {
	lines := nc.buffer.Lines()
	width := roi.bot_right.col - roi.top_left.col
	height := roi.bot_right.row - roi.top_left.row

	for y := range lines[row_offset : row_offset+height] {
		line_num := strconv.Itoa(row_offset + y + 1)
		for i, r := range line_num {
			screen_pos := view_pos_to_screen_pos(
				Point{col: width - 1 - len(line_num) + i, row: y},
				roi,
			)
			set_rune(screen, screen_pos, byte(r))
		}
	}
}

type BufferView struct {
	buffer        *Buffer
	number_column *NumberColumnView
	text          *TextView
	status_line   *StatusLine
}

func (view *BufferView) Draw(screen tcell.Screen, roi Rect) {
	numer_column_width := int(max(math.Log10(float64(len(view.buffer.Lines()))))) + 2
	status_line_height := 0

	if view.status_line != nil {
		status_line_height = 1
		sl_roi := roi
		sl_roi.top_left.row = roi.bot_right.row - 1
		view.status_line.Draw(screen, sl_roi)
	}

	text_roi := roi
	text_roi.top_left.col += numer_column_width
	text_roi.bot_right.row -= status_line_height
	view.text.Draw(screen, text_roi)

	column_roi := roi
	column_roi.bot_right.col = numer_column_width
	column_roi.bot_right.row -= status_line_height
	view.number_column.Draw(screen, column_roi, view.text.text_offset.row)
}

func selection_text_view(view *TextView, cursor *WindowCursor, anchorCursor *WindowCursor) {
	view.cursor_drawing = SelectoionViewCursor{
		selection: Range{cursor.index, anchorCursor.index},
		style:     tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorDarkGray),
	}
	view.view_shifter = CursorViewShifter{[]*WindowCursor{anchorCursor, cursor}}
}

func normal_text_view(view *TextView, cursor *WindowCursor) {
	view.cursor_drawing = CharacterViewCursor{position_in_buffer: cursor.index}
	view.view_shifter = CursorViewShifter{[]*WindowCursor{cursor}}
}

type TextView struct {
	buffer         *Buffer
	style          tcell.Style
	text_offset    Point
	view_shifter   ViewTextShifter
	cursor_drawing ViewCursor
}

func (v *TextView) Draw(screen tcell.Screen, roi Rect) {
	v.view_shifter.Shift(v, roi)
	to_draw := get_text_to_draw(roi, v.text_offset, v.buffer)

	v.draw_background(screen, roi)
	v.cursor_drawing.Draw(screen, v, to_draw, roi)
	v.draw_text(screen, to_draw, roi)

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
	if row >= roi.bot_right.row || row < roi.top_left.row || col >= roi.bot_right.col || col < roi.top_left.col {
		log.Fatalln("View should not draw outside it's roi")
	}
	return Point{col: col, row: row}
}

func text_pos_to_view_pos(pos Point, text_offset Point, roi Rect) Point {
	width := roi.bot_right.col - roi.top_left.col
	height := roi.bot_right.row - roi.top_left.row
	return Point{
		col: min(pos.col-text_offset.col, width-1),
		row: min(pos.row-text_offset.row, height-1),
	}
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

func (v *TextView) draw_text(screen tcell.Screen, to_draw []Range, roi Rect) {
	for row, line_range := range to_draw {
		for col, value := range v.buffer.content[line_range.start:line_range.end] {
			screen_pos := view_pos_to_screen_pos(Point{col: col, row: row}, roi)
			set_rune(screen, screen_pos, value)
		}
	}
}

type Rect struct {
	top_left, bot_right Point
}

func get_text_to_draw(roi Rect, text_offset Point, buffer *Buffer) []Range {
	to_draw := []Range{}
	lines := buffer.Lines()
	view_width := roi.bot_right.col - roi.top_left.col
	view_height := roi.bot_right.row - roi.top_left.row

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

		to_draw = append(to_draw, Range{buffer_line_start, buffer_line_end})
	}
	return to_draw
}

func set_rune(screen tcell.Screen, pos Point, value byte) {
	_, _, style, _ := screen.GetContent(pos.col, pos.row)
	if value == '\r' || value == '\n' {
		value = ' '
	}
	screen.SetContent(pos.col, pos.row, rune(value), nil, style)
}

func set_style(screen tcell.Screen, pos Point, style tcell.Style) {
	value, _, _, _ := screen.GetContent(pos.col, pos.row)
	screen.SetContent(pos.col, pos.row, value, nil, style)
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
	Draw(screen tcell.Screen, view *TextView, to_draw []Range, roi Rect)
}

type CharacterViewCursor struct {
	position_in_buffer int
}

func (cursor CharacterViewCursor) Draw(screen tcell.Screen, view *TextView, to_draw []Range, roi Rect) {
	log.Println("Drawing character cursor")
	screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	coord, err := view.buffer.Coord(cursor.position_in_buffer)
	if err != nil {
		log.Fatalln("Could not find cursor index: ", err)
	}
	pos := view.text_pos_to_view_pos(coord, roi)
	screen.ShowCursor(roi.top_left.col+pos.col, roi.top_left.row+pos.row)
}

type SelectoionViewCursor struct {
	selection Range
	style     tcell.Style
}

func (cursor SelectoionViewCursor) Draw(screen tcell.Screen, view *TextView, to_draw []Range, roi Rect) {
	screen.SetCursorStyle(tcell.CursorStyleDefault)
	screen.ShowCursor(-1, -1)
	log.Println("Drawing selection cursor")

	start_text_index := min(cursor.selection.start, cursor.selection.end)
	start_text_pos, err := view.buffer.Coord(start_text_index)
	if err != nil {
		log.Fatalln("Could not find start selection index: ", err)
	}
	start_view_pos := view.text_pos_to_view_pos(start_text_pos, roi)

	end_text_index := max(cursor.selection.start, cursor.selection.end)
	end_text_pos, err := view.buffer.Coord(end_text_index)
	if err != nil {
		log.Fatalln("Could not find end index: ", err)
	}
	end_view_pos := view.text_pos_to_view_pos(end_text_pos, roi)

	for row := start_view_pos.row; row <= end_view_pos.row; row++ {
		start_col := 0
		end_col := to_draw[row].end - to_draw[row].start
		if row == start_view_pos.row {
			start_col = start_text_pos.col
		}
		if row == end_view_pos.row {
			end_col = end_text_pos.col + 1
		}
		for i := start_col; i < end_col; i++ {
			set_style(screen, view_pos_to_screen_pos(Point{col: i, row: row}, roi), cursor.style)
		}
	}
}

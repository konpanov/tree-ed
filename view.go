package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

type View interface {
	Draw(screen tcell.Screen, start Point, end Point)
}

func selection_text_view(view TextView, cursor *WindowCursor, anchorCursor *WindowCursor) TextView {
	view.cursor_drawing = SelectoionViewCursor{
		selection: Range{cursor.index, anchorCursor.index},
		style:     tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorDarkGray),
	}
	view.view_shifter = CursorViewShifter{[]*WindowCursor{anchorCursor, cursor}}
	return view
}

func normal_text_view(
	view TextView, cursor *WindowCursor,
) TextView {
	view.cursor_drawing = CharacterViewCursor{position_in_buffer: cursor.index}
	view.view_shifter = CursorViewShifter{[]*WindowCursor{cursor}}
	return view
}

type TextView struct {
	buffer         *Buffer
	roi            Rect
	style          tcell.Style
	text_offset    Point
	view_shifter   ViewTextShifter
	cursor_drawing ViewCursor
}

func (v *TextView) Draw(screen tcell.Screen) {
	v.view_shifter.Shift(v)
	to_draw := get_text_to_draw(v.roi, v.text_offset, v.buffer)

	v.draw_background(screen)
	v.cursor_drawing.Draw(screen, v, to_draw)
	v.draw_text(screen, to_draw)

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

func (v TextView) text_pos_to_view_pos(pos Point) Point {
	return text_pos_to_view_pos(pos, v.text_offset, v.roi)
}

func (v *TextView) draw_background(screen tcell.Screen) {
	for row := v.roi.top_left.row; row < v.roi.bot_right.row; row++ {
		for col := v.roi.top_left.col; col < v.roi.bot_right.col; col++ {
			set_style(screen, Point{col: col, row: row}, v.style)
		}
	}
}

func (v *TextView) draw_text(screen tcell.Screen, to_draw []Range) {
	for row, line_range := range to_draw {
		for col, value := range v.buffer.content[line_range.start:line_range.end] {
			screen_pos := view_pos_to_screen_pos(Point{col: col, row: row}, v.roi)
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
	log.Printf("Lines: %+v\n", lines)
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
	Shift(view *TextView)
}

type CursorViewShifter struct {
	cursors []*WindowCursor
}

func (shifter CursorViewShifter) Shift(view *TextView) {
	view_width := view.roi.bot_right.col - view.roi.top_left.col
	view_height := view.roi.bot_right.row - view.roi.top_left.row
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
	Draw(screen tcell.Screen, view *TextView, to_draw []Range)
}

type CharacterViewCursor struct {
	position_in_buffer int
}

func (cursor CharacterViewCursor) Draw(screen tcell.Screen, view *TextView, to_draw []Range) {
	log.Println("Drawing character cursor")
	screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	coord, err := view.buffer.Coord(cursor.position_in_buffer)
	if err != nil {
		log.Fatalln("Could not find cursor index: ", err)
	}
	pos := view.text_pos_to_view_pos(coord)
	screen.ShowCursor(view.roi.top_left.col+pos.col, view.roi.top_left.row+pos.row)
}

type SelectoionViewCursor struct {
	selection Range
	style     tcell.Style
}

func (cursor SelectoionViewCursor) Draw(screen tcell.Screen, view *TextView, to_draw []Range) {
	screen.SetCursorStyle(tcell.CursorStyleDefault)
	screen.ShowCursor(-1, -1)
	log.Println("Drawing selection cursor")

	start_text_index := min(cursor.selection.start, cursor.selection.end)
	start_text_pos, err := view.buffer.Coord(start_text_index)
	if err != nil {
		log.Fatalln("Could not find start selection index: ", err)
	}
	start_view_pos := view.text_pos_to_view_pos(start_text_pos)

	end_text_index := max(cursor.selection.start, cursor.selection.end)
	end_text_pos, err := view.buffer.Coord(end_text_index)
	if err != nil {
		log.Fatalln("Could not find end index: ", err)
	}
	end_view_pos := view.text_pos_to_view_pos(end_text_pos)

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
			set_style(screen, view_pos_to_screen_pos(Point{col: i, row: row}, view.roi), cursor.style)
		}
	}
}

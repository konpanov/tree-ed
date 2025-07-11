package main

import (
	"fmt"
	"log"
	"math"

	"github.com/gdamore/tcell/v2"
)

type View interface {
	Draw()
	SetRoi(roi Rect)
	GetRoi() Rect
}

/******************************************************/
/*                     Utils                          */
/******************************************************/

func view_pos_to_screen_pos(pos Point, roi Rect) Point {
	col := pos.col + roi.left
	row := pos.row + roi.top
	screen_pos := Point{col: col, row: row}
	if row >= roi.bot || row < roi.top {
		log.Panicf("View should not draw outside it's roi (horizontal).\nView roi: %+v.\nView height: %d\nScreen position: %+v.\nView position: %+v.\n", roi, roi.Height(), screen_pos, pos)
	}
	if col >= roi.right || col < roi.left {
		log.Panicf("View should not draw outside it's roi (vertical).\nView roi: %+v.\nView width: %d.\nScreen position: %+v.\nView position: %+v.\n", roi, roi.Width(), screen_pos, pos)
	}
	return screen_pos
}


var ErrOutOfFrame = fmt.Errorf("")
var ErrLeftOfFrame = fmt.Errorf("Text position is left of window frame.%w", ErrOutOfFrame)
var ErrRightOfFrame = fmt.Errorf("Text position is right of window frame.%w", ErrOutOfFrame)
var ErrAboveFrame = fmt.Errorf("Text position is above window frame.%w", ErrOutOfFrame)
var ErrBelowFrame = fmt.Errorf("Text position is below window frame.%w", ErrOutOfFrame)
func text_pos_to_view_pos(pos Point, text_offset Point, roi Rect) (Point, error) {
	if pos.col < 0 || pos.row < 0 {
		log.Panicf("Text position coordinates should not be negative %+v", pos)
	}
	var err error
	if pos.col < text_offset.col {
		err = ErrLeftOfFrame
	} else if pos.col >= text_offset.col + roi.Width() {
		err = ErrRightOfFrame
	} else if pos.row < text_offset.row {
		err = ErrAboveFrame
	} else if pos.row >= text_offset.row + roi.Height(){
		err = ErrBelowFrame
	}
	if err != nil {
		return Point{}, fmt.Errorf("%w Text pos: %+v, Text offset: %+v, View roi: %+v", err, pos, text_offset, roi)
	}
	view_pos := Point{
		col: min(pos.col-text_offset.col, roi.Width()-1),
		row: min(pos.row-text_offset.row, roi.Height()-1),
	}
	if view_pos.row < 0 || view_pos.col < 0 {
		log.Panicf("View position should not be negative. View pos: %+v, Text offset: %+v, Text pos: %+v", view_pos, text_offset, pos)
	}
	return view_pos, nil
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

func default_buffer_line_number_max_width(buffer IBuffer) int {
	return int(math.Log10(float64(len(buffer.Lines())))) + 2
}

package main

import (
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

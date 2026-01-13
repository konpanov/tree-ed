package main

import (
	"log"
	"math"

	"github.com/gdamore/tcell/v2"
)

type View interface {
	Draw(ctx DrawContext)
}

type DrawContext struct {
	screen tcell.Screen
	roi    Rect
	theme  Theme
}

type StyleMod func(style tcell.Style) tcell.Style

type Theme struct {
	base         StyleMod
	text         StyleMod
	selection    StyleMod
	nodeOdd      StyleMod
	nodeEven     StyleMod
	node         StyleMod
	secondary    StyleMod
	secondary_bg StyleMod
}

var default_theme = DefaultTheme()

func CombineMods(mods []StyleMod) StyleMod {
	return func(style tcell.Style) tcell.Style {
		for _, m := range mods {
			style = m(style)
		}
		return style
	}
}

func DefaultTheme() Theme {
	type S = tcell.Style
	var hex = tcell.NewHexColor
	return Theme{
		base:         func(s S) S { return s.Background(hex(0x0C060F)).Foreground(hex(0xD6D6D6)) },
		selection:    func(s S) S { return s.Background(hex(0x426154)) },
		nodeOdd:      func(s S) S { return s.Background(hex(0x262D40)) },
		nodeEven:     func(s S) S { return s.Background(hex(0x3E2640)) },
		secondary:    func(s S) S { return s.Foreground(hex(0x938581)) },
		secondary_bg: func(s S) S { return s.Background(hex(0x211D1C)) },
		node:         func(s S) S { return s.Background(hex(0x2C232F)) },
	}
}

/******************************************************/
/*                     Utils                          */
/******************************************************/

func view_pos_to_screen_pos(pos Pos, roi Rect) Pos {
	col := pos.col + roi.left
	row := pos.row + roi.top
	screen_pos := Pos{col: col, row: row}
	if row >= roi.bot || row < roi.top {
		log.Panicf("View should not draw outside it's roi (horizontal).\nView roi: %+v.\nView height: %d\nScreen position: %+v.\nView position: %+v.\n", roi, roi.Height(), screen_pos, pos)
	}
	if col >= roi.right || col < roi.left {
		log.Panicf("View should not draw outside it's roi (vertical).\nView roi: %+v.\nView width: %d.\nScreen position: %+v.\nView position: %+v.\n", roi, roi.Width(), screen_pos, pos)
	}
	return screen_pos
}

func text_pos_to_view_pos(pos Pos, text_offset Pos, roi Rect) Pos {
	if debug {
		if pos.col < 0 || pos.row < 0 {
			log.Panicf("Text position should not be negative %+v", pos)
		}
		if pos.col < 0 || pos.row < 0 {
			log.Panicf("Text position should not be negative %+v", pos)
		} else if pos.col < text_offset.col {
			log.Panicln("Text position is left of window frame.")
		} else if pos.col >= text_offset.col+roi.Width() {
			log.Panicln("Text position is right of window frame.")
		} else if pos.row < text_offset.row {
			log.Panicln("Text position is above window frame.")
		} else if pos.row >= text_offset.row+roi.Height() {
			log.Panicln("Text position is below window frame.")
		}
	}

	view_pos := Pos{
		col: min(pos.col-text_offset.col, roi.Width()-1),
		row: min(pos.row-text_offset.row, roi.Height()-1),
	}
	if view_pos.row < 0 || view_pos.col < 0 {
		log.Panicf("View position should not be negative. View pos: %+v, Text offset: %+v, Text pos: %+v", view_pos, text_offset, pos)
	}
	return view_pos
}

func text_pos_to_screen(pos Pos, offset Pos, roi Rect) Pos {
	in_view := text_pos_to_view_pos(pos, offset, roi)
	in_screen := view_pos_to_screen_pos(in_view, roi)
	return in_screen
}

func set_style(screen tcell.Screen, pos Pos, style tcell.Style) {
	value, _, _, _ := screen.GetContent(pos.col, pos.row)
	screen.SetContent(pos.col, pos.row, value, nil, style)
}

func get_style(screen tcell.Screen, pos Pos) tcell.Style {
	_, _, style, _ := screen.GetContent(pos.col, pos.row)
	return style
}

func set_rune(screen tcell.Screen, pos Pos, value rune) {
	style := get_style(screen, pos)
	if value == '\r' || value == '\n' {
		value = ' '
	}
	screen.SetContent(pos.col, pos.row, value, nil, style)
}

func apply_mod(screen tcell.Screen, pos Pos, mod StyleMod) {
	style := get_style(screen, pos)
	style = mod(style)
	set_style(screen, pos, style)
}

func default_buffer_line_number_max_width(buffer IBuffer) int {
	return int(math.Log10(float64(len(buffer.Lines())))) + 2
}

func put_line(screen tcell.Screen, pos Pos, text string, stop int) {
	for i, r := range []rune(text) {
		row := pos.row
		col := pos.col + i
		if col >= stop {
			return
		}
		pos := Pos{row: row, col: col}
		set_rune(screen, pos, r)
	}
}

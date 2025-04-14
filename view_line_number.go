package main

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type LineNumberView interface{ Width() int }

type NoLineNumberView struct {
}

func (self NoLineNumberView) Width() int {
	return 0
}

func (self NoLineNumberView) Draw() {
}

type AbsoluteLineNumberView struct {
	screen      tcell.Screen
	roi         Rect
	buffer      IBuffer
	line_offset int
}

func (self AbsoluteLineNumberView) Width() int {
	return default_buffer_line_number_max_width(self.buffer)
}

func (self AbsoluteLineNumberView) Draw() {
	lines := self.buffer.Lines()
	width := self.roi.Width()
	height := self.roi.Height()
	start_line := max(min(self.line_offset, len(lines)), 0)
	end_line := max(min(start_line+height, len(lines)), 0)

	for y := range lines[start_line:end_line] {
		line_num := strconv.Itoa(start_line + y + 1)
		for i, r := range line_num {
			screen_pos := view_pos_to_screen_pos(
				Point{col: width - 1 - len(line_num) + i, row: y},
				self.roi,
			)
			set_rune(self.screen, screen_pos, r)
		}
	}
}

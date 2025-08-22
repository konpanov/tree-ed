package main

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type AbsoluteLineNumberView struct {
	screen tcell.Screen
	roi    Rect
	window *Window
}

func (self AbsoluteLineNumberView) Draw() {
	top := self.window.frame.top
	bot := min(self.window.frame.bot, len(self.window.buffer.Lines()))
	for line := top; line < bot; line++ {
		pos := view_pos_to_screen_pos(Point{row: line - top, col: 0}, self.roi)
		text := strconv.Itoa(line + 1)
		put_line(self.screen, pos, text, self.roi.right)
	}
}

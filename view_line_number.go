package main

import (
	"strconv"
)

type LineNumberView struct {
	window *Window
}

func (self LineNumberView) Draw(ctx DrawContext) {
	start := min(self.window.frame.top, len(self.window.buffer.Lines()))
	end := min(self.window.frame.bot, len(self.window.buffer.Lines()))

	for i := 0; i < end-start; i++ {
		pos := Pos{col: 0, row: i}
		line := strconv.Itoa(start + i + 1)
		pos = view_pos_to_screen_pos(pos, ctx.roi)
		put_line(ctx.screen, pos, line, ctx.roi.right)
	}

	for y := ctx.roi.top; y < ctx.roi.bot; y++ {
		for x := ctx.roi.left; x < ctx.roi.right; x++ {
			apply_mod(ctx.screen, Pos{row: y, col: x}, ctx.theme.secondary)
		}
	}
}

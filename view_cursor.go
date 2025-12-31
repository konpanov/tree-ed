package main

import (
	"github.com/gdamore/tcell/v2"
)

type CharacterViewCursor struct {
	window *Window
}

func (self *CharacterViewCursor) Draw(ctx DrawContext) {
	pos := self.window.cursor.Pos()
	offset := self.window.frame.TopLeft()
	rel_pos := self.window.frame.RelativePosition(pos)
	assert(rel_pos == Inside, "Main cursor should always be in frame")
	screen_pos := text_pos_to_screen(pos, offset, ctx.roi)
	ctx.screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	ctx.screen.ShowCursor(screen_pos.col, screen_pos.row)
}

type IndexViewCursor struct {
	window *Window
}

func (self *IndexViewCursor) Draw(ctx DrawContext) {
	pos := self.window.cursor.Pos()
	offset := self.window.frame.TopLeft()
	rel_pos := self.window.frame.RelativePosition(pos)
	assert(rel_pos == Inside, "Main cursor should always be in frame")
	screen_pos := text_pos_to_screen(pos, offset, ctx.roi)
	ctx.screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
	ctx.screen.ShowCursor(screen_pos.col, screen_pos.row)
}

type SelectionViewCursor struct {
	window *Window
}

func (self *SelectionViewCursor) Draw(ctx DrawContext) {
	buffer := self.window.buffer
	frame := self.window.frame

	start_index, end_index := self.window.getSelection()
	start_index = max(start_index, uint(buffer.Index(frame.TopLeft())))
	end_index = min(end_index, uint(buffer.Index(frame.BotRight())))

	cursor := self.window.cursor.AsEdge().ToIndex(int(start_index))
	for ; cursor.Index() < int(end_index); cursor = cursor.RuneNext() {
		pos := cursor.Pos()
		line := cursor.buffer.Lines()[pos.row]

		if frame.RelativePosition(pos) != Inside {
			continue
		}
		if cursor.IsLineBreak() && line.start != line.end {
			continue
		}

		screen_pos := text_pos_to_screen(pos, frame.TopLeft(), ctx.roi)
		apply_mod(ctx.screen, screen_pos, ctx.theme.selection)
	}

	ctx.screen.ShowCursor(-1, -1)
}

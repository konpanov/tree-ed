package main

import "github.com/gdamore/tcell/v2"

type ViewCursor2 interface {
}

type CharacterViewCursor2 struct {
	screen      tcell.Screen
	roi         Rect
	buffer      IBuffer
	cursor      BufferCursor
	text_offset Point
}

func NewCharacterViewCursor2(
	screen tcell.Screen,
	roi Rect,
	buffer IBuffer,
	cursor BufferCursor,
	text_offset Point,
) *CharacterViewCursor2 {
	return &CharacterViewCursor2{
		screen:      screen,
		roi:         roi,
		buffer:      buffer,
		cursor:      cursor,
		text_offset: text_offset,
	}
}

func (self *CharacterViewCursor2) Draw() {
	coord := self.cursor.RunePosition()
	pos := text_pos_to_view_pos(coord, self.text_offset, self.roi)
	pos = view_pos_to_screen_pos(pos, self.roi)

	self.screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	self.screen.ShowCursor(pos.col, pos.row)
}

type IndexViewCursor struct {
	screen      tcell.Screen
	roi         Rect
	buffer      IBuffer
	cursor      BufferCursor
	text_offset Point
}

func (self *IndexViewCursor) Draw() {
	coord := self.cursor.RunePosition()
	pos := text_pos_to_view_pos(coord, self.text_offset, self.roi)
	pos = view_pos_to_screen_pos(pos, self.roi)

	self.screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
	self.screen.ShowCursor(pos.col, pos.row)
}

type SelectionViewCursor struct {
	screen      tcell.Screen
	roi         Rect
	buffer      IBuffer
	cursorA     BufferCursor
	cursorB     BufferCursor
	text_offset Point
	style       tcell.Style
}

func (self *SelectionViewCursor) Draw() {
	start := self.cursorA
	end := self.cursorB
	if self.cursorA.index > self.cursorB.index {
		start, end = end, start
	}

	for cursor := start; cursor.index <= end.index; cursor, _ = cursor.RunesForward(1) {
		if !cursor.IsNewLine() {
			pos := text_pos_to_view_pos(cursor.RunePosition(), self.text_offset, self.roi)
			pos = view_pos_to_screen_pos(pos, self.roi)
			set_style(self.screen, pos, self.style)
		}
	}

	self.screen.ShowCursor(-1, -1)
}

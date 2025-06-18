package main

import "github.com/gdamore/tcell/v2"

type ViewCursor interface {
}

type CharacterViewCursor struct {
	screen      tcell.Screen
	roi         Rect
	buffer      IBuffer
	cursor      BufferCursor
	text_offset Point
}

func NewCharacterViewCursor(
	screen tcell.Screen,
	roi Rect,
	buffer IBuffer,
	cursor BufferCursor,
	text_offset Point,
) *CharacterViewCursor {
	return &CharacterViewCursor{
		screen:      screen,
		roi:         roi,
		buffer:      buffer,
		cursor:      cursor,
		text_offset: text_offset,
	}
}

func (self *CharacterViewCursor) GetRoi() Rect {
	return self.roi
}

func (self *CharacterViewCursor) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *CharacterViewCursor) Draw() {
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

func (self *IndexViewCursor) GetRoi() Rect {
	return self.roi
}

func (self *IndexViewCursor) SetRoi(roi Rect) {
	self.roi = roi
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

func (self *SelectionViewCursor) GetRoi() Rect {
	return self.roi
}

func (self *SelectionViewCursor) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *SelectionViewCursor) Draw() {
	// Get ordered start and end
	start, end := self.cursorA, self.cursorB
	if self.cursorA.index > self.cursorB.index {
		start, end = end, start
	}

	// If start is above the screen move start to the start of the screen
	screen_start_index, err := self.buffer.IndexFromRuneCoord(self.text_offset)
	panic_if_error(err)
	if start.Index() < screen_start_index {
		start, err = start.ToIndex(screen_start_index)
		panic_if_error(err)
	}

	height := self.roi.Height()
	for cursor := start; cursor.index < end.index; cursor, _ = cursor.RunesForward(1) {
		rune_pos := cursor.RunePosition()

		// If rune is below screen stop
		if rune_pos.row > self.text_offset.row+height {
			break
		}

		if !cursor.IsNewLine() {
			pos := text_pos_to_view_pos(rune_pos, self.text_offset, self.roi)
			pos = view_pos_to_screen_pos(pos, self.roi)
			set_style(self.screen, pos, self.style)
		}
	}

	self.screen.ShowCursor(-1, -1)
}

package main

import (
	"errors"
	"log"

	"github.com/gdamore/tcell/v2"
)

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
	pos, err := text_pos_to_view_pos(coord, self.text_offset, self.roi)
	if errors.Is(err, ErrOutOfFrame) {
		self.screen.ShowCursor(-1, -1)
		return
	} else if err != nil {
		log.Panic(err)
	}
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
	pos, err := text_pos_to_view_pos(coord, self.text_offset, self.roi)
	if errors.Is(err, ErrOutOfFrame) {
		self.screen.ShowCursor(-1, -1)
		return
	} else if err != nil {
		log.Panic(err)
	}
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
	// Order start and end cursors
	start, end := self.cursorA, self.cursorB
	if start.Index() > end.Index() {
		start, end = end, start
	}

	// If start is above the screen move start to the start of the screen
	screen_start_index, err := self.buffer.IndexFromRuneCoord(self.text_offset)
	panic_if_error(err)
	if start.Index() < screen_start_index {
		start, err = start.ToIndex(screen_start_index)
		panic_if_error(err)
	}

	for cursor := start; cursor.index <= end.index && err == nil; cursor, err = cursor.RunesForward(1) {
		rune_pos := cursor.RunePosition()

		pos, err := text_pos_to_view_pos(rune_pos, self.text_offset, self.roi)
		if err != nil {
			if errors.Is(err, ErrAboveFrame) {
				log.Panicf("Cursor is above frame during visual selection, but it should have been move to the start of the frame earlier, %s", err)
			} else if errors.Is(err, ErrRightOfFrame) {
				continue
			} else if errors.Is(err, ErrLeftOfFrame) {
				continue
			} else if errors.Is(err, ErrBelowFrame) {
				break
			}
		}
		pos = view_pos_to_screen_pos(pos, self.roi)
		set_style(self.screen, pos, self.style)

		if cursor.IsNewLine(){
			set_rune(self.screen, pos, '\u21B5')
		}

	}

	self.screen.ShowCursor(-1, -1)
}



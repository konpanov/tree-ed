package main

import "github.com/gdamore/tcell/v2"

type WindowView struct {
	screen       tcell.Screen
	roi          Rect
	buffer       IBuffer
	cursor       BufferCursor
	secondCursor BufferCursor
	mode         WindowMode
	text_offset  Point
}

func NewWindowView(
	screen tcell.Screen,
	roi Rect,
	buffer IBuffer,
	cursor BufferCursor,
	secondCursor BufferCursor,
	mode WindowMode,
) *WindowView {
	view := &WindowView{
		screen:       screen,
		roi:          roi,
		buffer:       buffer,
		cursor:       cursor,
		mode:         mode,
		secondCursor: secondCursor,
		text_offset:  Point{0, 0},
	}
	view.Update(roi, cursor, secondCursor, mode)
	return view
}

func (self *WindowView) Update(roi Rect, cursor BufferCursor, secondCursor BufferCursor, mode WindowMode) {
	self.roi = roi
	self.cursor = cursor
	self.secondCursor = secondCursor
	self.mode = mode
}

func (self *WindowView) Draw() {
	line_numbers_width := default_buffer_line_number_max_width(self.buffer)

	line_numbers_roi := self.roi.SetRight(line_numbers_width)
	text_roi := self.roi.SetLeft(line_numbers_roi.Right())

	text, text_offset := self.get_text_from_buffer(text_roi, self.text_offset)
	self.text_offset = text_offset

	line_numbers := AbsoluteLineNumberView{self.screen, line_numbers_roi, self.buffer, self.text_offset.row}
	text_view := NewTextView2(self.screen, text_roi, text)

	line_numbers.Draw()
	text_view.Draw()

	var cursor_view View2

	switch self.mode {
	case InsertMode:
		cursor_view = &IndexViewCursor{self.screen, text_roi, self.buffer, self.cursor, self.text_offset}
	case VisualMode:
		cursor_view = &SelectionViewCursor{self.screen, text_roi, self.buffer, self.cursor, self.secondCursor, self.text_offset, tcell.StyleDefault.Reverse(true)}
	default:
		cursor_view = &CharacterViewCursor2{self.screen, text_roi, self.buffer, self.cursor, self.text_offset}
	}

	cursor_view.Draw()
}

func (self *WindowView) get_text_from_buffer(roi Rect, text_offset Point) ([][]rune, Point) {
	width := roi.Width()
	height := roi.Height()
	coord := self.cursor.RunePosition()
	text_offset = Point{
		col: max(min(text_offset.col, coord.col), coord.col-width+1),
		row: max(min(text_offset.row, coord.row), coord.row-height+1),
	}

	lines := self.buffer.Lines()
	lines = lines[min(len(lines), self.text_offset.row):min(len(lines), self.text_offset.row+height)]
	text := [][]rune{}

	for _, region := range lines {
		line := []rune(string(self.buffer.Content()[region.start:region.end]))
		line = line[min(len(line), self.text_offset.col):min(len(line), self.text_offset.col+width)]
		text = append(text, line)
	}
	return text, text_offset
}

package main

import (
	"github.com/gdamore/tcell/v2"
)

type WindowView struct {
	screen      tcell.Screen
	roi         Rect
	window      *Window
	text_offset Point

	// Widgets
	status_line IStatusLine
	is_tree_view bool
}

func NewWindowView(
	screen tcell.Screen,
	roi Rect,
	window *Window,
) *WindowView {
	view := &WindowView{
		screen:      screen,
		roi:         roi,
		window:      window,
		text_offset: Point{0, 0},
		status_line: NoStatusLine{},
		is_tree_view: false, // TODO separate tree view from window view
	}
	view.Update(roi)
	return view
}

func (self *WindowView) GetRoi() Rect {
	return self.roi
}

func (self *WindowView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *WindowView) Update(roi Rect) {
	self.roi = roi
}

func (self *WindowView) Draw() {
	line_numbers_width := default_buffer_line_number_max_width(self.window.buffer)
	status_line_height := self.status_line.GetHeight()

	status_line_start_row := self.roi.Bot() - status_line_height
	line_numbers_end_col := self.roi.Left() + line_numbers_width

	line_numbers_roi := self.roi.SetRight(line_numbers_end_col).SetBot(status_line_start_row)
	self.status_line.SetRoi(self.roi.SetTop(status_line_start_row))
	main_roi := self.roi.SetLeft(line_numbers_end_col).SetBot(status_line_start_row)

	text_roi := main_roi
	if self.is_tree_view{
		text_roi = main_roi.SetRight(main_roi.Right() - main_roi.Width()/2)
	}

	text, text_offset := self.get_text_from_buffer(text_roi, self.text_offset)
	self.text_offset = text_offset

	line_numbers := AbsoluteLineNumberView{self.screen, line_numbers_roi, self.window.buffer, self.text_offset.row}
	text_view := NewTextView(self.screen, text_roi, text)

	var cursor_view View

	switch self.window.mode {
	case InsertMode:
		cursor_view = &IndexViewCursor{self.screen, text_roi, self.window.buffer, self.window.cursor, self.text_offset}
	case VisualMode, TreeMode:
		cursor_view = &SelectionViewCursor{
			self.screen,
			text_roi,
			self.window.buffer,
			self.window.cursor,
			self.window.secondCursor,
			self.text_offset,
			tcell.StyleDefault.Reverse(true).Underline(true),
		}
	default:
		cursor_view = &CharacterViewCursor{self.screen, text_roi, self.window.buffer, self.window.cursor, self.text_offset}
	}

	if self.is_tree_view {
		tree_roi := main_roi.SetLeft(main_roi.Left() + main_roi.Width()/2)
		tree_view := TreeView{screen: self.screen, roi: tree_roi, window: self.window}
		tree_view.Draw()
	}

	cursor_view.Draw()
	line_numbers.Draw()
	text_view.Draw()
	self.status_line.Draw()

}

func (self *WindowView) get_text_from_buffer(roi Rect, text_offset Point) ([][]rune, Point) {
	width := roi.Width()
	height := roi.Height()
	coord := self.window.cursor.RunePosition()
	text_offset = Point{
		col: max(min(text_offset.col, coord.col), coord.col-width+1),
		row: max(min(text_offset.row, coord.row), coord.row-height+1),
	}

	if coord.col < width {
		text_offset.col = 0
	}

	lines := self.window.buffer.Lines()
	lines = lines[min(len(lines), text_offset.row):min(len(lines), text_offset.row+height)]
	text := [][]rune{}

	for _, region := range lines {
		line := []rune(string(self.window.buffer.Content()[region.start:region.end]))
		line = line[min(len(line), text_offset.col):min(len(line), text_offset.col+width)]
		text = append(text, line)
	}
	return text, text_offset
}

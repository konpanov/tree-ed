package main

import (
	"errors"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type WindowView struct {
	screen      tcell.Screen
	roi         Rect
	window      *Window
	text_offset Point

	// Widgets
	status_line IStatusLine
	text_style  tcell.Style
	base_style  tcell.Style

	is_tree_view       bool
	is_newline_symbols bool
}

func NewWindowView(
	screen tcell.Screen,
	roi Rect,
	window *Window,
) *WindowView {
	base_style := tcell.StyleDefault
	base_style = base_style.Background(tcell.NewHexColor(SPACE_CADET))
	view := &WindowView{
		screen:      screen,
		roi:         roi,
		window:      window,
		text_offset: Point{0, 0},
		status_line: NoStatusLine{},
		base_style:  base_style,

		is_tree_view:       window.buffer.Tree() != nil && false, // TODO separate tree view from window view
		is_newline_symbols: true,
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

	main_roi, status_line_roi := self.roi.SplitH(self.roi.Height() - status_line_height)
	line_numbers_roi, main_roi := main_roi.SplitV(line_numbers_width)
	self.status_line.SetRoi(status_line_roi)

	var tree_roi Rect
	if self.is_tree_view {
		main_roi, tree_roi = main_roi.SplitV(main_roi.Width() / 2)
	}

	text, text_offset := self.get_text_from_buffer(main_roi, self.text_offset)
	self.text_offset = text_offset

	line_numbers := AbsoluteLineNumberView{self.screen, line_numbers_roi, self.window.buffer, self.text_offset.row}
	text_view := NewTextView(self.screen, main_roi, text)
	text_view.style = self.base_style

	var cursor_view View

	switch self.window.mode {
	case InsertMode:
		cursor_view = &IndexViewCursor{self.screen, main_roi, self.window.buffer, self.window.cursor, self.text_offset}
	case VisualMode, TreeMode:
		cursor_view = &SelectionViewCursor{
			self.screen,
			main_roi,
			self.window.buffer,
			self.window.cursor,
			self.window.secondCursor,
			self.text_offset,
			tcell.StyleDefault.Background(tcell.NewHexColor(GLACIOUS)),
		}
	default:
		cursor_view = &CharacterViewCursor{self.screen, main_roi, self.window.buffer, self.window.cursor, self.text_offset}
	}

	tree_color := &TreeColorView{
		screen:      self.screen,
		window:      self.window,
		text_offset: text_offset,
		base_style:  self.base_style,
	}
	tree_color.SetRoi(main_roi)
	line_numbers.Draw()
	text_view.Draw()
	self.status_line.Draw()
	if self.window.buffer.Tree() != nil {
		tree_color.Draw()
		if self.is_tree_view {
			tree_view := TreeView{screen: self.screen, roi: tree_roi, window: self.window}
			tree_view.style = self.base_style
			tree_view.Draw()
		}
	}
	cursor_view.Draw()

	pos, err := text_pos_to_screen(self.window.secondCursor.RunePosition(), text_offset, main_roi)
	if !errors.Is(err, ErrOutOfFrame) {
		panic_if_error(err)
		set_style(self.screen, pos, get_style(self.screen, pos).Background(tcell.NewHexColor(GLACIOUS)))
	}

}

func (self *WindowView) get_text_from_buffer(roi Rect, text_offset Point) ([][]rune, Point) {
	width := roi.Width()
	height := roi.Height()

	if self.window.mode == VisualMode || self.window.mode == TreeMode {
		coord := self.window.secondCursor.RunePosition()
		text_offset = Point{
			col: max(min(text_offset.col, coord.col), coord.col-width+1),
			row: max(min(text_offset.row, coord.row), coord.row-height+1),
		}
	}

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
		content := self.window.buffer.Content()
		start, end := region.start, region.end
		if self.is_newline_symbols {
			end = region.full_end
		}
		line := []rune(string(content[start:end]))
		line = line[min(len(line), text_offset.col):min(len(line), text_offset.col+width)]
		text = append(text, line)
	}
	return text, text_offset
}

type TreeColorView struct {
	screen      tcell.Screen
	roi         Rect
	text_offset Point
	window      *Window
	base_style  tcell.Style
}

func (self *TreeColorView) Draw() {
	if self.window.mode == TreeMode {
		depth := self.window.anchorDepth
		node := self.window.getNode()
		first_node := node
		for {
			sibl := PrevSiblingOrCousinDepth(first_node, depth)
			if sibl == nil {
				break
			}
			first_node = sibl
		}

		odd := true
		for next := first_node; next != nil; next = NextSiblingOrCousinDepth(next, depth) {
			var style tcell.Style
			if odd {
				style = self.base_style.Background(tcell.NewHexColor(0x384251))
			} else {
				style = self.base_style.Background(tcell.NewHexColor(0x504C4A))
			}
			odd = !odd
			self.ColorNode(next, style)
		}
	} else {
		node := self.window.getNode()
		self.ColorNode(node, self.base_style.Underline(true))
	}
}

func (self *TreeColorView) ColorNode(node *sitter.Node, style tcell.Style) {
	start, end := int(node.StartByte()), int(node.EndByte())
	cursor := BufferCursor{buffer: self.window.buffer}.AsChar().ToIndex(start)
	for ; !cursor.IsEnd() && cursor.Index() < end; cursor = cursor.RuneNext() {
		if cursor.IsNewLine() {
			continue
		}
		pos := cursor.RunePosition()
		pos, err := text_pos_to_screen(pos, self.text_offset, self.roi)
		if err == nil {
			set_style(self.screen, pos, style)
		}

	}
}

func (self *TreeColorView) GetRoi() Rect {
	return self.roi
}

func (self *TreeColorView) SetRoi(roi Rect) {
	self.roi = roi
}

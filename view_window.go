package main

import (
	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type WindowView struct {
	screen      tcell.Screen
	roi         Rect
	window      *Window
	text_offset Point

	// Widgets
	status_line  IStatusLine
	is_tree_view bool
	text_style   tcell.Style
	base_style   tcell.Style
}

func NewWindowView(
	screen tcell.Screen,
	roi Rect,
	window *Window,
) *WindowView {
	base_style := tcell.StyleDefault
	base_style = base_style.Background(tcell.NewHexColor(SPACE_CADET))
	// base_style = base_style.Foreground(tcell.NewHexColor(0xC3A995))
	view := &WindowView{
		screen:       screen,
		roi:          roi,
		window:       window,
		text_offset:  Point{0, 0},
		status_line:  NoStatusLine{},
		is_tree_view: true, // TODO separate tree view from window view
		base_style:   base_style,
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
			tcell.StyleDefault.Reverse(true).Underline(true),
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

	if self.is_tree_view {
		tree_view := TreeView{screen: self.screen, roi: tree_roi, window: self.window}
		tree_view.style = self.base_style
		tree_view.Draw()
	}

	cursor_view.Draw()
	line_numbers.Draw()
	text_view.Draw()
	tree_color.Draw()
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
			sibl := PrevCousinDepth(first_node, depth)
			if sibl == nil {
				break
			}
			first_node = sibl
		}

		odd := true
		node = first_node
		for cousin := NextCousinDepth(node, depth); cousin != nil; cousin = NextCousinDepth(cousin, depth) {
			var style tcell.Style
			if odd {
				style = self.base_style.Background(tcell.NewHexColor(0x384251))
			} else {
				style = self.base_style.Background(tcell.NewHexColor(0x504C4A))
			}
			odd = !odd
			self.ColorNode(cousin, style)
		}

		node = self.window.getNode()
		self.ColorNode(node, self.base_style.Background(tcell.NewHexColor(GLACIOUS)).Underline(true))

	} else {
		node := NodeLeaf(self.window.buffer.Tree().RootNode(), self.window.cursor.Index())
		self.ColorNode(node, self.base_style.Background(tcell.NewHexColor(GLACIOUS)))
	}

}

func (self *TreeColorView) ColorNode(node *sitter.Node, style tcell.Style) {
	for index := int(node.StartByte()); index < int(node.EndByte()); index++ {
		pos, err := self.window.buffer.RuneCoord(index)
		panic_if_error(err)
		pos, err = text_pos_to_screen(pos, self.text_offset, self.roi)
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

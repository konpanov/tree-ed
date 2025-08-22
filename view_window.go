package main

import (
	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type WindowView struct {
	screen tcell.Screen
	roi    Rect
	window *Window

	text_style tcell.Style
	base_style tcell.Style

	is_tree_view       bool
	is_newline_symbols bool
}

func NewWindowView(
	screen tcell.Screen,
	roi Rect,
	window *Window,
) *WindowView {
	view := &WindowView{
		screen: screen,
		roi:    roi,
		window: window,

		is_tree_view:       window.buffer.Tree() != nil && false, // TODO separate tree view from window view
		is_newline_symbols: false,
	}
	view.SetRoi(roi)
	return view
}

func (self *WindowView) GetRoi() Rect {
	return self.roi
}

func (self *WindowView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *WindowView) Draw() {
	line_numbers_width := default_buffer_line_number_max_width(self.window.buffer)
	line_numbers_roi, main_roi := self.roi.SplitV(line_numbers_width)

	self.window.ResizeFrame(main_roi.Width(), main_roi.Height())
	self.DrawFrameText(main_roi)
	self.DrawColoredTree(main_roi)
	self.DrawCursor(main_roi)

	line_numbers := AbsoluteLineNumberView{self.screen, line_numbers_roi, self.window}
	line_numbers.Draw()

	// var tree_roi Rect
	// if self.is_tree_view {
	// 	main_roi, tree_roi = main_roi.SplitV(main_roi.Width() / 2)
	// }
}

func (self *WindowView) DrawFrameText(roi Rect) {
	frame := self.window.frame
	offset := frame.TopLeft()
	cursor := self.window.cursor.AsEdge().MoveToRunePos(offset)
	for !cursor.IsEnd() {
		pos := cursor.RunePosition()
		rel_pos := frame.RelativePosition(pos)
		if rel_pos == Below {
			break
		}
		if rel_pos == Inside {
			for _, value := range RenderedRune(cursor.Rune()) {
				view_pos, err := text_pos_to_view_pos(pos, offset, roi)
				panic_if_error(err)
				screen_pos := view_pos_to_screen_pos(view_pos, roi)
				set_rune(self.screen, screen_pos, value)
			}
		}
		cursor = cursor.RuneNext()
	}
}

func (self *WindowView) DrawCursor(roi Rect) {
	var cursor_view View
	switch self.window.mode {
	case InsertMode:
		cursor_view = &IndexViewCursor{
			self.screen,
			roi,
			self.window.buffer,
			self.window.cursor,
			self.window.frame.TopLeft(),
		}
	case VisualMode, TreeMode:
		cursor_view = &SelectionViewCursor{
			self.screen,
			roi,
			self.window.buffer,
			self.window.cursor,
			self.window.anchor,
			self.window.frame.TopLeft(),
			tcell.StyleDefault.Background(tcell.NewHexColor(GLACIOUS)),
		}
	default:
		cursor_view = &CharacterViewCursor{
			self.screen,
			roi,
			self.window.buffer,
			self.window.cursor,
			self.window.frame.TopLeft(),
		}
	}
	cursor_view.Draw()
}

func (self *WindowView) DrawColoredTree(roi Rect) {
	tree_color := &TreeColorView{
		screen:      self.screen,
		window:      self.window,
		text_offset: self.window.frame.TopLeft(),
		base_style:  self.base_style,
	}
	tree_color.SetRoi(roi)
	if self.window.buffer.Tree() != nil {
		tree_color.Draw()
		// if self.is_tree_view {
		// 	tree_view := TreeView{screen: self.screen, roi: tree_roi, window: self.window}
		// 	tree_view.style = self.base_style
		// 	tree_view.Draw()
		// }
	}
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
		depth := self.window.originDepth
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
	frame := self.window.frame
	cursor := BufferCursor{buffer: self.window.buffer}.AsEdge().ToIndex(start)
	for !cursor.IsEnd() && cursor.Index() < end {
		pos := cursor.RunePosition()
		if frame.RelativePosition(pos) == Inside {
			pos, err := text_pos_to_screen(pos, frame.TopLeft(), self.roi)
			panic_if_error(err)
			set_style(self.screen, pos, style)
		}
		cursor = cursor.RuneNext()
	}

}

func (self *TreeColorView) GetRoi() Rect {
	return self.roi
}

func (self *TreeColorView) SetRoi(roi Rect) {
	self.roi = roi
}

package main

import (
	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type TreeView struct {
	screen tcell.Screen
	roi    Rect
	window *Window
	style  tcell.Style
}

func (self *TreeView) GetRoi() Rect {
	return self.roi
}

func (self *TreeView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *TreeView) Draw() {
	self.DrawNode(self.window.buffer.Tree().RootNode(), 0, 0)
}

func (self *TreeView) DrawNode(node *sitter.Node, row int, depth int) int {
	if row >= self.roi.Height() {
		return row
	}
	style := self.style
	pos := self.window.cursor.Index()
	if self.window.mode == TreeMode {
		if node == self.window.getNode() {
			style = tcell.StyleDefault.Background(tcell.ColorGray)
		}
	} else {
		if node.StartByte() <= uint32(pos) && uint32(pos) < node.EndByte() {
			style = tcell.StyleDefault.Background(tcell.ColorGray)
		}
	}
	text := []rune(node.Type())
	if len(text) == 0 {
		text = []rune(node.Content(self.window.buffer.Content()))
	}
	if depth != 0 {
		text = append([]rune("+-"), text...)
	}
	depth_offset := max(depth-1, 0) * 2
	width := self.roi.Width()
	text = text[:max(min(len(text), width-depth_offset), 0)]
	for i, r := range text {
		pos := Point{row: row, col: i + depth_offset}
		pos = view_pos_to_screen_pos(pos, self.roi)
		set_rune(self.screen, pos, r)
		set_style(self.screen, pos, style)
	}
	prev_row := row
	row += 1
	for i := 0; i < int(node.ChildCount()); i++ {
		for r := prev_row + 1; r < row; r++ {
			pos := Point{row: r, col: depth * 2}
			pos = view_pos_to_screen_pos(pos, self.roi)
			set_rune(self.screen, pos, '|')
		}
		next_row := self.DrawNode(node.Child(i), row, depth+1)
		prev_row = row
		row = next_row
	}
	return row
}

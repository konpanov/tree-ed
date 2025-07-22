package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
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
	win_node := self.window.getNode()
	self.DrawNode(self.window.buffer.Tree().RootNode(), win_node, 0, 0)
}

func (self *TreeView) DrawNode(node *sitter.Node, selected_node *sitter.Node, row int, depth int) int {
	if row >= self.roi.Height() {
		return row
	}
	style := self.style
	pos := self.window.cursor.Index()
	if self.window.mode == TreeMode {
		if node.StartByte() == selected_node.StartByte() && node.EndByte() == selected_node.EndByte() {
			style = tcell.StyleDefault.Background(tcell.ColorGray)
		}
		if Depth(node) == self.window.anchorDepth {
			style = style.Bold(true)
		}
	} else {
		if node.StartByte() <= uint(pos) && uint(pos) < node.EndByte() {
			style = tcell.StyleDefault.Background(tcell.ColorGray)
		}
	}
	text := []rune(node.Kind())
	if len(text) == 0 {
		text = []rune(node.Utf8Text(self.window.buffer.Content()))
	}
	text = append(text, []rune(fmt.Sprintf(" (%d-%d)", node.StartByte(), node.EndByte()))...)
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
	for i := uint(0); i < node.ChildCount() && row < self.roi.Height(); i++ {
		for r := prev_row + 1; r < row; r++ {
			pos := Point{row: r, col: depth * 2}
			pos = view_pos_to_screen_pos(pos, self.roi)
			set_rune(self.screen, pos, '|')
		}
		next_row := self.DrawNode(node.Child(i), selected_node, row, depth+1)
		prev_row = row
		row = next_row
	}
	return row
}

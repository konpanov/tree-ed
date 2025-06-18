package main

import (
	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type TreeView struct {
	screen tcell.Screen
	roi    Rect
	window *Window
}

func (self *TreeView) GetRoi() Rect {
	return self.roi
}

func (self *TreeView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *TreeView) Draw() {
	tree := self.window.buffer.Tree()
	root := tree.RootNode()
	pos := self.window.cursor.index
	style := tcell.StyleDefault

	stack := []*sitter.Node{root}
	row := 0
	for len(stack) > 0 && row < self.roi.Height() {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for i := node.ChildCount(); i > 0; i-- {
			stack = append(stack, node.Child(int(i)-1))
		}
		if node.StartByte() <= uint32(pos) && uint32(pos) <= node.EndByte() {
			style = tcell.StyleDefault.Background(tcell.ColorGray)
		} else {
			style = tcell.StyleDefault
		}
		for i, r := range []rune(node.Type()) {
			pos := Point{row: row, col: i}
			pos = view_pos_to_screen_pos(pos, self.roi)
			set_rune(self.screen, pos, r)
			set_style(self.screen, pos, style)
		}
		row++
	}
}

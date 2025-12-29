package main

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type TreeColorView struct {
	window *Window
}

func (self *TreeColorView) DrawNew(ctx DrawContext) {
	if self.window.buffer.Tree() == nil {
		return
	}
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
			if odd {
				self.ColorNode(ctx, next, ctx.theme.nodeOdd)
			} else {
				self.ColorNode(ctx, next, ctx.theme.nodeEven)
			}
			odd = !odd
		}
	} else {
		node := self.window.getNode()
		self.ColorNode(ctx, node, ctx.theme.node)
	}
}

func (self *TreeColorView) ColorNode(ctx DrawContext, node *sitter.Node, mod StyleMod) {
	frame := self.window.frame
	start, end := int(node.StartByte()), int(node.EndByte())
	start = max(start, self.window.buffer.Index(frame.TopLeft()))
	end = min(end, self.window.buffer.Index(frame.BotRight()))

	cursor := BufferCursor{buffer: self.window.buffer}.AsEdge().ToIndex(start)
	for ; !cursor.IsEnd() && cursor.Index() < end; cursor = cursor.RuneNext() {
		pos := cursor.Pos()
		line := cursor.buffer.Lines()[pos.row]
		if cursor.IsNewLine() && line.start != line.end {
			continue
		}
		if frame.RelativePosition(pos) == Inside {
			pos := text_pos_to_screen(pos, frame.TopLeft(), ctx.roi)
			apply_mod(ctx.screen, pos, mod)
		}
	}

}

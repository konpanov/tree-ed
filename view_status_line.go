package main

import (
	"strconv"
	"strings"
)

type StatusLine struct {
	editor *Editor
}

func (self StatusLine) DrawNew(ctx DrawContext) {
	left_parts := []string{}

	if self.editor.curwin != nil {
		curwin := self.editor.curwin
		pos := curwin.cursor.Pos()
		parseError := "Correct"
		if curwin.buffer.Tree() != nil && curwin.buffer.Tree().RootNode().HasError() {
			parseError = "Error"
		}
		newline := newlinesToSymbols([]rune(string(curwin.buffer.LineBreak())))
		left_parts = append(left_parts, "file: "+curwin.buffer.Filename())
		left_parts = append(left_parts, "line: "+strconv.Itoa(pos.row+1))
		left_parts = append(left_parts, "col: "+strconv.Itoa(pos.col+1))
		left_parts = append(left_parts, "mode: "+string(curwin.mode))
		left_parts = append(left_parts, "parse state: "+parseError)
		left_parts = append(left_parts, "newline: "+string(newline))
	}
	left_parts = append(left_parts, "input: "+KeyEventsToString(self.editor.scanner.state.Input()))

	text := []rune(strings.Join(left_parts, ", "))
	text = text[:min(ctx.roi.Width(), len(text))]
	mod := CombineMods([]StyleMod{ctx.theme.secondary, ctx.theme.secondary_bg})
	for y := ctx.roi.top; y < ctx.roi.bot; y++ {
		for x := ctx.roi.left; x < ctx.roi.right; x++ {
			apply_mod(ctx.screen, Pos{row: y, col: x}, mod)
		}
	}
	put_line(ctx.screen, ctx.roi.TopLeft(), string(text), ctx.roi.right)
}

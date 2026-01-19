package main

import (
	"fmt"
	"strconv"
	"strings"
)

type StatusLineView struct {
	editor *Editor
}

func (self StatusLineView) Draw(ctx DrawContext) {
	mode := self.modeDispaly()
	parse_state := self.parseStateDisplay()
	filename := self.filenameDisplay()
	pos := self.positionDisplay()
	linebreak := self.linebreakDisplay()
	input := self.inputDispaly()
	percent := self.percentDisplay()

	line1_left := fmt.Sprintf("%s %s", mode, parse_state)
	line1_right := fmt.Sprintf("%s %s", pos, percent)
	line1 := self.constructLine(ctx, line1_left, line1_right)
	put_line(ctx.screen, ctx.roi.TopLeft(), string(line1), ctx.roi.right)

	mod := CombineMods([]StyleMod{ctx.theme.secondary, ctx.theme.secondary_bg})
	for x := ctx.roi.left; x < ctx.roi.right; x++ {
		apply_mod(ctx.screen, Pos{row: ctx.roi.top, col: x}, mod)
	}
	if ctx.roi.Height() < 2 {
		return
	}

	line2_left := fmt.Sprintf("%s %s", filename, linebreak)
	line2_right := fmt.Sprintf("%s", input)
	line2 := self.constructLine(ctx, line2_left, line2_right)

	line2_start := ctx.roi.TopLeft()
	line2_start.row++
	put_line(ctx.screen, line2_start, string(line2), ctx.roi.right)

}

func (self StatusLineView) constructLine(ctx DrawContext, left string, right string) string {
	line := []rune(strings.Repeat(" ", ctx.roi.Width()))

	l := []rune(left)
	for i, char := range l[:min(len(line), len(l))] {
		line[i] = char
	}

	r := []rune(right)
	for i, char := range r {
		pos := max(0, len(line)-len(r)+i)
		line[pos] = char
	}

	return string(line)
}

func (self StatusLineView) modeDispaly() string {
	if self.editor.curwin == nil {
		return ""
	}
	return map[WindowMode]string{
		NormalMode: "[N]",
		VisualMode: "[V]",
		InsertMode: "[I]",
		TreeMode:   "[T]",
	}[self.editor.curwin.mode]
}

func (self StatusLineView) parseStateDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	curwin := self.editor.curwin
	if curwin.buffer.Tree() == nil {
		return ""
	}
	checkmark := "\u2713"
	crossmark := "\u2715"
	parseState := checkmark
	if curwin.buffer.Tree().RootNode().HasError() {
		parseState = crossmark
	}
	return parseState
}

func (self StatusLineView) filenameDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	return self.editor.curwin.buffer.Filename()
}

func (self StatusLineView) positionDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	curwin := self.editor.curwin
	pos := curwin.cursor.Pos()
	return fmt.Sprintf(
		"%s:%s",
		strconv.Itoa(pos.row+1),
		strconv.Itoa(pos.col+1),
	)
}

func (self StatusLineView) linebreakDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	curwin := self.editor.curwin
	res := []rune{}
	for _, r := range []rune(string(curwin.buffer.LineBreak())) {
		display, ok := map[rune][]rune{'\r': []rune("CR"), '\n': []rune("LF")}[r]
		if !ok {
			display = []rune{'X'}
		}
		res = append(res, display...)
	}
	return fmt.Sprintf("(%s)", string(res))
}

func (self StatusLineView) inputDispaly() string {
	keys := self.editor.scanner.Input()
	keys = keys[max(len(keys)-10, 0):]
	input := KeyEventsToString(keys)
	return fmt.Sprintf("%-10.10s", input)
}

func (self StatusLineView) percentDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	line_max := float32(len(self.editor.curwin.buffer.Lines()))
	line_cur := float32(self.editor.curwin.frame.bot)
	percent := min(max(line_cur/line_max, 0), 1)
	percent *= 100
	return fmt.Sprintf("%3.0f%%", percent)
}

func (self StatusLineView) horizonDisplay() string {
	if self.editor.curwin == nil {
		return ""
	}
	if self.editor.curwin.mode != TreeMode {
		return ""
	}
	line_max := float32(len(self.editor.curwin.buffer.Lines()))
	line_cur := float32(self.editor.curwin.frame.bot)
	percent := min(max(line_cur/line_max, 0), 1)
	percent *= 100
	return fmt.Sprintf("H%d", self.editor.curwin.originDepth)
}

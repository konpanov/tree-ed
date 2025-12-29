package main

import _ "embed"
import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

//go:embed assets/logo.txt
var logo []byte

type PreviewView struct{}

func (self PreviewView) DrawNew(ctx DrawContext) {
	lines := strings.Split(string(logo), string(getContentNewLine(logo)))
	lines_size := LinesSize(lines)
	lines_roi := CenterRoi(ctx.roi, lines_size)
	DrawLines(ctx.screen, lines_roi, lines)
}

func DrawLines(screen tcell.Screen, roi Rect, lines []string) {
	for row, line := range lines {
		for col, value := range []rune(line) {
			pos := Pos{row: row, col: col}
			pos = view_pos_to_screen_pos(pos, roi)
			set_rune(screen, pos, value)
		}
	}
}

func LinesSize(lines []string) Pos {
	h := len(lines)
	w := 0
	for _, line := range lines {
		w = max(len(line), w)
	}
	return Pos{row: h, col: w}
}

func CenterRoi(roi Rect, content_size Pos) Rect {
	new_roi := Rect{}
	new_roi.left = roi.left + roi.Width()/2 - content_size.col/2
	new_roi.top = roi.top + roi.Height()/2 - content_size.row/2
	new_roi.right = new_roi.left + content_size.col
	new_roi.bot = new_roi.top + content_size.row
	return new_roi
}

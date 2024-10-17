package main

import (
	"github.com/gdamore/tcell/v2"
)

type WindowCursor struct {
	index        int
	row          int
	column       int
	originColumn int
}

type Window struct {
	buffer        *Buffer
	cursor        *WindowCursor
	topLine       int // TODO: Should it be in WindowCursor?
	width, height int
}

func windowFromBuffer(buffer *Buffer, width int, height int) *Window {
	return &Window{
		buffer:  buffer,
		cursor:  &WindowCursor{0, 0, 0, 0},
		topLine: 0,
		width:   width,
		height:  height,
	}
}

func (window *Window) draw(screen tcell.Screen) {
	lineStart := 0
	screen.ShowCursor(window.cursor.column, window.cursor.row-window.topLine)
	for y, line := range window.buffer.lines {
		if y >= window.topLine && y < window.topLine+window.height {
			for x := 0; x < line.width; x++ {
				i := lineStart + x
				screen.SetContent(x, y-window.topLine, rune(window.buffer.content[i]), nil, tcell.StyleDefault)
			}
		}
		lineStart += line.width + len(window.buffer.newLineSeq)
	}
}

func (window *Window) cursorRight() {
	lineWidth := window.buffer.lines[window.cursor.row].width
	if window.cursor.column+1 == lineWidth {
		return
	}
	window.cursor.index++
	window.cursor.column++
	window.cursor.originColumn = window.cursor.column
}

func (window *Window) cursorLeft() {
	if window.cursor.column == 0 {
		return
	}
	window.cursor.index--
	window.cursor.column--
	window.cursor.originColumn = window.cursor.column
}

func (window *Window) cursorDown() {
	if window.cursor.row == len(window.buffer.lines)-1 {
		return
	}

	thisLineWidth := window.buffer.lines[window.cursor.row].width - 1
	nextLineWidth := window.buffer.lines[window.cursor.row+1].width - 1

	window.cursor.index -= window.cursor.column
	window.cursor.index += thisLineWidth + len(window.buffer.newLineSeq)
	window.cursor.column = max(min(window.cursor.originColumn, nextLineWidth), 0)
	window.cursor.row += 1
	window.cursor.index += window.cursor.column
	window.cursor.index += window.cursor.column
	window.topLine = max(window.topLine+window.height-1, window.cursor.row) - window.height + 1
}

func (window *Window) cursorUp() {
	if window.cursor.row == 0 {
		return
	}

	prevLineWidth := window.buffer.lines[window.cursor.row-1].width - 1

	window.cursor.row -= 1
	window.cursor.index -= prevLineWidth + len(window.buffer.newLineSeq)

	window.cursor.index -= window.cursor.column
	window.cursor.column = max(min(window.cursor.originColumn, prevLineWidth), 0)
	window.cursor.index += window.cursor.column

	window.topLine = min(window.topLine, window.cursor.row)
}

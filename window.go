package main

import (
	"github.com/gdamore/tcell/v2"
)

type WindowMode int

const (
	NormalMode WindowMode = iota
	InsertMode WindowMode = iota
	VisualMode WindowMode = iota
)

type WindowCursor struct {
	index        int
	row          int
	column       int
	originColumn int
}

type Window struct {
	mode          WindowMode
	buffer        *Buffer
	cursor        *WindowCursor
	secondCursor  *WindowCursor
	topLine       int // TODO: Should it be in WindowCursor?
	width, height int
}

func windowFromBuffer(buffer *Buffer, width int, height int) *Window {
	return &Window{
		mode:         NormalMode,
		buffer:       buffer,
		cursor:       &WindowCursor{0, 0, 0, 0},
		secondCursor: &WindowCursor{0, 0, 0, 0},
		topLine:      0,
		width:        width,
		height:       height,
	}
}

func (window *Window) draw(screen tcell.Screen) {
	switch window.mode {
	case NormalMode:
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	case VisualMode:
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	case InsertMode:
		screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
	}
	screen.ShowCursor(window.cursor.column, window.cursor.row-window.topLine)
	window.drawText(screen)
	if window.mode == VisualMode {
		window.drawSelection(screen, window.cursor, window.secondCursor)
	}
}

func (window *Window) drawText(screen tcell.Screen) {
	lineStart := 0
	for y, line := range window.buffer.lines {
		if y >= window.topLine && y < window.topLine+window.height {
			for x := 0; x < line.width; x++ {
				i := lineStart + x
				char := rune(window.buffer.content[i])
				style := tcell.StyleDefault
				screen.SetContent(x, y-window.topLine, char, nil, style)
			}
			// window.DEBUGdrawRN(screen, lineStart, y, line)
		}
		lineStart += line.width + len(window.buffer.newLineSeq)
	}
}

func (window *Window) drawSelection(screen tcell.Screen, a *WindowCursor, b *WindowCursor) {
	start := a
	end := b
	if a.index > b.index {
		start, end = end, start
	}
	lineStart := 0
	for y, line := range window.buffer.lines {
		if y >= start.row && y <= end.row {
			for x := 0; x < line.width; x++ {
				i := lineStart + x
				if i != window.cursor.index && start.index <= i && i <= end.index {
					char := rune(window.buffer.content[i])
					style := tcell.StyleDefault.Reverse(true)
					screen.SetContent(x, y-window.topLine, char, nil, style)
				}
			}
			// window.DEBUGdrawRN(screen, lineStart, y, line)
		}
		lineStart += line.width + len(window.buffer.newLineSeq)
	}
}

func (window *Window) DEBUGdrawRN(screen tcell.Screen, lineStart int, y int, line BufferLine) {
	for x := line.width; x < line.width+len(window.buffer.newLineSeq); x++ {
		i := lineStart + x
		if window.buffer.content[i] == '\r' {
			screen.SetContent(x, y-window.topLine, 'R', nil, tcell.StyleDefault)
		}
		if window.buffer.content[i] == '\n' {
			screen.SetContent(x, y-window.topLine, 'N', nil, tcell.StyleDefault)
		}
	}
}

func (window *Window) switchToInsert() {
	window.mode = InsertMode
}
func (window *Window) switchToNormal() {
	window.mode = NormalMode
}

func (window *Window) switchToVisual() {
	window.mode = VisualMode
	*window.secondCursor = *window.cursor
}

func (window *Window) cursorRight() {
	lineWidth := window.buffer.lines[window.cursor.row].width
	if lineWidth == 0 {
		return
	}
	if window.mode == NormalMode && window.cursor.column+1 == lineWidth {
		return
	}
	if window.mode == InsertMode && window.cursor.column == lineWidth {
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

	window.cursor.index -= window.cursor.column - 1
	window.cursor.index += thisLineWidth + len(window.buffer.newLineSeq)
	window.cursor.column = max(min(window.cursor.originColumn, nextLineWidth), 0)
	window.cursor.row += 1
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
	window.cursor.index += window.cursor.column - 1

	window.topLine = min(window.topLine, window.cursor.row)
}

func (window *Window) insert(value byte) {
	window.buffer.content = append(
		window.buffer.content[:window.cursor.index+1],
		window.buffer.content[window.cursor.index:]...,
	)
	window.buffer.content[window.cursor.index] = value
	window.buffer.lines[window.cursor.row].width++
}

func (window *Window) remove() {
	if window.cursor.index == 0 {
		return
	}
	if window.cursor.column == 0 {
		thisLineWidth := window.buffer.lines[window.cursor.row].width
		prevLineWidth := window.buffer.lines[window.cursor.row-1].width
		newLineSeqLen := len(window.buffer.newLineSeq)

		window.buffer.content = append(
			window.buffer.content[:window.cursor.index-newLineSeqLen],
			window.buffer.content[window.cursor.index:]...,
		)
		window.buffer.lines[window.cursor.row-1].width += thisLineWidth
		window.buffer.lines = append(
			window.buffer.lines[:window.cursor.row],
			window.buffer.lines[window.cursor.row+1:]...,
		)

		window.cursor.row--
		window.cursor.column = prevLineWidth
		window.cursor.index -= newLineSeqLen
		window.cursor.originColumn = window.cursor.column
	} else {
		window.buffer.content = append(
			window.buffer.content[:window.cursor.index-1],
			window.buffer.content[window.cursor.index:]...,
		)
		window.buffer.lines[window.cursor.row].width--
		window.cursorLeft()

	}
}

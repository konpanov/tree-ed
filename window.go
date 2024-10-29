package main

import (
	"github.com/gdamore/tcell/v2"
)

type WindowMode int

const (
	NormalMode     WindowMode = iota
	InsertMode     WindowMode = iota
	VisualMode     WindowMode = iota
	VisualTreeMode WindowMode = iota
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
	windowStartIndex := 0
	for _, line := range window.buffer.lines[:window.topLine] {
		windowStartIndex += line.width + len(window.buffer.newLineSeq)
	}

	switch window.mode {
	case NormalMode:
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
		screen.ShowCursor(window.cursor.column, window.cursor.row-window.topLine)
		window.drawText(screen, windowStartIndex)
	case VisualMode:
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
		screen.HideCursor()
		window.drawText(screen, windowStartIndex)
		window.drawSelection(screen, windowStartIndex, window.cursor.index, window.secondCursor.index)
	case InsertMode:
		screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
		screen.ShowCursor(window.cursor.column, window.cursor.row-window.topLine)
		window.drawText(screen, windowStartIndex)
	}
}

func (window *Window) drawText(screen tcell.Screen, i int) {
	visibleLines := window.buffer.lines[window.topLine : window.topLine+window.height]
	for y, line := range visibleLines {
		for x := 0; x < line.width; x++ {
			char := rune(window.buffer.content[i])
			style := tcell.StyleDefault
			screen.SetContent(x, y, char, nil, style)
			i++
		}
		// window.DEBUGdrawRN(screen, i, y, line)
		i += len(window.buffer.newLineSeq)
	}
}

func (window *Window) drawSelection(screen tcell.Screen, i int, a int, b int) {
	start, end := min(a, b), max(a, b)
	style := tcell.StyleDefault.Reverse(true)
	visibleLines := window.buffer.lines[window.topLine : window.topLine+window.height]
	for y, line := range visibleLines {
		if line.width == 0 && start <= i && i <= end {
			screen.SetContent(0, y, ' ', nil, style)
		}
		for x := 0; x < line.width; x++ {
			if start <= i && i <= end {
				char := rune(window.buffer.content[i])
				screen.SetContent(x, y, char, nil, style)
			}
			i++
		}
		i += len(window.buffer.newLineSeq)
	}
}

func (window *Window) DEBUGdrawRN(screen tcell.Screen, lineStart int, y int, line BufferLine) {
	for x := line.width; x < line.width+len(window.buffer.newLineSeq); x++ {
		i := lineStart + x
		if window.buffer.content[i] == '\r' {
			screen.SetContent(x, y, 'R', nil, tcell.StyleDefault)
		}
		if window.buffer.content[i] == '\n' {
			screen.SetContent(x, y, 'N', nil, tcell.StyleDefault)
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
	maxCol := lineWidth - 1
	if window.mode == InsertMode {
		maxCol = lineWidth
	}
	if window.cursor.column == maxCol {
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

package main

type WindowCursor struct {
	index        int
	row          int
	column       int
	originColumn int
}

type Window struct {
	buffer  *Buffer
	cursor  *WindowCursor
	topLine int // TODO: Should it be in WindowCursor?
}

func windowFromBuffer(buffer *Buffer) *Window {
	return &Window{
		buffer:  buffer,
		cursor:  &WindowCursor{0, 0, 0, 0},
		topLine: 0,
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
}

func (window *Window) cursorUp() {
	if window.cursor.row == 0 {
		return
	}

	prevLineWidth := window.buffer.lines[window.cursor.row-1].width - 1

	window.cursor.index -= window.cursor.column
	window.cursor.row -= 1
	window.cursor.index -= prevLineWidth + len(window.buffer.newLineSeq)
	window.cursor.column = max(min(window.cursor.originColumn, prevLineWidth), 0)
	window.cursor.index += window.cursor.column
}

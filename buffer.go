package main

import (
	"os"
)

type BufferLine struct {
	width int
}

type BufferCursor struct {
	index        int
	row          int
	column       int
	originColumn int
}

type Buffer struct {
	content    []byte
	newLineSeq []byte
	lines      []BufferLine
	cursor     BufferCursor
	quiting    bool
}

func bufferFromFile(filename string, newLineSeq []byte) *Buffer {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	buffer := &Buffer{
		content:    content,
		newLineSeq: newLineSeq,
		lines:      []BufferLine{},
		cursor:     BufferCursor{0, 0, 0, 0},
		quiting:    false,
	}

	prevNewLine := 0
	for i := 0; i < len(content); i++ {
		if matchBytes(content[i:], newLineSeq) {
			line := BufferLine{width: i - prevNewLine}
			buffer.lines = append(buffer.lines, line)
			prevNewLine = i + len(newLineSeq)
		}
	}
	line := BufferLine{width: len(content) - prevNewLine}
	buffer.lines = append(buffer.lines, line)

	return buffer
}

func (b *Buffer) cursorRight() {
	if b.cursor.column+1 == b.lines[b.cursor.row].width {
		return
	}
	b.cursor.index++
	b.cursor.column++
	b.cursor.originColumn = b.cursor.column
}

func (b *Buffer) cursorLeft() {
	if b.cursor.column == 0 {
		return
	}
	b.cursor.index--
	b.cursor.column--
	b.cursor.originColumn = b.cursor.column
}

func (b *Buffer) cursorDown() {
	if b.cursor.row+1 == len(b.lines) {
		return
	}
	b.cursor.index -= b.cursor.column
	b.cursor.index += b.lines[b.cursor.row].width + len(b.newLineSeq)
	b.cursor.row += 1
	b.cursor.column = max(min(b.cursor.originColumn, b.lines[b.cursor.row].width-1), 0)
	b.cursor.index += b.cursor.column
}

func (b *Buffer) cursorUp() {
	if b.cursor.row == 0 {
		return
	}
	b.cursor.index -= b.cursor.column
	b.cursor.row -= 1
	b.cursor.index -= b.lines[b.cursor.row].width + len(b.newLineSeq)
	b.cursor.column = max(min(b.cursor.originColumn, b.lines[b.cursor.row].width-1), 0)
	b.cursor.index += b.cursor.column
}


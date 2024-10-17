package main

import (
	"os"
)

type BufferLine struct {
	width int
}

type Buffer struct {
	content    []byte
	newLineSeq []byte
	lines      []BufferLine
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

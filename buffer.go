package main

import (
	"os"
)

type BufferLine struct {
	start, end int
}

type Buffer struct {
	content    []byte
	newLineSeq []byte
	lines      []BufferLine
	quiting    bool
}

func bufferFromContent(content []byte, newLineSeq []byte) *Buffer {
	buffer := &Buffer{
		content:    content,
		newLineSeq: newLineSeq,
		lines:      []BufferLine{},
		quiting:    false,
	}
	buffer.update()
	return buffer
}

func bufferFromFile(filename string, newLineSeq []byte) *Buffer {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return bufferFromContent(content, newLineSeq)
}

func (buffer *Buffer) update() {
	buffer.lines = []BufferLine{}
	prevNewLine := 0
	for i := 0; i < len(buffer.content); i++ {
		if matchBytes(buffer.content[i:], buffer.newLineSeq) {
			line := BufferLine{start: prevNewLine, end: i}
			buffer.lines = append(buffer.lines, line)
			prevNewLine = i + len(buffer.newLineSeq)
		}
	}
	line := BufferLine{start: prevNewLine, end: len(buffer.content)}
	buffer.lines = append(buffer.lines, line)
}

func (buffer *Buffer) insert(index int, value []byte) {
	if index == len(buffer.content) {
		buffer.content = append(buffer.content, value...)
	} else {
		before := buffer.content[:index]
		after := buffer.content[index:]
		buffer.content = append(before, append(value, after...)...)
	}
	buffer.update()
}

func (buffer *Buffer) erease(from int, to int) {
	buffer.content = append(buffer.content[:from-1], buffer.content[to:]...)
	buffer.update()
}

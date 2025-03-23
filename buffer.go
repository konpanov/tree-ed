package main

import (
	"context"
	"fmt"
	"log"
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type Range struct {
	start, end int
}

type Point struct {
	row, col int
}

type IBuffer interface {
	Insert(index int, value []byte) error
	Erase(r Range) error
	EraseLine(line_number int) error
	Coord(index int) (Point, error)
	Lines() []Range
}

var ErrIndexLessThanZero = fmt.Errorf("index cannot be less than zero")
var ErrIndexGreaterThanBufferSize = fmt.Errorf("index cannot be greater than buffer size")
var ErrLineIndexOutOfRange = fmt.Errorf("line index is negative or greater than or equal to number of lines")

type BufferNew struct {
	content []byte
	nl_seq  []byte
	tree    *sitter.Tree
	quiting bool
}

func bufferNewFromContent(content []byte, nl_seq []byte) (*BufferNew, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		log.Fatalln("Failed to parse buffer")
		return nil, err
	}

	buffer := &BufferNew{
		content: content,
		nl_seq:  nl_seq,
		tree:    tree,
	}

	return buffer, nil
}

func (b *BufferNew) Insert(index int, value []byte) error {
	err := b.check_index(index)
	if err != nil {
		return err
	}

	b.content = slices.Insert(b.content, index, value...)
	return nil
}

func (b *BufferNew) EraseLine(line_number int) error {
	lines := b.Lines()
	if line_number < 0 || len(lines) <= line_number {
		return ErrLineIndexOutOfRange
	}
	line := lines[line_number]
	line.end += len(b.nl_seq)
	b.Erase(line)
	return nil
}

func (b *BufferNew) Erase(r Range) error {
	var err error

	err = b.check_index(r.start)
	if err != nil {
		return err
	}

	err = b.check_index(r.end)
	if err != nil {
		return err
	}

	b.content = slices.Delete(b.content, r.start, r.end)
	return nil
}

func (b *BufferNew) Coord(index int) (Point, error) {
	var err error
	p := Point{0, 0}

	err = b.check_index(index)
	if err != nil {
		return p, err
	}

	for i := 0; i < index; {
		if matchBytes(b.content[i:], b.nl_seq) {
			i += len(b.nl_seq)
			p.row++
			p.col = 0
		} else {
			p.col++
			i++
		}
	}

	return p, nil
}

func (b *BufferNew) Lines() []Range {
	lines := []Range{}
	lines = append(lines, Range{0, 0})
	for i := 0; i < len(b.content); {
		if matchBytes(b.content[i:], b.nl_seq) {
			lines[len(lines)-1].end = i
			i += len(b.nl_seq)
			lines = append(lines, Range{start: i, end: i})
		} else {
			i += 1
		}
	}
	lines[len(lines)-1].end = len(b.content)
	return lines
}

func (b *BufferNew) check_index(index int) error {
	if index < 0 {
		return ErrIndexLessThanZero
	}
	if index > len(b.content) {
		return ErrIndexGreaterThanBufferSize
	}
	return nil

}

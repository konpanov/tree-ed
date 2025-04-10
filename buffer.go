package main

import (
	"context"
	"fmt"
	"log"
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// start and end specify "fenceposts" in between characters.
// |h|e|l|l|o|
// ^         ^
// start     end
// In other words start is inclusive and end is not
type Region struct {
	start, end int
}

func (r Region) Start() int {
	return min(r.start, r.end)
}
func (r Region) End() int {
	return max(r.start, r.end)
}

type Size struct {
	width, height int
}
type Point struct {
	row, col int
}

type IBuffer interface {
	Content() []byte
	Nl_seq() []byte
	IsQuiting() bool
	SetQuiting(v bool)
	CheckIndex(index int) error

	Insert(index int, value []byte) error
	Erase(r Region) error
	EraseLine(line_number int) error
	Coord(index int) (Point, error)
	RuneCoord(index int) (Point, error)

	Tree() *sitter.Tree

	// Returns ranges in which lines are contained, without the new line sequences.
	// New lines must be left out to treat the same last lines with new lines and without.
	Lines() []Region
}

var ErrIndexLessThanZero = fmt.Errorf("index cannot be less than zero")
var ErrIndexGreaterThanBufferSize = fmt.Errorf("index cannot be greater than buffer size")
var ErrLineIndexOutOfRange = fmt.Errorf("line index is negative or greater than or equal to number of lines")

type Buffer struct {
	content []byte
	nl_seq  []byte
	tree    *sitter.Tree
	quiting bool
}

func bufferFromContent(content []byte, nl_seq []byte) (*Buffer, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		log.Fatalln("Failed to parse buffer")
		return nil, err
	}

	buffer := &Buffer{
		content: content,
		nl_seq:  nl_seq,
		tree:    tree,
	}

	return buffer, nil
}

func (b *Buffer) Content() []byte {
	return b.content
}

func (b *Buffer) IsQuiting() bool {
	return b.quiting
}

func (b *Buffer) SetQuiting(v bool) {
	b.quiting = v
}

func (b *Buffer) Insert(index int, value []byte) error {
	err := b.CheckIndex(index)
	if err != nil {
		return err
	}

	b.content = slices.Insert(b.content, index, value...)
	return nil
}

func (b *Buffer) EraseLine(line_number int) error {
	lines := b.Lines()
	if line_number < 0 || len(lines) <= line_number {
		return ErrLineIndexOutOfRange
	}
	line := lines[line_number]
	line.end += len(b.nl_seq)
	b.Erase(line)
	return nil
}

func (b *Buffer) Erase(r Region) error {
	var err error

	err = b.CheckIndex(r.start)
	if err != nil {
		return err
	}

	err = b.CheckIndex(r.end)
	if err != nil {
		return err
	}

	b.content = slices.Delete(b.content, r.start, r.end)
	return nil
}

func (b *Buffer) Coord(index int) (Point, error) {
	var err error
	p := Point{0, 0}

	err = b.CheckIndex(index)
	if err != nil {
		log.Fatalln("Could not find cursor index: ", err)
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

func (b *Buffer) RuneCoord(index int) (Point, error) {
	var err error
	p := Point{0, 0}

	err = b.CheckIndex(index)
	if err != nil {
		log.Fatalln("Could not find cursor index: ", err)
		return p, err
	}

	for row, line := range b.Lines() {
		if line.start <= index && index <= line.end {
			col := len([]rune(string(b.content[line.start:index])))
			return Point{row: row, col: col}, nil
		}
	}

	return p, nil
}

func (b *Buffer) Lines() []Region {
	lines := []Region{}
	line_finished := false
	lines = append(lines, Region{0, 0})
	content := b.content
	for i := 0; i < len(content); {
		if line_finished {
			lines = append(lines, Region{i, i})
			line_finished = false
		}
		if matchBytes(content[i:], b.nl_seq) {
			lines[len(lines)-1].end = i
			i += len(b.nl_seq)
			line_finished = true
		} else {
			i += 1
		}
	}
	if !line_finished {
		lines[len(lines)-1].end = len(b.content)
	}
	return lines
}

func (b *Buffer) CheckIndex(index int) error {
	if index < 0 {
		return ErrIndexLessThanZero
	}
	if index > len(b.content) {
		return ErrIndexGreaterThanBufferSize
	}
	return nil

}

func (b *Buffer) Tree() *sitter.Tree {
	return b.tree
}

func (b *Buffer) Nl_seq() []byte {
	return b.nl_seq
}

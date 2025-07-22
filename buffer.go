package main

import (
	"fmt"
	"log"
	"slices"
	"unicode/utf8"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// start and end specify "fenceposts" in between characters.
// |h|e|l|l|o|
// ^         ^
// start     end
// In other words start is inclusive and end is not
type Region struct {
	start, end int
}

func NewRegion(a, b int) Region {
	return Region{
		start: min(a, b),
		end:   max(a, b),
	}
}

type Point struct {
	row, col int
}

func sitterPoint(p Point) sitter.Point {
	return sitter.Point{Row: uint(p.row), Column: uint(p.col)}
}

func (self Point) Add(other Point) Point {
	return Point{row: self.row + other.row, col: self.col + other.col}
}

// TODO Add filename field for file buffers that return nil for non file buffers
type IBuffer interface {
	Content() []byte
	Nl_seq() []byte
	IsQuiting() bool
	SetQuiting(v bool)
	CheckIndex(index int) error
	CheckLine(line int) error
	Row(index int) (int, error)

	// Modifications
	Edit(input ReplacementInput) error

	Coord(index int) (Point, error)
	RuneCoord(index int) (Point, error)

	// If point lies after the line end the index of the start of the next
	// line or the eof index is given.
	IndexFromRuneCoord(p Point) (int, error)

	Tree() *sitter.Tree

	// Returns ranges in which lines are contained, without the new line sequences.
	// New lines must be left out to treat the same last lines with new lines and without.
	Lines() []Region
	Line(line int) (Region, error)
	IsNewLine(index int) (bool, error)
	LastIndex() int

	Close()
}

var ErrIndexLessThanZero = fmt.Errorf("index cannot be less than zero")
var ErrIndexGreaterThanBufferSize = fmt.Errorf("index cannot be greater than buffer size")
var ErrLineIndexOutOfRange = fmt.Errorf("line index is negative or greater than or equal to number of lines")
var ErrCoordOutOfRange = fmt.Errorf("Coordinate does not exist in the buffer")
var ErrUnexpected = fmt.Errorf("An unexpected error occured")

type ReplacementInput struct {
	start       int
	end         int
	replacement []byte
}

type Buffer struct {
	content     []byte
	nl_seq      []byte
	tree_parser *sitter.Parser
	tree        *sitter.Tree
	quiting     bool
	lines       []Region
}

func NewEmptyBuffer(nl_seq []byte, parser *sitter.Parser) (*Buffer, error) {
	content := []byte{}
	var tree *sitter.Tree
	if parser != nil {
		tree = parser.Parse(content, nil)
	}

	buffer := &Buffer{
		content:     content,
		nl_seq:      nl_seq,
		tree_parser: parser,
		tree:        tree,
		lines:       []Region{{0, 0}},
	}

	return buffer, nil

}

func bufferFromContent(content []byte, nl_seq []byte, parser *sitter.Parser) (*Buffer, error) {
	buffer, err := NewEmptyBuffer(nl_seq, parser)
	panic_if_error(err)
	err = buffer.Edit(ReplacementInput{0, 0, content})
	panic_if_error(err)
	return buffer, nil
}

func (b *Buffer) Close() {
	if b.tree != nil {
		log.Println("Closing tree")
		b.tree.Close()
	}
	if b.tree_parser != nil {
		log.Println("Closing tree parser")
		b.tree_parser.Close()
	}
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

func (b *Buffer) Edit(input ReplacementInput) error {
	var err error
	if err = b.CheckIndex(input.start); err != nil {
		return err
	}

	if err = b.CheckIndex(input.end); err != nil {
		return err
	}

	start_point, err := b.Coord(input.start)
	panic_if_error(err)
	end_point, err := b.Coord(input.end)
	panic_if_error(err)

	b.content = slices.Replace(b.content, input.start, input.end, input.replacement...)
	b.lines = b.calculateLines()

	new_end := input.start + len(input.replacement)
	new_end_point, _ := b.Coord(new_end)

	if b.tree_parser != nil {
		b.tree.Edit(&sitter.InputEdit{
			StartByte:      uint(input.start),
			OldEndByte:     uint(input.end),
			NewEndByte:     uint(new_end),
			StartPosition:  sitterPoint(start_point),
			OldEndPosition: sitterPoint(end_point),
			NewEndPosition: sitterPoint(new_end_point),
		})
		b.tree = b.tree_parser.Parse(b.Content(), nil)
		panic_if_error(err)
	}
	return nil
}

func (b *Buffer) Row(index int) (int, error) {
	if err := b.CheckIndex(index); err != nil {
		return 0, err
	}
	for i, line := range b.Lines() {
		if line.start <= index && index < line.end+len(b.Nl_seq()) {
			return i, nil
		}
	}
	if index == len(b.Content()) {
		return len(b.Lines()) - 1, nil
	}
	log.Panicf("Could not find index that is in buffer range.\n Index: %d\n Buffer %+v", index, b)
	return 0, ErrUnexpected
}

func (b *Buffer) Coord(index int) (Point, error) {
	row, err := b.Row(index)
	if err != nil {
		return Point{}, err
	}
	line, err := b.Line(row)
	panic_if_error(err)
	return Point{row: row, col: index - line.start}, nil
}

func (b *Buffer) RuneCoord(index int) (Point, error) {
	row, err := b.Row(index)
	if err != nil {
		return Point{}, err
	}
	line, err := b.Line(row)
	panic_if_error(err)
	return Point{row: row, col: utf8.RuneCount(b.Content()[line.start:index])}, nil
}

func (b *Buffer) IndexFromRuneCoord(p Point) (int, error) {
	if err := b.CheckLine(p.row); err != nil {
		return 0, err
	}
	lines := b.Lines()
	line := lines[p.row]
	in_runes := []rune(string(b.Content()[line.start:line.end]))
	if p.col < 0 {
		return 0, fmt.Errorf("%w: coord col cannot be negative (%d)", ErrCoordOutOfRange, p.col)
	} else if p.col > len(in_runes) {
		next_line_row := p.row + 1
		if next_line_row != len(lines) {
			return lines[next_line_row].start, nil
		} else {
			return len(b.Content()), nil
		}
	} else {
		line_len_before_coord_in_bytes := len(string(in_runes[:p.col]))
		return lines[p.row].start + line_len_before_coord_in_bytes, nil
	}
}

func (b *Buffer) calculateLines() []Region {
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

func (b *Buffer) Lines() []Region {
	return b.lines
}

func (b *Buffer) Line(line int) (Region, error) {
	if err := b.CheckLine(line); err != nil {
		return Region{}, err
	}
	return b.Lines()[line], nil
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

func (b *Buffer) CheckLine(line int) error {
	if line < 0 {
		return fmt.Errorf("%w: coord row cannot be negative (%d)", ErrCoordOutOfRange, line)
	}
	lines := b.Lines()
	if line >= len(lines) {
		return fmt.Errorf("%w: coord row cannot be greater than the number of lines (%d > %d)", ErrCoordOutOfRange, line, len(lines))
	}
	return nil
}

func (b *Buffer) Tree() *sitter.Tree {
	return b.tree
}

func (b *Buffer) Nl_seq() []byte {
	return b.nl_seq
}

func (b *Buffer) LastIndex() int {
	line, err := b.Line(len(b.Lines()) - 1)
	panic_if_error(err)
	last_index := max(line.start, line.end-1)
	return last_index
}

func (b *Buffer) IsNewLine(index int) (bool, error) {
	row, err := b.Row(index)
	if err != nil {
		return false, err
	}
	line, err := b.Line(row)
	if err != nil {
		return false, err
	}
	return index == line.end, nil

}

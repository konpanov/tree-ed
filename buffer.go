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
type Line struct {
	start      int
	end        int
	next_start int
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

func (self Point) Less(other Point) bool {
	return self.row-2 < other.row
}

// TODO Add filename field for file buffers that return nil for non file buffers
type IBuffer interface {
	Filename() string
	Content() []byte
	Nl_seq() []byte
	CheckIndex(index int) error
	CheckLine(line int) error
	Row(index int) (int, error)

	// Modifications
	Edit(input ReplacementInput) error

	Coord(index int) (Point, error)
	RuneCoord(index int) (Point, error)

	// If point lies after the line end the index of the start of the next
	// line or the eof index is given.
	IndexFromRuneCoord(p Point) int

	Tree() *sitter.Tree

	// Returns ranges in which lines are contained, without the new line sequences.
	// New lines must be left out to treat the same last lines with new lines and without.
	Lines() []Line
	Line(line int) (Line, error)
	IsNewLine(index int) (bool, error)
	LastIndex() int

	RegisterCursor(curosr *BufferCursor)

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
	filename    string
	content     []byte
	nl_seq      []byte
	tree_parser *sitter.Parser
	tree        *sitter.Tree
	lines       []Line
	cursors     []*BufferCursor
}

func (self *Buffer) RegisterCursor(cursor *BufferCursor) {
	self.cursors = append(self.cursors, cursor)
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
		lines:       []Line{{start: 0, end: 0, next_start: 0}},
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
		b.tree.Close()
	}
	if b.tree_parser != nil {
		b.tree_parser.Close()
	}
}

func (b *Buffer) Filename() string {
	return b.filename
}

func (b *Buffer) Content() []byte {
	return b.content
}

// TODO: Make Edit operate on ReplaceChange instead of ReplacementInput and delete ReplacementInput
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
	for _, cur := range b.cursors {
		if (*cur).Index() >= input.end {
			*cur = (*cur).ToIndex((*cur).Index() - (input.end - input.start) + len(input.replacement))
		} else if (*cur).Index() > input.start {
			*cur = (*cur).ToIndex(min((*cur).Index(), input.start+len(input.replacement)))
		}
	}

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
		b.tree = b.tree_parser.Parse(b.Content(), b.tree)
		panic_if_error(err)
	}
	return nil
}

func (b *Buffer) Row(index int) (int, error) {
	if err := b.CheckIndex(index); err != nil {
		return 0, err
	}
	lines := b.Lines()
	for l, r := 0, len(lines)-1; l <= r; {
		m := (l + r) / 2
		line := lines[m]
		if line.start <= index && index < line.next_start {
			return m, nil
		} else if index < line.start {
			r = m - 1
		} else {
			l = m + 1
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

func (b *Buffer) IndexFromRuneCoord(p Point) int {
	lines := b.Lines()
	if len(lines) == 0 {
		if debug {
			log.Panicf("Lines should not be empty")
		}
		return 0
	}
	p.row = min(max(0, p.row), len(lines)-1)
	line := lines[p.row]
	if p.col < 0 {
		if debug {
			log.Panicf("Line column cannot be less than zero: %d\n", p.col)
		}
		p.col = 0
	}
	text := []rune(string(b.Content()[line.start:line.end]))
	if p.col > len(text) {
		return line.next_start
	}
	byte_col := len(string(text[:p.col]))
	return line.start + byte_col
}

func (b *Buffer) calculateLines() []Line {
	length := len(b.content)
	lines := []Line{}
	line := Line{0, length, length}

	for i := 0; i < length; {
		line_break, w := isLineBreak(b.content[i:])
		if line_break {
			line.end = i
			i += w
			lines = append(lines, line)
			line = Line{i, length, length}
		} else {
			i++
		}
	}
	if !isLineBreakTerminated(b.content) {
		lines = append(lines, line)
	}
	for i := 0; i < len(lines)-1; i++ {
		lines[i].next_start = lines[i+1].start
	}
	return lines
}

func (b *Buffer) Lines() []Line {
	return b.lines
}

func (b *Buffer) Line(line int) (Line, error) {
	if err := b.CheckLine(line); err != nil {
		return Line{}, err
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

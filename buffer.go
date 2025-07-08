package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"unicode/utf8"

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
	return sitter.Point{Row: uint32(p.row), Column: uint32(p.col)}
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

	// Modifications
	Insert(index int, value []byte) error
	EraseRune(index int) error
	EraseLine(line_number int) error
	Erase(r Region) error
	Edit(input ReplacementInput) error

	Coord(index int) (Point, error)
	RuneCoord(index int) (Point, error)

	// If point lies after the line end the index of the start of the next
	// line or the eof index is give. 
	IndexFromRuneCoord(p Point) (int, error)

	Changes() []BufferChange
	ChangeIndex() int
	Undo() error
	Redo() error

	Tree() *sitter.Tree

	// Returns ranges in which lines are contained, without the new line sequences.
	// New lines must be left out to treat the same last lines with new lines and without.
	Lines() []Region
}

var ErrIndexLessThanZero = fmt.Errorf("index cannot be less than zero")
var ErrIndexGreaterThanBufferSize = fmt.Errorf("index cannot be greater than buffer size")
var ErrLineIndexOutOfRange = fmt.Errorf("line index is negative or greater than or equal to number of lines")
var ErrCoordOutOfRange = fmt.Errorf("Coordinate does not exist in the buffer")
var ErrChangeIndexOutOfRange = fmt.Errorf("Change index does not exist")
var ErrChangesAreNotContinuous = fmt.Errorf("Changes are not continuoues")
var ErrNoChangesToUndo = fmt.Errorf("There are no changes to undo")
var ErrNoChangesToRedo = fmt.Errorf("There are no changes to redo")

type ReplacementInput struct {
	start       int
	end         int
	replacement []byte
}

type Buffer struct {
	content      []byte
	nl_seq       []byte
	tree_parser  *sitter.Parser
	tree         *sitter.Tree
	quiting      bool
	change_index int
	changes      []BufferChange
}

func NewEmptyBuffer(nl_seq []byte) (*Buffer, error) {
	content := []byte{}
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		log.Fatalln("Failed to parse empty buffer")
		return nil, err
	}

	buffer := &Buffer{
		content:      content,
		nl_seq:       nl_seq,
		tree_parser:  parser,
		tree:         tree,
		change_index: 0,
		changes:      []BufferChange{},
	}

	return buffer, nil

}

func bufferFromContent(content []byte, nl_seq []byte) (*Buffer, error) {
	buffer, err := NewEmptyBuffer(nl_seq)
	panic_if_error(err)
	err = buffer.Edit(ReplacementInput{0, 0, content})
	panic_if_error(err)
	buffer.change_index = 0
	buffer.changes = []BufferChange{}
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
	return b.Edit(ReplacementInput{index, index, value})
}

func (b *Buffer) EraseRune(index int) error {
	text := b.Content()[index:]
	_, length := utf8.DecodeRune(text)
	return b.Erase(NewRegion(index, index+length))
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
	return b.Edit(ReplacementInput{r.start, r.end, nil})
}

func (b *Buffer) Edit(input ReplacementInput) error {
	start := input.start
	end := input.end
	after := input.replacement
	if err := b.CheckIndex(start); err != nil {
		return err
	}

	if err := b.CheckIndex(end); err != nil {
		return err
	}

	start_point, _ := b.Coord(start)
	end_point, _ := b.Coord(end)

	_before := b.content[start:end]
	before := make([]byte, len(_before))
	copy(before, _before)

	b.content = slices.Replace(b.content, start, end, after...)

	new_end := start + len(after)
	new_end_point, _ := b.Coord(new_end)

	change := BufferChange{
		start_index:   start,
		new_end_index: new_end,
		old_end_index: end,
		start_pos:     start_point,
		new_end_pos:   new_end_point,
		old_end_pos:   end_point,
		before:        before,
		after:         after,
	}
	b.UpdateTree(change)
	b.changes = b.changes[:b.change_index]
	b.changes = append(b.changes, change)
	b.change_index++
	return nil
}

func (b *Buffer) UpdateTree(change BufferChange) {
	edit := sitter.EditInput{}
	edit.StartIndex = uint32(change.start_index)
	edit.OldEndIndex = uint32(change.old_end_index)
	edit.NewEndIndex = uint32(change.new_end_index)
	edit.StartPoint = sitterPoint(change.start_pos)
	edit.OldEndPoint = sitterPoint(change.old_end_pos)
	edit.NewEndPoint = sitterPoint(change.new_end_pos)
	b.tree.Edit(edit)
	var err error
	b.tree, err = b.tree_parser.ParseCtx(context.Background(), b.tree, b.Content())
	panic_if_error(err)
}

func (b *Buffer) Undo() error {
	if b.change_index == 0 {
		return ErrNoChangesToUndo
	}
	if len(b.changes) < b.change_index {
		log.Panicln("Change index is higher than number of changes which should not happen")
	}
	if b.change_index < 0 {
		log.Panicln("Change index is negative, which should no happen")
	}

	change := b.changes[b.change_index-1]
	b.content = change.Undo(b.content)
	b.UpdateTree(change.Reverse())
	b.change_index--
	return nil
}

func (b *Buffer) Redo() error {
	if b.change_index == len(b.changes) {
		return ErrNoChangesToUndo
	}
	if b.change_index < 0 {
		log.Panicln("Change index is negative, which should no happen")
	}

	change := b.changes[b.change_index]
	b.content = change.Reverse().Undo(b.content)
	b.UpdateTree(change)
	b.change_index++
	return nil
}

func (b *Buffer) Changes() []BufferChange {
	return b.changes
}

func (b *Buffer) ChangeIndex() int {
	return b.change_index
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
		next_line_row := p.row+1
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

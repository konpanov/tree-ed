package main

import (
	"fmt"
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

type Pos struct {
	row, col int
}

func sitterPoint(p Pos) sitter.Point {
	return sitter.Point{Row: uint(p.row), Column: uint(p.col)}
}

type IBuffer interface {
	Filename() string
	Content() []byte
	Length() int
	LineBreak() []byte
	Row(index int) int
	Edit(input ReplacementInput) error
	BytePos(index int) Pos
	RunePos(index int) Pos
	Index(p Pos) int

	Tree() *sitter.Tree
	Lines() []Line
	RegisterCursor(curosr *BufferCursor)
	Close()
}

var ErrIndexLessThanZero = fmt.Errorf("index cannot be less than zero")
var ErrIndexGreaterThanBufferSize = fmt.Errorf("index cannot be greater than buffer size")

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
	err := b.checkIndex(input.start)
	if err == nil {
		err = b.checkIndex(input.end)
	}
	if err != nil {
		return err
	}

	sitter_input := &sitter.InputEdit{}
	sitter_input.StartByte = uint(input.start)
	sitter_input.OldEndByte = uint(input.end)
	sitter_input.StartPosition = sitterPoint(b.BytePos(input.start))
	sitter_input.OldEndPosition = sitterPoint(b.BytePos(input.end))

	b.content = slices.Replace(b.content, input.start, input.end, input.replacement...)
	b.lines = b.calculateLines(input)
	for _, cur := range b.cursors {
		*cur = (*cur).Update(input)
	}

	new_end := input.start + len(input.replacement)
	sitter_input.OldEndByte = uint(new_end)
	sitter_input.StartPosition = sitterPoint(b.BytePos(new_end))

	if b.tree_parser != nil {
		b.tree.Edit(sitter_input)
		b.tree = b.tree_parser.Parse(b.Content(), b.tree)
	}
	return nil
}

func (b *Buffer) Row(index int) int {
	if index < 0 {
		return 0
	}
	lines := b.Lines()
	for l, r := 0, len(lines)-1; l <= r; {
		m := (l + r) / 2
		if index < lines[m].start {
			r = m - 1
		} else if index >= lines[m].next_start {
			l = m + 1
		} else {
			return m
		}

	}
	return len(lines) - 1
}

func (b *Buffer) BytePos(index int) Pos {
	row := b.Row(index)
	line := b.Lines()[row]
	return Pos{row: row, col: index - line.start}
}

func (b *Buffer) RunePos(index int) Pos {
	row := b.Row(index)
	line := b.Lines()[row]
	line_text := b.Content()[line.start:index]
	col := utf8.RuneCount(line_text)
	return Pos{row: row, col: col}
}

func (b *Buffer) Index(p Pos) int {
	lines := b.Lines()
	row := clip(p.row, 0, len(lines)-1)
	line := b.Lines()[row]
	line_text := b.Content()[line.start:line.end]
	line_runes := []rune(string(line_text))
	if p.col > len(line_runes) {
		return line.next_start
	}
	byte_col := len(string(line_runes[:p.col]))
	return line.start + byte_col
}

func (b *Buffer) calculateLines(input ReplacementInput) []Line {
	length := len(b.content)
	row := b.Row(input.start)
	lines := b.lines[:row]
	line := Line{b.lines[row].start, length, length}

	for i := line.start; i < length; {
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
	empty := len(b.lines) == 0
	assert(!empty, "Lines should not be empty")
	return b.lines
}

func (b *Buffer) Tree() *sitter.Tree {
	return b.tree
}

func (b *Buffer) LineBreak() []byte {
	return b.nl_seq
}

func (b *Buffer) Length() int {
	return len(b.content)
}

func (b *Buffer) checkIndex(index int) error {
	if index < 0 {
		return ErrIndexLessThanZero
	}
	if index > len(b.content) {
		return ErrIndexGreaterThanBufferSize
	}
	return nil
}

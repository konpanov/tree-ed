package main

import (
	"log"
	"math"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type WindowMode string

const (
	NormalMode WindowMode = "Normal"
	InsertMode WindowMode = "Insert"
	VisualMode WindowMode = "Visual"
	TreeMode   WindowMode = "Tree"
)

type Movement int

const (
	Right Movement = iota
	Left
	Up
	Down
	NodeUp
	NodeDown
	NodeLeft
	NodeRight
)

type WindowCursor struct {
	index               int
	row                 int
	col                 int
	originColumn        int
	invalidOriginColumn bool
}

func (c WindowCursor) log() {
	log.Printf("Cursor: %+v\n", c)
}

type Window struct {
	filename          string
	mode              WindowMode
	buffer            IBuffer
	cursor            BufferCursor
	anchorCursor      BufferCursor
	secondCursor      BufferCursor
	topLine           int
	leftColumn        int
	width, height     int
	numberColumnWidth int
	node              *sitter.Node
	parser            Parser
}

func windowFromBuffer(buffer IBuffer, width int, height int) *Window {
	return &Window{
		mode:              NormalMode,
		buffer:            buffer,
		cursor:            NewBufferCursor(buffer),
		anchorCursor:      NewBufferCursor(buffer),
		secondCursor:      NewBufferCursor(buffer),
		topLine:           0,
		leftColumn:        0,
		numberColumnWidth: int(max(math.Log10(float64(len(buffer.Lines()))))) + 2,
		width:             width,
		height:            height,
		node:              buffer.Tree().RootNode(),
		parser:            &NormalParser{},
	}
}

func (window *Window) Parse(ev tcell.Event) (Operation, error) {
	return window.parser.Parse(ev)
}

func (window *Window) switchToInsert() {
	window.mode = InsertMode
	window.parser = &InsertParser{}
}
func (window *Window) switchToNormal() {
	window.mode = NormalMode
	window.parser = &NormalParser{}
}

func (window *Window) switchToVisual() {
	window.mode = VisualMode
	window.secondCursor = window.cursor
	window.parser = &VisualParser{}
}

func (window *Window) switchToTree() {
	log.Println("Switch to tree mode")
	window.node = window.buffer.Tree().RootNode()
	window.mode = TreeMode
	window.parser = &TreeParser{}
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(int(window.node.EndByte()))
}

// Tree movements
func (window *Window) nodeUp() {
	if window.node.Equal(window.buffer.Tree().RootNode()) {
		return
	}
	window.node = window.node.Parent()
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(int(window.node.EndByte()))

}

func (window *Window) nodeDown() {
	if window.node.ChildCount() == 0 {
		return
	}
	window.node = window.node.Child(0)
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(int(window.node.EndByte()))
}

func (window *Window) nodeRight() {
	sibling := window.node.NextSibling()
	if sibling == nil {
		return
	}
	window.node = sibling
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(int(window.node.EndByte()))
}

func (window *Window) nodeLeft() {
	sibling := window.node.PrevSibling()
	if sibling == nil {
		return
	}
	window.node = sibling
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(int(window.node.EndByte()))
}

// Cursor movements
func (window *Window) cursorRight() {
	log.Println("Moving cursor to the right")
	log.Printf("Cursor at index: %d\n", window.cursor.index)
	if window.cursor.IsNewLine() {
		log.Println("Cursor at the end of the line. Cursor stays in place")
		return
	}
	next, err := window.cursor.RunesForward(1)
	if err != nil {
		log.Println("Cannot move cursor right: %w", err)
		return
	}
	if window.mode != InsertMode && (next.IsEnd() || next.IsNewLine()) {
		log.Println("Only in insert mode cursor can be at the end of the line")
		return
	}
	window.cursor = next
	window.anchorCursor = next
	log.Printf("Cursor moved to index: %d\n", window.cursor.Index())
}

func (window *Window) cursorLeft() {
	log.Println("Moving cursor to the left")
	log.Printf("Cursor at index: %d\n", window.cursor.index)

	prev, err := window.cursor.RunesBackward(1)
	if err != nil {
		log.Println("Cannot move cursor left: %w", err)
		return
	}
	line_start, err := window.cursor.SearchBackward(window.buffer.Nl_seq())
	if err != nil {
		line_start, _ = line_start.ToIndex(0)
	} else {
		line_start, _ = line_start.BytesForward(len(window.buffer.Nl_seq()))
	}
	if prev.index < line_start.index {
		log.Println("Cursor at the start of the line. Cursor stays in place")
		return
	}
	window.cursor = prev
	window.anchorCursor = prev
	log.Printf("Cursor moved to index: %d\n", window.cursor.index)
}

func (window *Window) cursorDown() {
	log.Println("Moving cursor down")
	log.Printf("Cursor at index: %d\n", window.cursor.index)

	lines := window.buffer.Lines()

	cursor_pos := window.cursor.RunePosition()
	if cursor_pos.row == len(window.buffer.Lines())-1 {
		log.Println("Cursor is already at the last line")
		return
	}
	next_line := lines[cursor_pos.row+1]
	next, err := window.cursor.ToIndex(next_line.start)
	if err != nil {
		log.Printf("Could not move cursor to the start of the next line")
		return
	}

	next, err = next.RunesForward(window.anchorCursor.RunePosition().col)
	if next.index >= next_line.end {
		next, err = next.ToIndex(max(next_line.end-1, next_line.start))
	}
	window.cursor = next
	log.Printf("Cursor moved to index: %d. Buffer length %d\n", window.cursor.index, len(window.buffer.Content()))
}

func (window *Window) cursorUp() {
	log.Println("Moving cursor up")
	log.Printf("Cursor at index: %d\n", window.cursor.index)

	lines := window.buffer.Lines()

	cursor_pos := window.cursor.RunePosition()
	if cursor_pos.row == 0 {
		log.Println("Cursor is already at the frist line")
		return
	}
	prev_line := lines[cursor_pos.row-1]
	prev, err := window.cursor.ToIndex(prev_line.start)
	if err != nil {
		log.Printf("Could not move cursor to the start of the prev line")
		return
	}

	prev, err = prev.RunesForward(window.anchorCursor.RunePosition().col)
	if prev.index >= prev_line.end {
		prev, err = prev.ToIndex(max(prev_line.end-1, prev_line.start))
	}
	window.cursor = prev
	log.Printf("Cursor moved to index: %d. Buffer length %d\n", window.cursor.index, len(window.buffer.Content()))

}

func (window *Window) insert(value []byte) {
	log.Printf("Inserting %c at index %d\n", value, window.cursor.index)
	window.buffer.Insert(window.cursor.index, value)
	var err error
	window.cursor, err = window.cursor.BytesForward(len(value))
	if err != nil {
		log.Printf("Could not move cursor after insert: %s", err)
	}
}

func (window *Window) remove() {
	log.Printf("Removing at index %d\n", window.cursor.index)

	// Do not remove from empty buffer
	if window.cursor.IsBegining() && window.cursor.IsEnd() {
		log.Printf("Cannot remove from empty buffer\n")
		return
	}

	// Do not remove from empty line
	line := window.buffer.Lines()[window.cursor.RunePosition().row]
	if line.start == line.end {
		log.Printf("Cannot remove from empty line\n")
		return
	}

	// Erase rune at cursor index
	text := window.buffer.Content()[window.cursor.index:line.end]
	_, length := utf8.DecodeRune(text)
	window.buffer.Erase(Region{window.cursor.index, window.cursor.index + length})

	// If cursor is at the end of the line after removal and the line is not empty move cursor back
	if window.cursor.index != line.start && window.cursor.index+length == line.end {
		window.cursor, _ = window.cursor.RunesBackward(1)
	}

	log.Println("Removed succesfully")
}

func (window *Window) deleteRange(r Region) {
	window.buffer.Erase(r)
}

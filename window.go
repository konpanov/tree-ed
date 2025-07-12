package main

import (
	"log"

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

type Window struct {
	filename     string
	mode         WindowMode
	buffer       IBuffer
	cursor       BufferCursor
	cursorAnchor int
	secondCursor BufferCursor
	anchorDepth  int
	node         *sitter.Node
	tree         *sitter.Tree
	parser       Scanner
	undotree     *ChangeTree
	shift_node   *sitter.Node
	shift_tree   *sitter.Tree
}

func windowFromBuffer(buffer IBuffer) *Window {
	window := &Window{
		mode:         NormalMode,
		buffer:       buffer,
		cursor:       NewBufferCursor(buffer),
		cursorAnchor: 0,
		secondCursor: NewBufferCursor(buffer),
		anchorDepth:  0,
		node:         buffer.Tree().RootNode(),
		tree:         buffer.Tree(),
		parser:       &NormalScanner{},
		undotree:     &ChangeTree{buffer, []Modification{}, 0},
	}

	return window
}

func (window *Window) Scan(ev tcell.Event) (Operation, error) {
	return window.parser.Scan(ev)
}

func (window *Window) switchToInsert() {
	window.mode = InsertMode
	window.parser = &InsertScanner{}
}
func (window *Window) switchToNormal() {
	window.mode = NormalMode
	window.parser = &NormalScanner{}
}

func (window *Window) switchToVisual() {
	window.mode = VisualMode
	window.secondCursor = window.cursor
	window.parser = &VisualScanner{}
}

func (window *Window) switchToTree() {
	log.Println("Switch to tree mode")
	window.mode = TreeMode
	window.parser = &TreeScanner{}
	window.setNode(NodeLeaf(window.buffer.Tree().RootNode(), window.cursor.Index()))
	window.anchorDepth = Depth(window.getNode())
}

func (window *Window) setNode(node *sitter.Node) {
	if node == nil {
		log.Panic("Cannot set node do nil value")
	}
	log.Println("Setting node")
	window.node = node
	window.shift_node = node
	window.cursor, _ = window.cursor.ToIndex(int(window.node.StartByte()))
	window.secondCursor, _ = window.cursor.ToIndex(max(int(window.node.EndByte())-1, 0))
	log.Println("Node set")
}

func (window *Window) getNode() *sitter.Node {
	return window.node
}

// Tree movements
func (window *Window) nodeUp() {
	parent := window.getNode().Parent()
	if parent == nil {
		return
	}
	window.setNode(parent)
	window.anchorDepth = Depth(window.getNode())

}

func (window *Window) nodeDown() {
	if window.getNode().ChildCount() == 0 {
		return
	}
	window.setNode(window.getNode().Child(0))
	window.anchorDepth = Depth(window.getNode())
}

func (window *Window) nodeNextSibling() {
	if sibling := window.getNode().NextSibling(); sibling != nil {
		window.setNode(sibling)
	}
}

func (window *Window) nodePrevSibling() {
	if sibling := window.getNode().PrevSibling(); sibling != nil {
		window.setNode(sibling)
	}
}

func (window *Window) nodeNextCousin() {
	if cousin := NextCousinDepth(window.getNode(), window.anchorDepth); cousin != nil {
		window.setNode(cousin)
	}
}

func (window *Window) nodePrevCousin() {
	if cousin := PrevCousinDepth(window.getNode(), window.anchorDepth); cousin != nil {
		window.setNode(cousin)
	}
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
	window.cursorAnchor = next.RunePosition().col
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
	window.cursorAnchor = prev.RunePosition().col
	log.Printf("Cursor moved to index: %d\n", window.cursor.index)
}

func (window *Window) cursorDown() {
	next, err := window.cursor.VerticalShift(1, window.cursorAnchor)
	panic_if_error(err)
	window.cursor = next
}

func (window *Window) cursorUp() {
	next, err := window.cursor.VerticalShift(-1, window.cursorAnchor)
	panic_if_error(err)
	window.cursor = next
}

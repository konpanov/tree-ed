package main

import (
	"log"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type WindowMode string

const (
	NormalMode WindowMode = "Normal"
	InsertMode WindowMode = "Insert"
	VisualMode WindowMode = "Visual"
	TreeMode   WindowMode = "Tree"
)

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
	scanner      Scanner
	undotree     *ChangeTree
}

func windowFromBuffer(buffer IBuffer) *Window {
	window := &Window{
		mode:         NormalMode,
		buffer:       buffer,
		cursor:       BufferCursor{buffer: buffer, index: 0, as_edge: false},
		cursorAnchor: 0,
		secondCursor: BufferCursor{buffer: buffer, index: 0, as_edge: false},
		anchorDepth:  0,
		tree:         buffer.Tree(),
		scanner:      &NormalScanner{},
		undotree:     &ChangeTree{buffer, []Change{}, 0},
	}

	return window
}

func (self *Window) switchToInsert() {
	self.mode = InsertMode
	self.setCursor(self.cursor.AsEdge(), true)
	self.secondCursor = self.secondCursor.AsEdge()
	self.scanner = &InsertScanner{}
}
func (self *Window) switchToNormal() {
	self.mode = NormalMode
	self.setCursor(self.cursor.AsChar(), true)
	self.secondCursor = self.secondCursor.AsChar()
	self.scanner = &NormalScanner{}
}

func (self *Window) switchToVisual() {
	self.mode = VisualMode
	self.setCursor(self.cursor.AsChar(), true)
	self.secondCursor = self.secondCursor.AsChar()
	self.scanner = &VisualScanner{}
}

func (self *Window) switchToTree() {
	if self.buffer.Tree() != nil {
		self.mode = TreeMode
		self.cursor.as_edge = false
		self.secondCursor.as_edge = false
		self.scanner = &TreeScanner{}
	}
}

func (self *Window) setCursor(cursor BufferCursor, updateAnchor bool) {
	self.cursor = cursor
	if updateAnchor {
		self.cursorAnchor = self.cursor.RunePosition().col
	}
	if self.mode == InsertMode || self.mode == NormalMode {
		self.secondCursor = self.cursor
	}
}

func (self *Window) setNode(node *sitter.Node, updateDepth bool) {
	if self.buffer.Tree() == nil {
		return
	}
	if node == nil {
		log.Panic("Cannot set node to nil value")
	}
	self.setCursor(self.cursor.ToIndex(int(node.StartByte())), true)
	self.secondCursor = self.cursor.ToIndex(int(node.EndByte()) - 1)
	if updateDepth {
		self.anchorDepth = Depth(node)
	}
}

func (self *Window) getNode() *sitter.Node {
	if self.buffer.Tree() == nil {
		return nil
	}
	var start, end uint
	if self.mode == VisualMode || self.mode == TreeMode {
		start, end = order(uint(self.cursor.Index()), uint(self.secondCursor.Index()))
	} else {
		start, end = order(uint(self.cursor.Index()), uint(self.cursor.Index()))
	}
	end++
	node := MinimalNodeDepth(self.buffer.Tree().RootNode(), start, end, self.anchorDepth)
	if node == nil {
		node = self.buffer.Tree().RootNode()
	}
	return node
}

func (self *Window) getSelection() (uint, uint) {
	start, end := order(uint(self.cursor.Index()), uint(self.secondCursor.Index()))
	return start, end + 1
}

// Tree movements
func (self *Window) nodeUp() {
	if self.buffer.Tree() == nil {
		return
	}
	node := self.getNode()
	parent := node.Parent()
	if parent == nil {
		return
	}
	self.setNode(parent, true)
}

func (self *Window) nodeDown() {
	if self.buffer.Tree() == nil {
		return
	}
	if self.getNode().ChildCount() == 0 {
		return
	}
	self.setNode(self.getNode().Child(0), true)
}

func (self *Window) nodeNextSibling() {
	if self.buffer.Tree() == nil {
		return
	}
	if sibling := self.getNode().NextSibling(); sibling != nil {
		self.setNode(sibling, false)
	}
}

func (self *Window) nodePrevSibling() {
	if self.buffer.Tree() == nil {
		return
	}
	if sibling := self.getNode().PrevSibling(); sibling != nil {
		self.setNode(sibling, false)
	}
}

func (self *Window) nodeNextCousin() {
	if self.buffer.Tree() == nil {
		return
	}
	if cousin := NextCousinDepth(self.getNode(), self.anchorDepth); cousin != nil {
		self.setNode(cousin, false)
	}
}

func (self *Window) nodePrevCousin() {
	if self.buffer.Tree() == nil {
		return
	}
	if cousin := PrevCousinDepth(self.getNode(), self.anchorDepth); cousin != nil {
		self.setNode(cousin, false)
	}
}

func (self *Window) nodeToFirstSibling() {
	if self.buffer.Tree() == nil {
		return
	}
	self.setNode(FirstSibling(self.getNode()), false)
}

func (self *Window) nodeToLastSibling() {
	if self.buffer.Tree() == nil {
		return
	}
	self.setNode(LastSibling(self.getNode()), false)
}

func (self *Window) cursorRight(count int) {
	col := self.cursor.RunePosition().col + count
	self.setCursor(self.cursor.MoveToCol(col), true)
}

func (self *Window) cursorLeft(count int) {
	col := self.cursor.RunePosition().col - count
	self.setCursor(self.cursor.MoveToCol(col), true)
}

func (self *Window) cursorUp(count int) {
	pos := Point{row: self.cursor.Row() - count, col: self.cursorAnchor}
	self.setCursor(self.cursor.MoveToRunePos(pos), false)
}

func (self *Window) cursorDown(count int) {
	pos := Point{row: self.cursor.Row() + count, col: self.cursorAnchor}
	self.setCursor(self.cursor.MoveToRunePos(pos), false)
}

func (self *Window) eraseLineAtCursor(count int) {
	composite := CompositeChange{}
	pos := self.cursor.RunePosition()
	for range count {
		mod := NewEraseLineModification(self, self.cursor.Row())
		mod.cursorBefore = self.cursor.Index()
		mod.Apply(self)
		self.setCursor(self.cursor.MoveToRunePos(Point{pos.row, self.cursorAnchor}), true)
		self.secondCursor = self.cursor
		mod.cursorAfter = self.cursor.Index()
		composite.changes = append(composite.changes, mod)
	}
	self.undotree.Push(composite)
}

func (self *Window) insertContent(continuous bool, content []byte) {
	if continuous {
		if mod := self.undotree.Back(); mod != nil {
			mod.Reverse().Apply(self)
		}
	}
	mod := NewReplacementModification(self.cursor.Index(), []byte{}, content)
	mod.cursorBefore = self.cursor.Index()
	mod.cursorAfter = self.cursor.Index() + len(mod.after)
	mod.Apply(self)
	self.undotree.Push(mod)
}

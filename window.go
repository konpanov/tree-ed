package main

import (
	"log"
	"slices"

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
	filename         string
	mode             WindowMode
	buffer           IBuffer
	cursor           BufferCursor
	secondCursor     BufferCursor
	cursorAnchor     int
	anchorDepth      int
	undotree         *UndoTree
	continuousInsert bool
}

func windowFromBuffer(buffer IBuffer) *Window {
	window := &Window{
		mode:         NormalMode,
		buffer:       buffer,
		cursor:       BufferCursor{buffer: buffer, index: 0, as_edge: false},
		cursorAnchor: 0,
		secondCursor: BufferCursor{buffer: buffer, index: 0, as_edge: false},
		anchorDepth:  0,
		undotree:     &UndoTree{buffer, []UndoState{}, 0},
	}
	window.buffer.RegisterCursor(&window.cursor)
	window.buffer.RegisterCursor(&window.secondCursor)
	return window
}

func (self *Window) switchToInsert() {
	self.mode = InsertMode
	self.setCursor(self.cursor.AsEdge(), true)
	self.secondCursor = self.secondCursor.AsEdge()
}
func (self *Window) switchToNormal() {
	self.mode = NormalMode
	self.setCursor(self.cursor.AsChar(), true)
	self.secondCursor = self.secondCursor.AsChar()
}

func (self *Window) switchToVisual() {
	self.mode = VisualMode
	self.setCursor(self.cursor.AsChar(), true)
	self.secondCursor = self.secondCursor.AsChar()
}

func (self *Window) switchToTree() {
	if self.buffer.Tree() != nil {
		self.mode = TreeMode
		self.cursor.as_edge = false
		self.secondCursor.as_edge = false
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

func (self *Window) nodeNextSiblingOrCousin() {
	if self.buffer.Tree() == nil {
		return
	}
	if node := NextSiblingOrCousinDepth(self.getNode(), self.anchorDepth); node != nil {
		self.setNode(node, false)
	}
}

func (self *Window) nodePrevSiblingOrCousin() {
	if self.buffer.Tree() == nil {
		return
	}
	if node := PrevSiblingOrCousinDepth(self.getNode(), self.anchorDepth); node != nil {
		self.setNode(node, false)
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
	for range count {
		change := NewEraseLineChange(self, self.cursor.Row())
		change.cursorBefore = self.cursor.Index()
		change.secondCursorBefore = self.cursor.Index()
		change.Apply(self)
		self.setCursor(self.cursor.MoveToCol(self.cursorAnchor), false)
		self.secondCursor = self.cursor
		change.cursorAfter = self.cursor.Index()
		change.secondCursorAfter = self.cursor.Index()
		composite.changes = append(composite.changes, change)
	}
	self.undotree.Push(UndoState{change: composite}, true)
}

// Add test if cursor after change is equal to current cursor
func (self *Window) insertContent(continuous bool, content []byte) {
	assert(len(content) != 0, "Inserted content should not be empty")
	var change ReplaceChange
	last_change := self.undotree.Curr()
	replace, is_replace := last_change.(ReplaceChange)
	cursor_pos := self.cursor.Index()
	if continuous && last_change != nil && is_replace {
		self.undotree.Back()
		last_change.Reverse().Apply(self)
		change = replace
		change.after = append(change.after, content...) // TODO: Adjust after adding insert left/right movements
	} else {
		change = NewReplacementChange(self.cursor.Index(), []byte{}, content)
	}
	change.cursorAfter = cursor_pos + len(content)
	change.secondCursorAfter = change.cursorAfter
	change.Apply(self)
	self.undotree.Push(UndoState{change: change}, false)
}

func (self *Window) eraseContent(continuous bool) {
	var change ReplaceChange
	last_change := self.undotree.Curr()
	replace, is_replace := last_change.(ReplaceChange)
	cursor_before := self.cursor
	cursor_after := cursor_before.RunePrev()
	if continuous && last_change != nil && is_replace {
		self.undotree.Back()
		last_change.Reverse().Apply(self)
		change = replace
		if len(change.after) != 0 {
			change.after = change.after[:len(change.after)-1] // TODO: Adjust after adding insert left/right movements
		} else {
			start := cursor_after.Index()
			change.before = slices.Clone(self.buffer.Content()[start : change.at+len(change.before)])
			change.at = start
		}
	} else {
		change = NewEraseChange(self, cursor_before.Index(), cursor_after.Index())
	}

	change.cursorAfter = cursor_after.Index()
	change.secondCursorAfter = change.cursorAfter
	change.Apply(self)
	self.undotree.Push(UndoState{change: change}, false)
}

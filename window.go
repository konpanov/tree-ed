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
	mode             WindowMode
	buffer           IBuffer
	cursor           BufferCursor
	anchor           BufferCursor
	originColumn     int
	originDepth      int
	history          *History
	continuousInsert bool
	frame            Rect
}

func windowFromBuffer(buffer IBuffer, width int, height int) *Window {
	window := &Window{
		mode:         NormalMode,
		buffer:       buffer,
		cursor:       BufferCursor{buffer: buffer, index: 0, as_edge: false},
		originColumn: 0,
		anchor:       BufferCursor{buffer: buffer, index: 0, as_edge: false},
		originDepth:  0,
		history:      &History{buffer, []HistoryState{}, 0},
		frame:        Rect{},
	}
	window.buffer.RegisterCursor(&window.cursor)
	window.buffer.RegisterCursor(&window.anchor)
	window.ResizeFrame(width, height)
	return window
}

func (self *Window) ResizeFrame(width int, height int) {
	self.frame.right = self.frame.left + width
	self.frame.bot = self.frame.top + height
	self.frame = self.frame.ShiftToInclude(self.anchor.Pos())
	self.frame = self.frame.ShiftToInclude(self.cursor.Pos())
}

func (self *Window) switchToInsert() {
	self.mode = InsertMode
	self.setCursor(self.cursor.AsEdge(), true)
	self.setAnchor(self.anchor.AsEdge())
}
func (self *Window) switchToNormal() {
	self.mode = NormalMode
	self.setCursor(self.cursor.AsChar(), true)
	self.setAnchor(self.anchor.AsChar())
}

func (self *Window) switchToVisual() {
	self.mode = VisualMode
	self.setCursor(self.cursor.AsChar(), true)
	self.setAnchor(self.anchor.AsChar())
}

func (self *Window) switchToTree() {
	if self.buffer.Tree() != nil {
		self.mode = TreeMode
		self.setCursor(self.cursor.AsChar(), false)
		self.setAnchor(self.anchor.AsChar())
	}
}

func (self *Window) setCursor(cursor BufferCursor, setOriginColumn bool) {
	self.cursor = cursor
	if setOriginColumn {
		self.originColumn = self.cursor.Pos().col
	}
	if self.mode == InsertMode || self.mode == NormalMode {
		self.setAnchor(self.cursor)
	}
	self.frame = self.frame.ShiftToInclude(self.cursor.Pos())
}

func (self *Window) setAnchor(anchor BufferCursor) {
	self.anchor = anchor
}

func (self *Window) setNode(node *sitter.Node, updateDepth bool) {
	if self.buffer.Tree() == nil {
		return
	}
	if node == nil {
		log.Panic("Cannot set node to nil value")
	}
	self.setCursor(self.cursor.ToIndex(int(node.StartByte())), true)
	self.setAnchor(self.cursor.ToIndex(int(node.EndByte()) - 1))
	if updateDepth {
		self.originDepth = Depth(node)
	}
}

func (self *Window) getNode() *sitter.Node {
	if self.buffer.Tree() == nil {
		return nil
	}
	var start, end uint
	if self.mode == VisualMode || self.mode == TreeMode {
		start, end = order(uint(self.cursor.Index()), uint(self.anchor.Index()))
	} else {
		start, end = order(uint(self.cursor.Index()), uint(self.cursor.Index()))
	}
	end++
	node := MinimalNodeDepth(self.buffer.Tree().RootNode(), start, end, self.originDepth)
	if node == nil {
		node = self.buffer.Tree().RootNode()
	}
	return node
}

func (self *Window) getSelection() (uint, uint) {
	start, end := order(uint(self.cursor.Index()), uint(self.anchor.Index()))
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
	if node := NextSiblingOrCousinDepth(self.getNode(), self.originDepth); node != nil {
		self.setNode(node, false)
	}
}

func (self *Window) nodePrevSiblingOrCousin() {
	if self.buffer.Tree() == nil {
		return
	}
	if node := PrevSiblingOrCousinDepth(self.getNode(), self.originDepth); node != nil {
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
	col := self.cursor.Pos().col + count
	self.setCursor(self.cursor.MoveToCol(col), true)
}

func (self *Window) cursorLeft(count int) {
	col := self.cursor.Pos().col - count
	self.setCursor(self.cursor.MoveToCol(col), true)
}

func (self *Window) cursorUp(count int) {
	pos := Pos{row: self.cursor.Row() - count, col: self.originColumn}
	self.setCursor(self.cursor.MoveToRunePos(pos), false)
}

func (self *Window) cursorDown(count int) {
	pos := Pos{row: self.cursor.Row() + count, col: self.originColumn}
	self.setCursor(self.cursor.MoveToRunePos(pos), false)
}

func (self *Window) eraseLineAtCursor(count int) {
	composite := CompositeChange{}
	for range count {
		change := NewEraseLineChange(self, self.cursor.Row())
		change.cursorBefore = self.cursor.Index()
		change.anchorBefore = self.cursor.Index()
		change.Apply(self)
		self.setCursor(self.cursor.MoveToCol(self.originColumn), false)
		self.setAnchor(self.cursor)
		change.cursorAfter = self.cursor.Index()
		change.anchorAfter = self.cursor.Index()
		composite.changes = append(composite.changes, change)
	}
	self.history.Push(HistoryState{change: composite})
}

// Add test if cursor after change is equal to current cursor
func (self *Window) insertContent(content []byte) {
	if len(content) == 0 {
		return
	}
	var change ReplaceChange
	last_change := self.history.Curr()
	replace, is_replace := last_change.(ReplaceChange)
	cursor_pos := self.cursor.Index()
	if self.continuousInsert && last_change != nil && is_replace {
		self.history.Back()
		last_change.Reverse().Apply(self)
		change = replace
		change.after = append(change.after, content...) // TODO: Adjust after adding insert left/right movements
	} else {
		change = NewReplacementChange(self.cursor.Index(), []byte{}, content)
	}
	change.cursorAfter = cursor_pos + len(content)
	change.anchorAfter = change.cursorAfter
	change.Apply(self)
	self.history.Push(HistoryState{change: change})
}

func (self *Window) eraseContent() {
	var change ReplaceChange
	last_change := self.history.Curr()
	replace, is_replace := last_change.(ReplaceChange)
	cursor_before := self.cursor
	cursor_after := cursor_before.RunePrev()
	if self.continuousInsert && last_change != nil && is_replace {
		self.history.Back()
		last_change.Reverse().Apply(self)
		change = replace
		if len(change.after) != 0 {
			size := cursor_before.Index() - cursor_after.Index()
			change.after = change.after[:len(change.after)-size] // TODO: Adjust after adding insert left/right movements
		} else {
			start := cursor_after.Index()
			change.before = slices.Clone(self.buffer.Content()[start : change.at+len(change.before)])
			change.at = start
		}
	} else {
		change = NewEraseChange(self, cursor_before.Index(), cursor_after.Index())
	}

	change.cursorAfter = cursor_after.Index()
	change.anchorAfter = change.cursorAfter
	change.Apply(self)
	self.history.Push(HistoryState{change: change})
}

package main

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type WindowMode int

const (
	NormalMode WindowMode = iota
	InsertMode WindowMode = iota
	VisualMode WindowMode = iota
	TreeMode   WindowMode = iota
)

func modeToString(mode WindowMode) string {
	switch mode {
	case NormalMode:
		return "Normal mode"
	case InsertMode:
		return "Insert mode"
	case VisualMode:
		return "Visual mode"
	case TreeMode:
		return "Tree mode"
	default:
		return "Unkown mode"
	}
}

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
	mode          WindowMode
	buffer        *Buffer
	cursor        *WindowCursor
	secondCursor  *WindowCursor
	topLine       int // TODO: Should it be in WindowCursor?
	leftColumn    int
	width, height int
	tree          *sitter.Tree
	node          *sitter.Node
}

func windowFromBuffer(buffer *Buffer, width int, height int) *Window {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), nil, buffer.content)
	rootNode := tree.RootNode()
	return &Window{
		mode:   NormalMode,
		buffer: buffer,
		cursor: &WindowCursor{
			index:               0,
			row:                 0,
			col:                 0,
			originColumn:        0,
			invalidOriginColumn: false,
		},
		secondCursor: &WindowCursor{
			index:               0,
			row:                 0,
			col:                 0,
			originColumn:        0,
			invalidOriginColumn: false,
		},
		topLine:    0,
		leftColumn: 0,
		width:      width,
		height:     height,
		tree:       tree,
		node:       rootNode,
	}
}

func (window *Window) draw(screen tcell.Screen) {
	normalStyle := tcell.StyleDefault
	selectStyle := tcell.StyleDefault.Reverse(true)
	selectStart, selectEnd := window.cursor.index, window.secondCursor.index
	if window.mode == TreeMode {
		selectStart = int(window.node.StartByte())
		selectEnd = int(window.node.EndByte())
	}
	selectStart, selectEnd = min(selectStart, selectEnd), max(selectStart, selectEnd)

	for y, line := range window.buffer.lines[window.topLine:] {
		if line.start == line.end {
			if line.start == window.cursor.index {
				switch window.mode {
				case NormalMode:
					screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
					screen.ShowCursor(0, y)
				case InsertMode:
					screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
					screen.ShowCursor(0, y)
				default:
					screen.HideCursor()
				}
			}
		} else {
			end := line.end
			start := min(line.start+window.leftColumn, end)
			if window.mode == InsertMode {
				end++
			}
			for x, value := range window.buffer.content[start:end] {
				style := normalStyle
				index := start + x
				switch window.mode {
				case NormalMode:
					if window.cursor.index == index {
						screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
						screen.ShowCursor(x, y)
					}
				case InsertMode:
					if window.cursor.index == index {
						screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
						screen.ShowCursor(x, y)
					}
				case TreeMode, VisualMode:
					if isInRange(index, selectStart, selectEnd) {
						style = selectStyle
					}
					screen.HideCursor()
				}
				if value == '\r' {
					value = 'R'
				} else if value == '\n' {
					value = 'N'
				}
				screen.SetContent(x, y, rune(value), nil, style)
			}
		}
	}
}

func (window *Window) switchToInsert() {
	window.mode = InsertMode
}
func (window *Window) switchToNormal() {
	window.mode = NormalMode
}

func (window *Window) switchToVisual() {
	window.mode = VisualMode
	*window.secondCursor = *window.cursor
}

func (window *Window) switchToTree() {
	window.mode = TreeMode
}

// Tree movements
func (window *Window) nodeUp() {
	if window.node.Equal(window.tree.RootNode()) {
		return
	}
	window.node = window.node.Parent()
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())

	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	window.shiftToCursor(window.secondCursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) nodeDown() {
	if window.node.ChildCount() == 0 {
		return
	}
	window.node = window.node.Child(0)
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())
	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	window.shiftToCursor(window.secondCursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) nodeRight() {
	sibling := window.node.NextSibling()
	if sibling == nil {
		return
	}
	window.node = sibling
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())
	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	window.shiftToCursor(window.secondCursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) nodeLeft() {
	sibling := window.node.PrevSibling()
	if sibling == nil {
		return
	}
	window.node = sibling
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())
	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	window.shiftToCursor(window.secondCursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) moveCursor(move Movement) {
	switch move {
	case Up:
		window.cursorUp()
	case Down:
		window.cursorDown()
	case Left:
		window.cursorLeft()
	case Right:
		window.cursorRight()
	}
	window.normalizeCursor(window.cursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) normalizeCursor(cursor *WindowCursor) {
	lines := window.buffer.lines

	for lines[cursor.row].start > cursor.index {
		cursor.row--
	}
	for lines[cursor.row].end < cursor.index {
		cursor.row++
	}
	cursor.col = cursor.index - window.buffer.lines[cursor.row].start
	if window.cursor.invalidOriginColumn {
		log.Println("Updating origin column")
		cursor.originColumn = cursor.col
		cursor.invalidOriginColumn = false
	}
}

func (window *Window) shiftToCursor(cursor *WindowCursor) {
	if window.leftColumn+window.width <= cursor.col {
		log.Println("Cursor is to the right of the window. Moving window right.")
		window.leftColumn = cursor.col - window.width + 1
	} else if window.leftColumn > cursor.col {
		log.Println("Cursor is to the left of the window. Moving window left.")
		window.leftColumn = cursor.col
	}

	if window.topLine+window.height <= cursor.row {
		log.Println("Cursor is below window. Moving window down.")
		window.topLine = cursor.row - window.height + 1
	} else if window.topLine > cursor.row {
		log.Println("Cursor is above window. Moving window up.")
		window.topLine = cursor.row
	}

}

// Cursor movements
func (window *Window) cursorRight() {
	log.Println("Moving cursor to the right")
	log.Printf("Cursor at index: %d\n", window.cursor.index)
	line := window.buffer.lines[window.cursor.row]
	if window.mode == InsertMode {
		line.end++
	}
	if line.start == line.end {
		log.Println("Line is empty. Cursor stays in place")
	} else if window.mode != InsertMode && window.cursor.index == line.end-1 {
		log.Println("Cursor at the end of the line. Cursor stays in place")
	} else {
		window.cursor.index = min(window.cursor.index+1, line.end-1)
		window.cursor.invalidOriginColumn = true
		log.Printf("Cursor moved to index: %d\n", window.cursor.index)
		log.Printf("Cursor row: %d\n", window.cursor.row)
	}
}

func (window *Window) cursorLeft() {
	log.Println("Moving cursor to the left")
	log.Printf("Cursor at index: %d\n", window.cursor.index)

	line := window.buffer.lines[window.cursor.row]
	if line.start == line.end {
		log.Println("Line is empty. Cursor stays in place")
	} else if window.cursor.index == line.start {
		log.Println("Cursor at the start of the line. Cursor stays in place")
	} else {
		window.cursor.index = max(window.cursor.index-1, line.start)
		window.cursor.invalidOriginColumn = true
		log.Printf("Cursor moved to index: %d\n", window.cursor.index)
		log.Printf("Cursor row: %d\n", window.cursor.row)
	}
}

func (window *Window) cursorDown() {
	log.Println("Moving cursor down")
	cursor := window.cursor
	buffer := window.buffer
	log.Printf("Cursor at index: %d\n", cursor.index)
	if cursor.row == len(buffer.lines)-1 {
		log.Println("Cursor is already at the last line")
		return
	}
	// cursor.row++
	line := buffer.lines[cursor.row+1]
	if line.start == line.end {
		log.Println("Moved onto an empty line")
		cursor.index = line.end
	} else {
		cursor.index = min(line.start+cursor.originColumn, line.end-1)
	}
	log.Printf("Cursor moved to index: %d\n", cursor.index)
}

func (window *Window) cursorUp() {
	log.Println("Moving cursor up")
	cursor := window.cursor
	buffer := window.buffer
	log.Printf("Cursor at index: %d\n", cursor.index)
	if cursor.row == 0 {
		log.Println("Cursor is already at the first line")
		return
	}
	// cursor.row--
	line := buffer.lines[cursor.row-1]
	if line.start == line.end {
		log.Println("Moved onto an empty line")
		cursor.index = line.end
	} else {
		cursor.index = min(line.start+cursor.originColumn, line.end-1)
	}
	log.Printf("Cursor moved to index: %d\n", cursor.index)
}

func (window *Window) insert(value []byte) {
	log.Printf("Inserting %c and index %d\n", value, window.cursor.index)
	window.buffer.insert(window.cursor.index, value)
	window.cursor.index += len(value)
	window.cursor.invalidOriginColumn = true
	window.normalizeCursor(window.cursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) remove() {
	log.Printf("Removing at index %d\n", window.cursor.index)
	length := 1
	if matchBytes(window.buffer.content[window.cursor.index:], window.buffer.newLineSeq) {
		length = len(window.buffer.newLineSeq)
	}
	window.buffer.erease(window.cursor.index, window.cursor.index+length-1)
	window.cursor.index -= length
	window.cursor.invalidOriginColumn = true
	window.normalizeCursor(window.cursor)
	window.shiftToCursor(window.cursor)
}

func (window *Window) deleteRange(from int, to int) {
	window.buffer.erease(from, to)
}

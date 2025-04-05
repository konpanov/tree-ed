package main

import (
	"log"
	"math"
	"strconv"

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

func (c WindowCursor) log() {
	log.Printf("Cursor: %+v\n", c)
}

type Window struct {
	mode              WindowMode
	buffer            *Buffer
	cursor            *WindowCursor
	secondCursor      *WindowCursor
	topLine           int
	leftColumn        int
	width, height     int
	numberColumnWidth int
	node              *sitter.Node
}

func windowFromBuffer(buffer *Buffer, width int, height int) *Window {
	return &Window{
		mode:              NormalMode,
		buffer:            buffer,
		cursor:            &WindowCursor{},
		secondCursor:      &WindowCursor{},
		topLine:           0,
		leftColumn:        0,
		numberColumnWidth: int(max(math.Log10(float64(len(buffer.Lines()))))) + 2,
		width:             width,
		height:            height,
		node:              buffer.tree.RootNode(),
	}
}

func colorNode(styles []tcell.Style, node *sitter.Node) {
	if node == nil {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		start := int(node.Child(i).StartByte())
		styles[start] = styles[start].Background(tcell.ColorDarkGray)
	}
}

func (window *Window) shift_to_fit_cursor(cursor WindowCursor) {
	log.Println("Fitting window to cursor")
	log.Printf("Window top line before: %d\n", window.topLine)
	cursor.log()
	window.topLine = max(min(window.topLine, cursor.row), cursor.row-window.height+1)
	window.leftColumn = max(min(window.leftColumn, cursor.col), cursor.col-(window.width-window.numberColumnWidth)+1)
	log.Printf("Window top line after: %d\n", window.topLine)
}

func (window *Window) draw_cursor(screen tcell.Screen, cursor WindowCursor) {
	screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	coord, err := window.buffer.Coord(cursor.index)
	if err != nil {
		log.Fatal("Could not find cursor index\n")
	}
	screen.ShowCursor(coord.col+window.numberColumnWidth-window.leftColumn, coord.row-window.topLine)
}

func (window *Window) draw_text(screen tcell.Screen) {
	lines := window.buffer.Lines()
	normalStyle := tcell.StyleDefault
	for y, line := range lines[window.topLine:] {
		end := line.end
		if window.mode == InsertMode || window.mode == TreeMode {
			end++
		}
		end = max(min(end, len(window.buffer.content)-1), 0)
		start := min(line.start+window.leftColumn, end)
		// end = max(min(end, start+endCol-startCol), 0)
		for x, value := range window.buffer.content[start:end] {
			// index := start + x
			if value == '\r' {
				value = ' '
			} else if value == '\n' {
				value = ' '
			}
			screen.SetContent(x+window.numberColumnWidth, y, rune(value), nil, normalStyle)
		}
	}
}

func (window *Window) draw_line_number_column(screen tcell.Screen) {
	lines := window.buffer.Lines()
	normalStyle := tcell.StyleDefault
	for y := range lines[window.topLine:] {
		line_num := strconv.Itoa(window.topLine + y + 1)
		for i, r := range line_num {
			screen.SetContent(window.numberColumnWidth-1-len(line_num)+i, y, r, nil, normalStyle)
		}
	}
}

func (window *Window) draw_normal(screen tcell.Screen) {
	log.Printf("Window: %+v\n", window)
	lines := window.buffer.Lines()

	window.numberColumnWidth = int(max(math.Log10(float64(len(lines))))) + 2
	window.shift_to_fit_cursor(*window.cursor)
	window.draw_cursor(screen, *window.cursor)
	window.draw_text(screen)
	window.draw_line_number_column(screen)
}

func (window *Window) draw(screen tcell.Screen) {
	window.draw_normal(screen)
	return
	// lines := window.buffer.Lines()
	// log.Println("Window height: ", window.height)
	// window.shift_to_fit_cursor(*window.cursor)
	//
	// // // if window.leftColumn+window.width <= cursor.col {
	// // // 	window.leftColumn = cursor.col - window.width + 1
	// // // } else if window.leftColumn > cursor.col {
	// // // 	window.leftColumn = cursor.col
	// // // }
	// //
	// // height := window.height - padding.top - padding.bottom
	// // if window.topLine > window.cursor.row {
	// // 	window.topLine = window.cursor.row
	// // } else if window.topLine+height <= window.cursor.row {
	// // 	window.topLine = window.cursor.row
	// // }
	//
	// log.Println("Starting to draw to screen")
	// log.Printf("Cursor index: %d", window.cursor.index)
	// normalStyle := tcell.StyleDefault
	// selectStyle := tcell.StyleDefault.Reverse(true)
	//
	// selectStart, selectEnd := window.cursor.index, window.secondCursor.index
	// if window.mode == TreeMode {
	// 	log.Println(window.node.String())
	// 	log.Println(window.node.Content(window.buffer.content))
	// 	selectStart = int(window.node.StartByte())
	// 	selectEnd = int(window.node.EndByte()) - 1
	// }
	// selectStart, selectEnd = order(selectStart, selectEnd)
	// log.Printf("%d %d", selectStart, selectEnd)
	// selectStart = clip(selectStart, 0, len(window.buffer.content))
	// selectEnd = clip(selectEnd, 0, len(window.buffer.content))
	// log.Printf("%d %d", selectStart, selectEnd)
	//
	// styles := []tcell.Style{}
	// for i := 0; i <= len(window.buffer.content); i++ {
	// 	styles = append(styles, normalStyle)
	// }
	// if window.mode == TreeMode {
	// 	if window.node != window.buffer.tree.RootNode() {
	// 		colorNode(styles, window.node.Parent())
	// 	}
	// }
	//
	// log.Println("Setting styles")
	// if window.mode == TreeMode || window.mode == VisualMode {
	// 	screen.HideCursor()
	// 	for i := selectStart; i <= selectEnd; i++ {
	// 		styles[i] = selectStyle
	// 	}
	// } else if window.mode == NormalMode {
	// 	screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	// 	coord, err := window.buffer.Coord(window.cursor.index)
	// 	if err != nil {
	// 		log.Fatal("Could not find cursor index\n")
	// 	}
	// 	screen.ShowCursor(coord.col+padding.left, coord.row+padding.top)
	// } else if window.mode == InsertMode {
	// 	screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
	// 	coord, err := window.buffer.Coord(window.cursor.index)
	// 	if err != nil {
	// 		log.Fatal("Could not find cursor index\n")
	// 	}
	// 	screen.ShowCursor(coord.col+padding.left, coord.row+padding.top)
	// }
	//
	// log.Println("Drawing content")
	// log.Println("Top line: ", window.topLine)
	// for y, line := range lines[window.topLine:] {
	// 	end := line.end
	// 	if window.mode == InsertMode || window.mode == TreeMode {
	// 		end++
	// 	}
	// 	end = max(min(end, len(window.buffer.content)-1), 0)
	// 	start := min(line.start+window.leftColumn, end)
	// 	// end = max(min(end, start+endCol-startCol), 0)
	// 	for x, value := range window.buffer.content[start:end] {
	// 		index := start + x
	// 		if value == '\r' {
	// 			value = ' '
	// 		} else if value == '\n' {
	// 			value = ' '
	// 		}
	// 		screen.SetContent(x+padding.left, y+padding.top, rune(value), nil, styles[min(index, len(styles)-1)])
	// 	}
	//
	// 	line_num := strconv.Itoa(window.topLine + y + 1)
	// 	for i, r := range line_num {
	// 		screen.SetContent(padding.left-1-len(line_num)+i, y+padding.top, r, nil, normalStyle)
	// 	}
	// }
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
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())
	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
}

// Tree movements
func (window *Window) nodeUp() {
	if window.node.Equal(window.buffer.tree.RootNode()) {
		return
	}
	window.node = window.node.Parent()
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())

	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
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
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
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
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
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
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
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
	// window.shiftToCursor(window.cursor)
}

func (window *Window) normalizeCursor(cursor *WindowCursor) {
	lines := window.buffer.Lines()
	log.Println("Normalizing cursor")
	log.Printf("%d\n", cursor.index)
	cursor.index = max(min(cursor.index, len(window.buffer.content)-1), 0)

	for lines[cursor.row].start > cursor.index {
		cursor.row--
	}
	for lines[cursor.row].end < cursor.index {
		cursor.row++
	}
	cursor.col = cursor.index - window.buffer.Lines()[cursor.row].start
	if window.cursor.invalidOriginColumn {
		log.Println("Updating origin column")
		cursor.originColumn = cursor.col
		cursor.invalidOriginColumn = false
	}
}

func (window *Window) shiftToCursor(cursor *WindowCursor) {
	log.Println("Shifting window to cursor")
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
	line := window.buffer.Lines()[window.cursor.row]
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

	line := window.buffer.Lines()[window.cursor.row]
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
	if cursor.row == len(buffer.Lines())-1 {
		log.Println("Cursor is already at the last line")
		return
	}
	line := buffer.Lines()[cursor.row+1]
	if line.start == line.end {
		log.Println("Moved onto an empty line")
		cursor.index = line.end
	} else {
		cursor.index = min(line.start+cursor.originColumn, line.end-1)
	}
	log.Printf("Cursor moved to index: %d. Buffer length %d\n", cursor.index, len(buffer.content))
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
	line := buffer.Lines()[cursor.row-1]
	if line.start == line.end {
		log.Println("Moved onto an empty line")
		cursor.index = line.end
	} else {
		cursor.index = min(line.start+cursor.originColumn, line.end-1)
	}
	log.Printf("Cursor moved to index: %d\n", cursor.index)
}

func (window *Window) insert(value []byte) {
	log.Printf("Inserting %c at index %d\n", value, window.cursor.index)
	window.buffer.Insert(window.cursor.index, value)
	window.cursor.index += len(value)
	window.cursor.invalidOriginColumn = true
	window.normalizeCursor(window.cursor)
	// window.shiftToCursor(window.cursor)
}

func (window *Window) remove() {
	log.Printf("Removing at index %d\n", window.cursor.index)
	length := 1
	if matchBytes(window.buffer.content[window.cursor.index:], window.buffer.nl_seq) {
		length = len(window.buffer.nl_seq)
	}
	toDeleteRange := Region{max(window.cursor.index-1, 0), window.cursor.index + length - 2}
	window.buffer.Erase(toDeleteRange)
	log.Println("Removed succesfully")
	window.cursor.index -= toDeleteRange.end - toDeleteRange.start
	window.cursor.index = max(min(window.cursor.index, len(window.buffer.content)-1), 0)
	window.cursor.invalidOriginColumn = true
	window.normalizeCursor(window.cursor)
	// window.shiftToCursor(window.cursor)
}

// TODO: update deleteRange call to not use returned range
func (window *Window) deleteRange(r Region) {
	window.buffer.Erase(r)
	window.cursor.index = max(min(r.start, len(window.buffer.content)-1), 0)
	window.cursor.invalidOriginColumn = true
	*window.secondCursor = *window.cursor
	window.normalizeCursor(window.cursor)
	// window.shiftToCursor(window.cursor)
}

func (window *Window) deleteNode() {
	start := window.cursor.index
	end := window.secondCursor.index
	start, end = order(start, end)
	end--
	r := Region{start, end}
	window.deleteRange(r)
	window.node = window.buffer.tree.RootNode()
	window.cursor.index = int(window.node.StartByte())
	window.secondCursor.index = int(window.node.EndByte())
	window.normalizeCursor(window.cursor)
	window.normalizeCursor(window.secondCursor)
	// window.shiftToCursor(window.secondCursor)
	// window.shiftToCursor(window.cursor)
	log.Printf("%d %d", window.cursor.index, window.secondCursor.index)
}

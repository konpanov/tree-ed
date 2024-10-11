package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

func main() {
	filename := "main.go"
	if len(os.Args) >= 2 {
		filename = os.Args[1]
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	buf := bufferFromFile(filename, []byte("\r\n"))

	defer quit(screen, buf)
	for !buf.quiting {
		drawWindow(screen, buf)
		handleEvents(screen.PollEvent(), buf)
	}
}

func drawWindow(screen tcell.Screen, buf *Buffer) {
	screen.Clear()
	drawCharachters(screen, buf)
	s.ShowCursor(buf.cursor.column, buf.cursor.row)
	// drawHighlighted(screen, buf)
	screen.Show()
}

func drawCharachters(s tcell.Screen, buf *Buffer) {
	lineStart := 0
	for y, line := range buf.lines {
		for x := 0; x < line.width; x++ {
			i := lineStart + x
			s.SetContent(x, y, rune(buf.content[i]), nil, tcell.StyleDefault)
		}
		lineStart += line.width + len(buf.newLineSeq)
	}
}

func handleEvents(ev tcell.Event, buf *Buffer) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		buf.quiting = ev.Key() == tcell.KeyCtrlC
		// handleInsertModeEvents(buf, ev)
		handleNormalModeEvents(buf, ev)
		// handleVisualModeEvents(buf, ev)
	}
}

func quit(screen tcell.Screen, buf *Buffer) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func handleNormalModeEvents(buf *Buffer, ev *tcell.EventKey) {
	// if buf.mode != NormalMode {
	// 	return
	// }
	switch ev.Key() {
	// case tcell.KeyCtrlS:
	// 	writeFile(buf)
	case tcell.KeyRune:
		switch ev.Rune() {
		// case 'i':
		// 	enterInsertMode(buf)
		// case 'a':
		// 	enterInsertMode(buf)
		// 	cursorRight(buf)
		// case 'v':
		// 	enterVisualMode(buf)
		case 'h':
			buf.cursorLeft()
		case 'j':
			buf.cursorDown()
		case 'k':
			buf.cursorUp()
		case 'l':
			buf.cursorRight()
		}
	}
}

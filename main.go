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

	buffer := bufferFromFile(filename, []byte("\r\n"))
	window := windowFromBuffer(buffer)

	defer quit(screen)
	for !buffer.quiting {
		drawWindow(screen, window)
		handleEvents(screen.PollEvent(), window)
	}
}

func drawWindow(screen tcell.Screen, window *Window) {
	screen.Clear()
	drawCharachters(screen, window)
	screen.ShowCursor(window.cursor.column, window.cursor.row)
	// drawHighlighted(screen, window)
	screen.Show()
}

func drawCharachters(s tcell.Screen, window *Window) {
	lineStart := 0
	for y, line := range window.buffer.lines {
		for x := 0; x < line.width; x++ {
			i := lineStart + x
			s.SetContent(x, y, rune(window.buffer.content[i]), nil, tcell.StyleDefault)
		}
		lineStart += line.width + len(window.buffer.newLineSeq)
	}
}

func handleEvents(ev tcell.Event, window *Window) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		window.buffer.quiting = ev.Key() == tcell.KeyCtrlC
		// handleInsertModeEvents(window, ev)
		handleNormalModeEvents(window, ev)
		// handleVisualModeEvents(window, ev)
	}
}

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func handleNormalModeEvents(window *Window, ev *tcell.EventKey) {
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
			window.cursorLeft()
		case 'j':
			window.cursorDown()
		case 'k':
			window.cursorUp()
		case 'l':
			window.cursorRight()
		}
	}
}

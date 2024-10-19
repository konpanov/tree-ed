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

	width, height := screen.Size()
	buffer := bufferFromFile(filename, []byte("\r\n"))
	window := windowFromBuffer(buffer, width, height)

	defer quit(screen)
	for !buffer.quiting {
		screen.Clear()
		window.draw(screen)
		screen.Show()

		handleEvents(screen.PollEvent(), window)
	}
}

func handleEvents(ev tcell.Event, window *Window) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		window.buffer.quiting = ev.Key() == tcell.KeyCtrlC
		handleInsertModeEvents(window, ev)
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
	if window.mode != NormalMode {
		return
	}
	switch ev.Key() {
	// case tcell.KeyCtrlS:
	// 	writeFile(window)
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'i':
			window.switchToInsert()
		case 'a':
			window.switchToInsert()
			window.cursorRight()
		// case 'v':
		// 	enterVisualMode(window)

		//TODO: add some timeout?
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

func handleInsertModeEvents(window *Window, ev *tcell.EventKey) {
	if window.mode != InsertMode {
		return
	}
	switch ev.Key() {
	case tcell.KeyEsc:
		window.switchToNormal()
	// case tcell.KeyBS:
	// 	removeUnderCursor(window)
	// case tcell.KeyEnter:
	// 	splitLineUnderCursor(window)
	case tcell.KeyRune:
		window.insert(byte(ev.Rune()))
		window.cursorRight()
	}
}

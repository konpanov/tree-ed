package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

// aaasdsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaasdsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaasdsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaasdsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
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
	buffer := bufferFromFile(filename, []byte("\n"))
	window := windowFromBuffer(buffer, width, height)

	defer quit(screen)
	for !buffer.quiting {
		screen.Clear()
		window.draw(screen)
		screen.Show()

		handleEvents(screen.PollEvent(), window)
	}
}

func handleEvents(ev tcell.Event, window *Window) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		window.buffer.quiting = ev.Key() == tcell.KeyCtrlC
		return (handleInsertModeEvents(window, ev) ||
			handleNormalModeEvents(window, ev) ||
			handleVisualModeEvents(window, ev) ||
			handleTreeModeEvents(window, ev) ||
			false)
	}
	return false
}

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func handleNormalModeEvents(window *Window, ev *tcell.EventKey) bool {
	if window.mode != NormalMode {
		return false
	}
	if handleNormalMovements(window, ev) {
		return true
	}
	switch ev.Key() {
	// case tcell.KeyCtrlS:
	// 	writeFile(window)
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'i':
			window.switchToInsert()
			return true
		case 'a':
			window.switchToInsert()
			window.cursorRight()
			return true
		case 'v':
			window.switchToVisual()
			return true
		case 't':
			window.switchToTree()
			return true
		}
	}
	return false
}

func handleInsertModeEvents(window *Window, ev *tcell.EventKey) bool {
	if window.mode != InsertMode {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEsc:
		window.switchToNormal()
		return true
	case tcell.KeyBS:
		window.remove()
		return true
	// case tcell.KeyEnter:
	// 	splitLineUnderCursor(window)
	case tcell.KeyRune:
		window.insert(byte(ev.Rune()))
		window.cursorRight()
		return true
	}
	return false
}

func handleVisualModeEvents(window *Window, ev *tcell.EventKey) bool {
	if window.mode != VisualMode {
		return false
	}
	if handleNormalMovements(window, ev) {
		return true
	}
	switch ev.Key() {
	case tcell.KeyEsc:
		window.switchToNormal()
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'i':
			window.switchToInsert()
			return true
		case 'a':
			window.switchToInsert()
			window.cursorRight()
			return true
		case 'v':
			window.switchToNormal()
			return true
		}
	}
	return false
}

func handleTreeModeEvents(window *Window, ev *tcell.EventKey) bool {
	if window.mode != TreeMode {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEsc:
		window.switchToNormal()
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 't':
			window.switchToNormal()
			return true
		case 'k':
			window.nodeUp()
			return true
		case 'j':
			window.nodeDown()
			return true
		case 'l':
			window.nodeRight()
			return true
		case 'h':
			window.nodeLeft()
			return true
		}
	}
	return false
}

func handleNormalMovements(window *Window, ev *tcell.EventKey) bool {
	//TODO: add some timeout?
	if ev.Key() == tcell.KeyRune {
		switch ev.Rune() {
		case 'h':
			window.cursorLeft()
			return true
		case 'j':
			window.cursorDown()
			return true
		case 'k':
			window.cursorUp()
			return true
		case 'l':
			window.cursorRight()
			return true
		}
	}
	return false
}

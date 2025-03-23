package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/gdamore/tcell/v2"
)

func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Log file initiated.")

	f, err = os.Create("cpuprofile")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	log.Println("Cpu profile initiated")

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

	buffer := bufferFromFile(filename, getSystemNewLine())
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
	log.Println("[MODE: " + modeToString(window.mode) + "]")
	switch ev := ev.(type) {
	case *tcell.EventKey:
		log.Println("Registered key: ", tcell.KeyNames[ev.Key()])
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
		case 'd':
			line := window.buffer.lines[window.cursor.row]
			window.deleteRange(line)
		}
	}
	return false
}

func handleInsertModeEvents(window *Window, ev *tcell.EventKey) bool {
	if window.mode != InsertMode {
		return false
	}
	log.Println("Handling insert mode events")
	switch ev.Key() {
	case tcell.KeyEsc:
		window.switchToNormal()
		return true
	case tcell.KeyBackspace:
	case tcell.KeyBackspace2:
		window.remove()
		return true
	case tcell.KeyEnter:
		window.insert(getSystemNewLine())
	case tcell.KeyRune:
		window.insert([]byte{byte(ev.Rune())})
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
		case 'd':
			start := window.cursor.index
			end := window.secondCursor.index
			start, end = order(start, end)
			window.deleteRange(Range{start, end})
			window.cursor.index = start
			window.normalizeCursor(window.cursor)
			window.shiftToCursor(window.cursor)
			window.switchToNormal()
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
		case 'd':
			window.deleteNode()
		}
	}
	return false
}

func handleNormalMovements(window *Window, ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune {
		switch ev.Rune() {
		case 'h':
			window.moveCursor(Left)
			return true
		case 'j':
			window.moveCursor(Down)
			return true
		case 'k':
			window.moveCursor(Up)
			return true
		case 'l':
			window.moveCursor(Right)
			return true
		}
	}
	return false
}

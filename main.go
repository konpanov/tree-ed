package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/gdamore/tcell/v2"
)

var debug = true

func main() {
	// Setup logging to file
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	panic_if_error(err)
	defer f.Close()
	log.SetOutput(f)
	debug_logln("Log file initiated.")

	if debug {
		// Setup cpuprofile
		f, err = os.Create("cpuprofile")
		panic_if_error(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		debug_logln("Cpu profile initiated")
	}

	// Setup screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	defer quit(screen)

	editor := NewEditor(screen)

	if len(os.Args) >= 2 {
		filename := os.Args[1]
		editor.OpenFileInWindow(filename)
	}
	editor.Start()
}

func quit(screen tcell.Screen) {
	maybePanic := recover()
	screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

package main

import (
	// "log"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Editor struct {
	screen  tcell.Screen
	scanner *Scanner
	buffers []IBuffer
	windows []*Window
	curwin  *Window
	view    View
	theme   Theme

	is_quiting bool
}

func NewEditor(screen tcell.Screen) *Editor {
	editor := &Editor{
		screen:  screen,
		scanner: &Scanner{},
		buffers: []IBuffer{},
		windows: []*Window{},
		theme:   default_theme,
	}
	editor.view = &ViewEditor{editor: editor}
	return editor
}

func (self *Editor) OpenFileInWindow(filename string) {
	filename = filepath.Clean(filename)
	content, err := os.ReadFile(filename)
	panic_if_error(err)

	language := ParserLanguageByFileType(GetFiletype(filename))
	var parser *sitter.Parser
	if language != nil {
		parser = sitter.NewParser()
		parser.SetLanguage(language)
	}
	buffer, err := bufferFromContent(content, getContentLineBreak(content), parser)
	buffer.filename = filename
	panic_if_error(err)
	self.OpenBuffer(buffer)
}

func (self *Editor) OpenBuffer(buffer IBuffer) {
	self.buffers = append(self.buffers, buffer)
	w, h := self.screen.Size()
	window := windowFromBuffer(buffer, w, h)
	self.curwin = window
}

func (self *Editor) Close() {
	for _, buf := range self.buffers {
		buf.Close()
	}
}

func (self *Editor) Redraw() {
	width, height := self.screen.Size()
	roi := Rect{left: 0, right: width, top: 0, bot: height}
	self.view.Draw(DrawContext{screen: self.screen, roi: roi, theme: self.theme})
	self.screen.Show()

}
func (self *Editor) Start() {
	defer self.Close()

	events := make(chan tcell.Event, 100)

	go func() {
		var ev tcell.Event
		for do := true; do; do = ev != nil {
			log.Println("Polling event")
			ev = self.screen.PollEvent()
			events <- ev
		}
	}()

	got_new_event := true
	for !self.is_quiting {
		if got_new_event {
			self.Redraw()
			got_new_event = false
		}

		waiting_for_event := true
		for waiting_for_event {
			select {
			case e := <-events:
				log.Printf("Polled event: %+v\n", e)
				switch v := e.(type) {
				case *tcell.EventKey:
					log.Printf(eventKeyToString(v))
				}
				self.scanner.Push(e)
				got_new_event = true
			case <-time.Tick(2 * time.Millisecond):
				waiting_for_event = false
			}
		}

		for got_new_event && !self.is_quiting {
			if self.curwin != nil {
				self.scanner.mode = self.curwin.mode
			}
			op, res := self.scanner.Scan()
			self.scanner.Update(res)
			if res == ScanStop {
				break
			}
			if res == ScanFull && op != nil {
				op.Execute(self, 1)
			}
		}
	}
}

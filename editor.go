package main

import (
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Editor struct {
	screen  tcell.Screen
	buffers []IBuffer
	windows []*Window
	curwin  *Window
	view    *WindowView

	is_quiting bool
}

func NewEditor(screen tcell.Screen) *Editor {
	return &Editor{
		screen:  screen,
		buffers: []IBuffer{},
		windows: []*Window{},
	}
}

func (self *Editor) OpenFileInWindow(filename string) {
	content, err := os.ReadFile(filename)
	panic_if_error(err)

	language := ParserLanguageByFileType(GetFiletype(filename))
	parser := sitter.NewParser()
	parser.SetLanguage(language)
	buffer, err := bufferFromContent(content, getContentNewLine(content), parser)
	panic_if_error(err)
	self.OpenBuffer(buffer)
}

func (self *Editor) OpenBuffer(buffer IBuffer) {
	self.buffers = append(self.buffers, buffer)
	window := windowFromBuffer(buffer)
	self.curwin = window
	self.view = NewWindowView(self.screen, self.GetRoi(), self.curwin)
	self.view.status_line = NewStatusLine(self.screen, self.curwin, self.view)

}

func (self *Editor) Close() {
	for _, buf := range self.buffers {
		buf.Close()
	}
}

func (self *Editor) GetRoi() Rect {
	width, height := self.screen.Size()
	return Rect{left: 0, right: width, top: 0, bot: height}
}

func (self *Editor) Prepare() {
}
func (self *Editor) Redraw() {
	self.view.Update(self.GetRoi())
	self.screen.Fill(' ', self.view.base_style)
	self.view.Draw()
	self.screen.Show()

}
func (self *Editor) Start() {
	defer self.Close()

	events := make(chan tcell.Event, 10000)

	go func() {
		for {
			ev := self.screen.PollEvent()
			log.Printf("Polled event: %+v\n", ev)

			switch v := ev.(type) {
			case *tcell.EventKey:
				log.Printf(eventKeyToString(v))
			}
			events <- ev
		}
	}()

	scanner := NewOmniScanner()
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
				scanner.Push(e)
				got_new_event = true
			case <-time.Tick(10 * time.Millisecond):
				waiting_for_event = false
			}
		}

		for got_new_event && !self.is_quiting {
			scanner.mode = self.curwin.mode
			op, err := scanner.Scan()
			if op == nil || err != nil {
				break
			}
			log.Printf("Executing %T: %+v\n", op, op)
			op.Execute(self, 1)
		}

	}
}

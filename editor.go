package main

import (
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
		scanner: NewScanner(),
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

	events := make(chan tcell.Event, 10000)

	go func() {
		for {
			ev := self.screen.PollEvent()
			if ev == nil {
				break
			}
			log.Printf("Polled event: %+v\n", ev)

			switch v := ev.(type) {
			case *tcell.EventKey:
				log.Printf(eventKeyToString(v))
			}
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
				self.scanner.Push(e)
				got_new_event = true
			case <-time.Tick(10 * time.Millisecond):
				waiting_for_event = false
			}
		}

		combo := &OperationCombiner{}
		for got_new_event && !self.is_quiting {
			if self.curwin != nil {
				self.scanner.mode = self.curwin.mode
			}
			op, res := self.scanner.Scan()
			if res == ScanStop {
				break
			}
			if res == ScanFull {
				combo.Push(op)
				op = combo.Get()
				if op != nil {
					op.Execute(self, 1)
				}
			}
		}
	}
}

type OperationCombiner struct {
	input []Operation
}

func (self *OperationCombiner) Push(op Operation) {
	self.input = append(self.input, op)
}

func (self *OperationCombiner) Get() Operation {
	if len(self.input) == 0 {
		return nil
	}
	op := self.input[0]
	count_op, is_count := op.(OpCount)
	if !is_count {
		self.input = self.input[1:]
		return op
	}
	if len(self.input) < 2 {
		return nil
	}
	count_op.op = self.input[1]
	self.input = self.input[2:]
	return count_op
}

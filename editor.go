package main

import (
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type IEditor interface{}

type Editor struct {
	screen  tcell.Screen
	buffers []IBuffer
	windows []*Window
	curwin  *Window

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
	self.buffers = append(self.buffers, buffer)

	window := windowFromBuffer(buffer)
	window.filename = filename
	self.windows = append(self.windows, window)
	self.curwin = window
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

func (self *Editor) Start() {
	window_view := NewWindowView(self.screen, self.GetRoi(), self.curwin)
	window_view.status_line = NewStatusLine(self.screen, self.curwin, window_view)

	defer self.Close()
	timestamp := time.Now()
	for !self.is_quiting {
		window_view.Update(self.GetRoi())
		self.screen.Fill(' ', window_view.base_style)
		window_view.Draw()
		self.screen.Show()

		ev := self.screen.PollEvent()
		log.Printf("%+v\n", ev)
		op, _ := GlobalScanner{}.Scan(ev)
		if op == nil {
			op, _ = self.curwin.Scan(ev)
		}
		if op != nil {
			op.Execute(self, 1)
		}
	}
}

package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
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

	buffer, err := bufferFromContent(content, getContentNewLine(content))
	panic_if_error(err)
	self.buffers = append(self.buffers, buffer)

	width, height := self.screen.Size()
	window := windowFromBuffer(buffer, width, height)
	window.filename = filename
	self.windows = append(self.windows, window)
	self.curwin = window
}

func (self *Editor) GetRoi() Rect {
	width, height := self.screen.Size()
	return Rect{Point{col: 0, row: 0}, Point{col: width, row: height}}
}

func (self *Editor) Start() {
	window_view := NewWindowView(self.screen, self.GetRoi(), self.curwin)
	window_view.status_line = NewStatusLine(self.screen, self.curwin, window_view)

	for !self.is_quiting {
		window_view.Update(self.GetRoi())
		self.screen.Clear()
		window_view.Draw()
		self.screen.Show()

		ev := self.screen.PollEvent()
		op, _ := GlobalParser{}.Parse(ev)
		if op == nil {
			op, _ = self.curwin.Parse(ev)
		}
		if op != nil {
			op.Execute(self)
		}
	}
}

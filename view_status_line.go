package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type IStatusLine interface {
	GetHeight() int
	SetRoi(roi Rect)
	GetRoi() Rect
	Draw()
}

type StatusLine struct {
	screen      tcell.Screen
	roi         Rect
	window      *Window
	window_view *WindowView
}

func NewStatusLine(screen tcell.Screen, window *Window, window_view *WindowView) *StatusLine {
	return &StatusLine{
		screen:      screen,
		window:      window,
		window_view: window_view,
	}
}

func (self *StatusLine) GetHeight() int {
	return 2
}

func (self *StatusLine) GetRoi() Rect {
	return self.roi
}

func (self *StatusLine) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *StatusLine) Draw() {
	pos := self.window.cursor.RunePosition()
	offset := self.window_view.text_offset

	input := ""
	if self.window.mode == NormalMode {
		history_parser, ok := self.window.parser.(*NormalScanner)
		if ok {
			for i := 0; i < len(history_parser.history); i++ {
				input += string(history_parser.history[i].Rune())
			}
		}
		
	}

	
	log.Println("Drawing status line")
	parseError := "Correct"
	if self.window.buffer.Tree().RootNode().HasError() {
		parseError = "Error"
	}

	text := [][]rune{
		[]rune(strings.Repeat("-", self.roi.Width())),
		[]rune(strings.Join(
			[]string{
				"file: " + self.window.filename,
				"line: " + strconv.Itoa(pos.row),
				"col: " + strconv.Itoa(pos.col),
				"mode: " + string(self.window.mode),
				"offset: " + strconv.Itoa(offset.row) + ":" + strconv.Itoa(offset.col),
				"input: " + input,
				"parse state: " + parseError,
			},
			", ",
		)),
	}
	text = text[:min(self.roi.Height(), len(text))]
	for i := 0; i < len(text); i++ {
		text[i] = text[i][:min(self.roi.Width(), len(text[i]))]
	}

	text_view := NewTextView(self.screen, self.roi, text)
	text_view.Draw()
}

type NoStatusLine struct {
}

func (self NoStatusLine) GetHeight() int {
	return 0
}

func (self NoStatusLine) GetRoi() Rect {
	return Rect{}
}

func (self NoStatusLine) SetRoi(roi Rect) {
}

func (self NoStatusLine) Draw() {
}

package main

import (
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
	screen tcell.Screen
	roi    Rect

	editor *Editor
}

func NewStatusLine(screen tcell.Screen, editor *Editor) *StatusLine {
	return &StatusLine{
		screen: screen,
		editor: editor,
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
	left_parts := []string{}

	if self.editor.curwin != nil {
		curwin := self.editor.curwin
		pos := curwin.cursor.RunePosition()
		parseError := "Correct"
		if curwin.buffer.Tree() != nil && curwin.buffer.Tree().RootNode().HasError() {
			parseError = "Error"
		}
		newline := newlinesToSymbols([]rune(string(curwin.buffer.Nl_seq())))
		left_parts = append(left_parts, "file: "+curwin.filename)
		left_parts = append(left_parts, "line: "+strconv.Itoa(pos.row+1))
		left_parts = append(left_parts, "col: "+strconv.Itoa(pos.col+1))
		left_parts = append(left_parts, "mode: "+string(curwin.mode))
		left_parts = append(left_parts, "parse state: "+parseError)
		left_parts = append(left_parts, "newline: "+string(newline))
	}
	// TODO: Move text offset state to window
	// offset := self.window_view.text_offset
	left_parts = append(left_parts, "input: "+KeyEventsToString(self.editor.scanner.state.Input()))

	text := [][]rune{
		[]rune(strings.Repeat("-", self.roi.Width())),
		[]rune(strings.Join(
			left_parts,
			// []string{
			// 	"file: " + self.window.filename,
			// 	"line: " + strconv.Itoa(pos.row+1),
			// 	"col: " + strconv.Itoa(pos.col+1),
			// 	"mode: " + string(self.window.mode),
			// 	// "offset: " + strconv.Itoa(offset.row) + ":" + strconv.Itoa(offset.col),
			// 	"input: " + input,
			// 	"parse state: " + parseError,
			// 	"newline: " + string(newline),
			// },
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

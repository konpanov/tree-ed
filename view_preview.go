package main

import (
	"github.com/gdamore/tcell/v2"
)

type PreviewView struct {
	screen tcell.Screen
	roi    Rect
}

func (self *PreviewView) GetRoi() Rect {
	return self.roi
}

func (self *PreviewView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *PreviewView) Draw() {
	width := self.roi.Width()
	height := self.roi.Height()

	self.screen.SetContent(width/2, height/2, 'X', nil, tcell.StyleDefault)
}

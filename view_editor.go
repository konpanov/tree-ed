package main

import (
	"github.com/gdamore/tcell/v2"
)

type ViewEditor struct {
	screen      tcell.Screen
	roi         Rect
	editor      *Editor
	main        View
	status_line IStatusLine
}

func (self *ViewEditor) GetRoi() Rect {
	return self.roi
}

func (self *ViewEditor) SetRoi(roi Rect) {
	self.roi = roi
}

func (self *ViewEditor) Draw() {
	status_line_height := self.status_line.GetHeight()
	main_roi, status_line_roi := self.roi.SplitH(self.roi.Height() - status_line_height)

	self.main.SetRoi(main_roi)
	self.status_line.SetRoi(status_line_roi)
	self.main.Draw()
	self.status_line.Draw()
}

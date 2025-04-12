package main

import (
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type StatusLine2 struct {
	screen   tcell.Screen
	roi      Rect
	filename string
	cursor   BufferCursor
	buffer   IBuffer
	mode     string
}

func (sl *StatusLine2) GetHeight() int {
	return 2
}

func (self *StatusLine2) Draw() {
	pos := self.cursor.RunePosition()
	log.Println("Drawing status line")
	info := ""
	info += "file: " + self.filename
	info += ", "
	info += "line: " + strconv.Itoa(pos.row)
	info += ", "
	info += "col: " + strconv.Itoa(pos.col)
	info += ", "
	info += "mode: " + self.mode

	for col := self.roi.Left(); col < self.roi.Right(); col++ {
		set_rune(self.screen, Point{col: col, row: self.roi.Top()}, '-')
	}

	for col, value := range info {
		pos := view_pos_to_screen_pos(Point{row: 1, col: col}, self.roi)
		set_rune(self.screen, pos, value)
	}
}

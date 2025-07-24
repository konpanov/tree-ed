package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

// Text view draws given text (splitted by lines) on the screen in a given ROI
// Implements View interface
type TextView struct {
	screen tcell.Screen
	roi    Rect
	text   [][]rune
	style  tcell.Style
}

func NewTextView(screen tcell.Screen, roi Rect, text [][]rune) *TextView {
	return &TextView{
		screen: screen,
		roi:    roi,
		text:   text,
	}
}

func (self *TextView) GetRoi() Rect {
	return self.roi
}

func (self *TextView) SetRoi(roi Rect) {
	self.roi = roi
}

func (self TextView) Draw() {
	number_of_lines := len(self.text)
	height := self.roi.Height()
	width := self.roi.Width()
	if number_of_lines > height {
		text := "Text does not fit roi height. Height: %d, Number of lines: %d\n"
		log.Panicf(text, height, number_of_lines)
	}
	for row, line := range self.text {
		length := len(line)
		if length > width {
			text := "Text does not fit roi width. Width: %d, Line id: %d, Length: %d\n"
			log.Panicf(text, width, row, length)
		}

		for col, value := range line {
			screen_pos := view_pos_to_screen_pos(Point{col: col, row: row}, self.roi)
			set_rune(self.screen, screen_pos, value)
			set_style(self.screen, screen_pos, self.style)
		}
	}
}

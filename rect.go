package main

import "fmt"

// TODO change to have 4 int members and add method to get points
type Rect struct {
	top_left, bot_right Point
}

func (r Rect) StringInfo() string {
	return fmt.Sprintf(
		"{top: %d, left: %d, right: %d, bot: %d, width: %d, height: %d}",
		r.Top(), r.Left(), r.Right(), r.Bot(), r.Width(), r.Height(),
	)

}

func (r Rect) Width() int {
	return r.bot_right.col - r.top_left.col
}

func (r Rect) Height() int {
	return r.bot_right.row - r.top_left.row
}

func (r Rect) Size() Point {
	return Point{col: r.Width(), row: r.Height()}
}

func (r Rect) Top() int {
	return r.top_left.row
}

func (r Rect) Left() int {
	return r.top_left.col
}

func (r Rect) Right() int {
	return r.bot_right.col
}

func (r Rect) Bot() int {
	return r.bot_right.row
}

func (r Rect) AdjustTop(by int) Rect {
	r.top_left.row += by
	return r
}

func (r Rect) AdjustLeft(by int) Rect {
	r.top_left.col += by
	return r
}

func (r Rect) AdjustRight(by int) Rect {
	r.bot_right.col += by
	return r
}

func (r Rect) AdjustBot(to int) Rect {
	r.bot_right.row += to
	return r
}

func (r Rect) SetTop(to int) Rect {
	r.top_left.row = to
	return r
}

func (r Rect) SetLeft(to int) Rect {
	r.top_left.col = to
	return r
}

func (r Rect) SetRight(to int) Rect {
	r.bot_right.col = to
	return r
}

func (r Rect) SetBot(to int) Rect {
	r.bot_right.row = to
	return r
}

func (r Rect) IsPointAbove(p Point) bool {
	return r.Top() > p.row
}

// Position in buffer: i
//
//	0,..., i-2, i-1, i, i+1, i+2, ..., n
//
//
// Position in bytes splitted by lines: y, x
//	y = line = number of new line before i
//	x = char = number of bytes after last new line or
//	N   = Number of lines
//	N_y = Number of chars in line y
//
//	(0,0), ..., (0,x), ..., ..., (0,N_0)
//	...  , ..., ...  , ..., ...
//	(y,0), ..., (y,x), ..., ..., ..., (y,N_y)
//	...  , ..., ...  , ..., ...
//	(N,0), ..., (N,x), ..., (0,N_N)

package main

type Rect struct {
	left, right, top, bot int
}

func (r Rect) Width() int {
	return r.right - r.left
}

func (r Rect) Height() int {
	return r.bot - r.top
}

// Split by a horizontal line
func (r Rect) SplitH(at int) (Rect, Rect) {
	upper, lower := r, r
	upper.bot = r.top + at
	lower.top = r.top + at
	return upper, lower
}

// Split by a vertical line
func (r Rect) SplitV(at int) (Rect, Rect) {
	to_the_left, to_the_right := r, r
	to_the_left.right = r.left + at
	to_the_right.left = r.left + at
	return to_the_left, to_the_right
}

func (r Rect) Size() Point {
	return Point{col: r.Width(), row: r.Height()}
}

func (r Rect) IsPointAbove(p Point) bool {
	return r.top > p.row
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

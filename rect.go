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

func (r Rect) AxisH() int {
	return (r.left + r.right) / 2
}

func (r Rect) AxisV() int {
	return (r.top + r.bot) / 2
}

func (r Rect) TopLeft() Point {
	return Point{row: r.top, col: r.left}
}

func (r Rect) BotRight() Point {
	return Point{row: r.bot, col: r.right}
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

func (r Rect) ShiftToInclude(pos Point) Rect {
	w := r.Width()
	r.left = max(min(r.left, pos.col), pos.col-w+1)
	r.right = r.left + w

	h := r.Height()
	r.top = max(min(r.top, pos.row), pos.row-h+1)
	r.bot = r.top + h

	return r
}

func (r Rect) Shift(pos Point) Rect {
	r.bot = pos.row + r.Height()
	r.right = pos.col + r.Width()
	r.top = pos.row
	r.left = pos.col
	return r
}

func (r Rect) Size() Point {
	return Point{col: r.Width(), row: r.Height()}
}

type RelativePosition int

const (
	Above RelativePosition = iota
	Below
	LeftOf
	RightOf
	Inside
)

// Above  \ Above  \ Above
// -------+--------+--------
// LeftOf \ Inside \ RightOf
// -------+--------+--------
// Below  \ Below  \ Below
func (r Rect) RelativePosition(pos Point) RelativePosition {
	switch {
	case pos.row < r.top:
		return Above
	case pos.row >= r.bot:
		return Below
	case pos.col < r.left:
		return LeftOf
	case pos.col >= r.right:
		return RightOf
	default:
		return Inside
	}
}

func (r Rect) IsPointAbove(p Point) bool {
	return r.top > p.row
}

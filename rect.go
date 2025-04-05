package main

// TODO change to have 4 int members and add method to get points
type Rect struct {
	top_left, bot_right Point
}

func (r Rect) Width() int {
	return r.bot_right.col - r.top_left.col
}

func (r Rect) Height() int {
	return r.bot_right.row - r.top_left.row
}

func (r Rect) Size() Size {
	return Size{width: r.Width(), height: r.Height()}
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

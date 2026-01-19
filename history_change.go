package main

type Change interface {
	Apply(win *Window)
	Reverse() Change
}

type EmptyChange struct{}

func (self EmptyChange) Apply(win *Window) {}
func (self EmptyChange) Reverse() Change   { return self }

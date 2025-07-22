package main

type Change interface {
	Apply(buf *Window)
	Reverse() Change
	Shift(index int) int
	IsEmpty() bool
}

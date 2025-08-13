package main

type Change interface {
	Apply(buf *Window)
	Reverse() Change
	IsEmpty() bool
}

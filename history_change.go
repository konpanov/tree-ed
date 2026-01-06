package main

type Change interface {
	Apply(win *Window)
	Reverse() Change
}

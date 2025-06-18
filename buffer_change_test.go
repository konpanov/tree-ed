package main

import (
	"testing"
)

func TestBufferChangeReverse(t *testing.T) {
	change := BufferChange{
		before:        []byte("1"),
		after:         []byte("s2\nlines3"),
		start_index:   4,
		new_end_index: 13,
		old_end_index: 5,
		start_pos:     Point{0, 4},
		new_end_pos:   Point{1, 6},
		old_end_pos:   Point{0, 5},
	}
	reverse := change.Reverse()
	expected := BufferChange{
		before:        []byte("s2\nlines3"),
		after:         []byte("1"),
		start_index:   4,
		new_end_index: 5,
		old_end_index: 13,
		start_pos:     Point{0, 4},
		new_end_pos:   Point{0, 5},
		old_end_pos:   Point{1, 6},
	}

	if !reverse.Equal(expected) {
		t.Error("Unexpected reverse of a buffer change")
	}
}

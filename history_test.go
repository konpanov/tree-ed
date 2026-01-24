package main

import (
	"slices"
	"testing"
)

func TestWindowHistoryReplacementChange(t *testing.T) {
	buffer := mkTestBuffer(t, "hello", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := NewReplacementChange(5, []byte{}, []byte(" bye"))
	change.Apply(window)
	window.history.Push(HistoryState{change})
	content, expected := buffer.Content(), []byte("hello bye")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	window.history.Back().Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	window.history.Forward().Apply(window)
	content, expected = buffer.Content(), []byte("hello bye")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowEmptyChange(t *testing.T) {
	buffer := mkTestBuffer(t, "hello", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := EmptyChange{}
	change.Apply(window)
	content, expected := buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	change.Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowHistoryForwardOnEmptyHistory(t *testing.T) {
	buffer := mkTestBuffer(t, "hello", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	window.history.Forward().Apply(window)
	content, expected := buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowCompositeChange(t *testing.T) {
	buffer := mkTestBuffer(t, "hello", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := CompositeChange{
		[]Change{
			NewReplacementChange(5, []byte{}, []byte(" bye")),
			NewReplacementChange(0, []byte{}, []byte("hi ")),
			NewReplacementChange(3, []byte("hell"), []byte{}),
			EmptyChange{},
			NewReplacementChange(3, []byte("o"), []byte("and")),
		},
	}
	change.Apply(window)
	content, expected := buffer.Content(), []byte("hi and bye")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	change.Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowCompositeSwapChange(t *testing.T) {
	buffer := mkTestBuffer(t, "xxxabcxxxjklxxx", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := NewSwapChange(window, 3, 6, 9, 12)
	change.Apply(window)
	content, expected := buffer.Content(), []byte("xxxjklxxxabcxxx")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	change.Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("xxxabcxxxjklxxx")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowCompositeSwapChangeReverseOrder(t *testing.T) {
	buffer := mkTestBuffer(t, "xxxabcxxxjklxxx", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := NewSwapChange(window, 9, 12, 3, 6)
	change.Apply(window)
	content, expected := buffer.Content(), []byte("xxxjklxxxabcxxx")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	change.Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("xxxabcxxxjklxxx")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

func TestWindowEraseRuneChange(t *testing.T) {
	buffer := mkTestBuffer(t, "hello", "\n")
	window := windowFromBuffer(buffer, 10, 10)
	change := NewEraseRuneChange(window, 2)
	change.Apply(window)
	content, expected := buffer.Content(), []byte("helo")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
	change.Reverse().Apply(window)
	content, expected = buffer.Content(), []byte("hello")
	if slices.Compare(content, expected) != 0 {
		t.Errorf("Unexpected content \"%+v\", expected \"%+v\"", string(content), string(expected))
	}
}

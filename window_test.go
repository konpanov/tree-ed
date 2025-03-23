package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestInsertEmptyContent(t *testing.T) {
	content := []byte("")
	buffer, _ := bufferNewFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.content, value)
}

func TestDeleteLinesAndInsertEmptyContent(t *testing.T) {
	content := []byte("line\nline\n")
	buffer, _ := bufferNewFromContent(content, []byte("\n"))
	window := windowFromBuffer(buffer, 10, 10)
	lines := window.buffer.Lines()
	window.deleteRange(Range{start: lines[1].start, end: lines[1].end + len(buffer.nl_seq)})
	window.deleteRange(Range{start: lines[0].start, end: lines[0].end + len(buffer.nl_seq)})
	value := []byte("hello")
	window.insert(value)
	assertBytesEqual(t, window.buffer.content, value)
}

func TestDrawEmptyContentInNormalMode(t *testing.T) {
	content := []byte("")
	screen, err := tcell.NewScreen()
	if err != nil {
		t.Fatalf("Could not create screen")
	}
	buffer, _ := bufferNewFromContent(content, []byte("\n"))
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	window.switchToNormal()
	window.draw(screen)
}

func TestDrawEmptyContentInInsertMode(t *testing.T) {
	content := []byte("")
	screen, err := tcell.NewScreen()
	if err != nil {
		t.Fatalf("Could not create screen")
	}
	buffer, _ := bufferNewFromContent(content, []byte("\n"))
	w, h := screen.Size()
	window := windowFromBuffer(buffer, w, h)
	window.switchToInsert()
	window.draw(screen)
}

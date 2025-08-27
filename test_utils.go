package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func mkTestScreen(t *testing.T, charset string) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen(charset)
	if s == nil {
		t.Fatalf("Failed to get simulation screen")
	}
	if e := s.Init(); e != nil {
		t.Fatalf("Failed to initialize screen: %v", e)
	}
	s.SetSize(20, 20)
	return s
}

func mkTestEditor(t *testing.T, size Point) *Editor {
	screen := mkTestScreen(t, "")
	screen.SetSize(size.col, size.row)
	editor := NewEditor(screen)
	return editor
}

func as_content(lines []string, nl string) []byte {
	return []byte(strings.Join(lines, nl))
}

func assertScreenRunes(t *testing.T, screen tcell.Screen, lines []string) {
	correct := true

	expected := borderedRunes(stringsToRunes(lines))
	received := borderedRunes(screenToRunes(screen))
	to_color := []Point{}
	for y, line := range expected {
		for x, r := range line {
			actual := received[y][x]
			if actual != r {
				to_color = append(to_color, Point{y, x})
				correct = false
			}
		}
	}
	if !correct {
		for i := len(to_color) - 1; i >= 0; i-- {
			highlightError(expected, to_color[i])
			highlightError(received, to_color[i])
		}
		expected := runesToString(expected)
		received := runesToString(received)
		t.Errorf("\nExpected screen: \n%s\nRecieved screen: \n%s\n", expected, received)
	}
}

func highlightError(runes [][]rune, pos Point) [][]rune {
	x, y := pos.col, pos.row
	pre := string(runes[y][:x])
	this := string(runes[y][x : x+1])
	post := string(runes[y][x+1:])
	runes[y] = []rune(fmt.Sprintf("%s%s%s", pre, redBg(this), post))
	return runes
}

func redBg(text string) string {
	return fmt.Sprintf("\033[41m%s\033[0m", text)
}

func screenToRunes(screen tcell.Screen) [][]rune {
	runes := [][]rune{}
	w, h := screen.Size()
	for y := range h {
		runes = append(runes, []rune{})
		for x := range w {
			primary, _, _, _ := screen.GetContent(x, y)
			runes[len(runes)-1] = append(runes[len(runes)-1], primary)
		}
	}
	return runes
}

func stringsToRunes(lines []string) [][]rune {
	output := [][]rune{}
	for _, line := range lines {
		output = append(output, []rune(line))
	}
	return output
}

func runesToString(runes [][]rune) string {
	output := []string{}
	for _, line := range runes {
		output = append(output, string(line))
	}
	return strings.Join(output, "\n")

}
func borderedRunes(runes [][]rune) [][]rune {
	hline := []rune{}
	for range len(runes[0]) + 2 {
		hline = append(hline, '-')
	}
	output := [][]rune{hline}
	for _, line := range runes {
		line := append(line, '|')
		line = append([]rune{'|'}, line...)
		output = append(output, line)
	}
	output = append(output, hline)
	return output
}

package main

import (
	"bytes"
	"cmp"
	"log"
	"runtime"
	"testing"

	"github.com/gdamore/tcell/v2"
)

const (
	NewLineWindows string = "\r\n"
	NewLineUnix    string = "\n"
)

func getSystemNewLine() []byte {
	switch runtime.GOOS {
	case "windows":
		return []byte(NewLineWindows)
	default:
		return []byte(NewLineUnix)
	}
}

func matchBytes(a []byte, b []byte) bool {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isInRange(value int, start int, end int) bool {
	return start <= value && value <= end
}

func order[T cmp.Ordered](a T, b T) (T, T) {
	return min(a, b), max(a, b)
}

func assertIntEqual(t *testing.T, a int, b int) {
	if a != b {
		t.Errorf("%d != %d", a, b)
	}
}

func assertIntEqualMsg(t *testing.T, a int, b int, msg string) {
	if a != b {
		t.Errorf("%s%d != %d", msg, a, b)
	}
}

func assertBytesEqual(t *testing.T, a []byte, b []byte) {
	if !bytes.Equal(a, b) {
		t.Errorf("%s != %s", string(a), string(b))
	}
}

func assertStringEqual(t *testing.T, a string, b string) {
	if a != b {
		t.Errorf("%s != %s", a, b)
	}
}

func assertPointsEqual(t *testing.T, result Point, expected Point) {
	if result != expected {
		t.Errorf("Recieved point does not match expected value %#v != %#v", result, expected)
	}
}

func assertNoErrors(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Got an unexpected error: %s", err)
	}
}

func assertNoErrorsMsg(t *testing.T, err error, msg string) {
	if err != nil {
		t.Errorf("%s%s", msg, err)
	}
}

func clip(value int, bot int, top int) int {
	return max(min(value, top), bot)
}

func rune_grid_to_string_slice(grid [][]rune) []string {
	ret := []string{}
	for _, line := range grid {
		ret = append(ret, string(line))
	}
	return ret
}

func list_colors() {
	for _, color := range tcell.ColorNames {
		log.Println(color)
	}
}

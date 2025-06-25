package main

import (
	"bytes"
	"cmp"
	"log"
	"runtime"
	"testing"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	NewLineWindows string = "\r\n"
	NewLineUnix    string = "\n"
	NewLineMac string = "\r"
)

func getContentNewLine(content []byte) []byte {
	nl_windows := []byte(NewLineWindows)
	nl_unix := []byte(NewLineUnix)
	nl_mac := []byte(NewLineMac)
	for i := range content{
		if matchBytes(content[i:], nl_windows){
			return nl_windows
		} else if matchBytes(content[i:], nl_unix) {
			return nl_unix
		} else if matchBytes(content[i:], nl_mac){
			return nl_mac
		}
	}
	return nl_unix
}

func getSystemNewLine() []byte {
	switch runtime.GOOS {
	case "windows":
		log.Println("Windows new lines")
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

func panic_if_error(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func rune_class(value rune) int {
	if unicode.IsSpace(value) {
		return 0
	} else if unicode.IsLetter(value) || unicode.IsDigit(value) || value == '_' {
		return 1
	} else if unicode.IsPunct(value) {
		return 2
	} else {
		return 3
	}
}

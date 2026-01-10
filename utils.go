package main

import (
	"bytes"
	"cmp"
	"fmt"
	"log"
	"runtime"
	"testing"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var (
	LineBreakWindows = []byte("\r\n")
	LineBreakPosix   = []byte("\n")
)
var LineBreaks = [][]byte{LineBreakWindows, LineBreakPosix}

func lineBreakDisplay(text []rune) []rune {
	res := []rune{}
	for _, r := range text {
		display, ok := map[rune][]rune{
			'\r': []rune("CR"),
			'\n': []rune("LF"),
		}[r]
		if !ok {
			display = []rune{'X'}
		}
		res = append(res, display...)
	}
	return res
}

func assert(is_valid bool, message string) {
	if debug && !is_valid {
		log.Panic(message)
	}
}

func getContentLineBreak(content []byte) []byte {
	for i := range content {
		for _, line_break := range LineBreaks {
			if matchBytes(content[i:], line_break) {
				return line_break
			}
		}
	}
	if runtime.GOOS == "windows" {
		return LineBreakWindows
	}
	return LineBreakPosix
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

func last[T any](stuff []T) T {
	return stuff[len(stuff)-1]
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
		t.Errorf("\"%s\" != \"%s\"", string(a), string(b))
	}
}

func assertStringEqual(t *testing.T, a string, b string) {
	if a != b {
		t.Errorf("%s != %s", a, b)
	}
}

func assertPositionsEqual(t *testing.T, result Pos, expected Pos) {
	if result != expected {
		t.Errorf("Recieved position does not match expected value %#v != %#v", result, expected)
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

func panic_if_error(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

type RuneClass int

const (
	RuneClassSpace RuneClass = iota
	RuneClassChar
	RuneClassPunct
	RuneClassOther
)

func rune_class(value rune) RuneClass {
	if unicode.IsSpace(value) {
		return RuneClassSpace
	} else if unicode.IsLetter(value) || unicode.IsDigit(value) || value == '_' {
		return RuneClassChar
	} else if unicode.IsPunct(value) {
		return RuneClassPunct
	} else {
		return RuneClassOther
	}
}

func IsLineBreak(content []byte) (bool, int) {
	for _, line_break := range LineBreaks {
		if matchBytes(content, line_break) {
			return true, len(line_break)
		}
	}
	return false, 0
}

func isLineBreakTerminated(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	is_lb, w := IsLineBreak(content[len(content)-1:])
	if is_lb {
		return true
	}
	if len(content) == 1 {
		return false
	}
	is_lb, w = IsLineBreak(content[len(content)-2:])
	return is_lb && w == 2
}

// No intersection
//          A___A \ A___A
//  B___B         \         B___B
// Intersection
//  A_______A     \     A___A     \     A_______A \ A___________A
//      B_______B \ B___________B \ B_______B     \     B___B

func isIntersection(startA int, endA int, startB int, endB int) bool {
	if debug && (startA > endA || startB > endB) {
		log.Printf(
			"regions in isIntersection call should be ordered: [%d, %d] and [%d, %d]\n",
			startA, endA, startB, endB,
		)
	}
	return !(endB <= startA || endA <= startB)
}

func KeyEventsToString(events []*tcell.EventKey) string {
	input := ""
	for _, key := range events {
		input += eventKeyToString(key)
	}
	return input
}

func e2a(ek *tcell.EventKey) string {
	return eventKeyToString(ek)
}
func eventKeyToString(ek *tcell.EventKey) string {
	out := ""
	if ek.Modifiers()&tcell.ModShift != 0 {
		out += "Shift "
	}
	if ek.Key() != tcell.KeyRune {
		out += tcell.KeyNames[ek.Key()]
	} else {
		out += string(ek.Rune())
	}
	return out
}

func RuneKey(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

func StringToEvents(s string) []tcell.Event {
	events := []tcell.Event{}
	for _, r := range s {
		events = append(events, RuneKey(r))
	}
	return events
}

func IsDigitKey(key_event *tcell.EventKey) bool {
	return key_event.Key() == tcell.KeyRune && unicode.IsDigit(key_event.Rune())
}

func EventKeysToInteger(events []*tcell.EventKey) int {
	integer := 0
	for _, ek := range events {
		if !IsDigitKey(ek) {
			assert(false, fmt.Sprintf("Unexpected non digit key %+v when parsing integer", ek))
		}
		integer *= 10
		integer += int(ek.Rune() - '0')
	}
	return integer
}

func IsTextInputKey(key_event *tcell.EventKey) bool {
	key := key_event.Key()
	return Any(
		key == tcell.KeyRune,
		key == tcell.KeyEnter,
		key == tcell.KeyLF,
		key == tcell.KeyTab,
	)
}

func Any(conditions ...bool) bool {
	for _, cond := range conditions {
		if cond {
			return true
		}
	}
	return false
}

func RenderedRune(value rune) string {
	// switch value {
	// case '\r':
	// 	return string(SymborForCarriageReturn)
	// case '\n':
	// 	return string(SymborForLineFeed)
	// }
	return string(value)
}

package main

import (
	"bytes"
	"cmp"
	"fmt"
	"log"
	"testing"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	NewLineWindows string = "\r\n"
	NewLineUnix    string = "\n"
	NewLineMac     string = "\r"
)

const (
	SymborForLineFeed       rune = 0x240A // LF \n
	SymborForCarriageReturn rune = 0x240D // CR \r
)

func newlinesToSymbols(text []rune) []rune {
	for i, r := range text {
		switch r {
		case '\r':
			text[i] = SymborForCarriageReturn
		case '\n':
			text[i] = SymborForLineFeed
		}
	}
	return text
}

func assert(is_valid bool, message string) {
	if debug && !is_valid {
		log.Panic(message)
	}
}

func getContentNewLine(content []byte) []byte {
	nl_windows := []byte(NewLineWindows)
	nl_unix := []byte(NewLineUnix)
	nl_mac := []byte(NewLineMac)
	for i := range content {
		if matchBytes(content[i:], nl_windows) {
			return nl_windows
		} else if matchBytes(content[i:], nl_unix) {
			return nl_unix
		} else if matchBytes(content[i:], nl_mac) {
			return nl_mac
		}
	}
	return nl_unix
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

func isNewLine(content []byte) (bool, int) {
	if matchBytes(content, []byte(NewLineWindows)) {
		return true, len([]byte(NewLineWindows))
	}
	if matchBytes(content, []byte(NewLineUnix)) {
		return true, len([]byte(NewLineUnix))
	}
	if matchBytes(content, []byte(NewLineMac)) {
		return true, len([]byte(NewLineMac))
	}
	return false, 0
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

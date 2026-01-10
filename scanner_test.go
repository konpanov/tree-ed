package main

import (
	"testing"
)

func TestScannerState(t *testing.T) {
	scanner := &Scanner{}
	scanner.Push(RuneKey('j'))
	if scanner.IsEnd() {
		t.Errorf("Scanner state is at the end, when expected to be on 'j'\n")
	}
	curr := scanner.Peek()
	if curr.Rune() != 'j' {
		t.Errorf("Expected curr to be rune key event with rune 'j', but got key event %+v\n", curr)
	}
	curr = scanner.Advance()
	if !scanner.IsEnd() {
		t.Errorf("Expected scanner to be at the end")
	}
}

func TestNormalScannerScanCursorDown(t *testing.T) {
	ek := RuneKey('j')
	scanner := NewScanner()
	scanner.mode = NormalMode
	scanner.Push(ek)
	op, res := scanner.Scan()
	if expected := ScanFull; res != expected {
		t.Errorf("Unexpected scan result %+v, expected %+v\n", res, expected)
	}
	_, is_expected := op.(OpCursorDown)
	if !is_expected {
		t.Errorf("Expected NormalCursorDown operation, but got %T: %+v\n", op, op)
	}
	scanner.Clear()
	if input := scanner.Input(); len(input) != 0 {
		t.Errorf("Expected input to be empty, but got %+v\n", input)
	}
}

func TestScanIntegerNumberPartial(t *testing.T) {
	scanner := &Scanner{}
	for _, ev := range StringToEvents("123") {
		scanner.Push(ev)
	}
	result := scanner.ScanZeroOrMore(scanner.ScanDigit)
	if expected := ScanStop; result != expected {
		t.Errorf("Unexpected integer scan result %+v, expected %+v\n", result, expected)
	}
	integer := EventKeysToInteger(scanner.Scanned())
	if expected := 123; integer != expected {
		t.Errorf("Unexpected scanned integer %d, expected %d\n", integer, expected)
	}
}

func TestScanIntegerNumberFull(t *testing.T) {
	scanner := &Scanner{}
	for _, ev := range StringToEvents("123j") {
		scanner.Push(ev)
	}
	result := scanner.ScanZeroOrMore(scanner.ScanDigit)
	if expected := ScanFull; result != expected {
		t.Errorf("Unexpected integer scan result %+v, expected %+v\n", result, expected)
	}
	integer := EventKeysToInteger(scanner.Scanned())
	if expected := 123; integer != expected {
		t.Errorf("Unexpected scanned integer %d, expected %d\n", integer, expected)
	}

	if scanner.IsEnd() {
		t.Errorf("Did not expected scanner state to be at the end")
	}
	ev := scanner.Peek()
	if expected, actual := string('j'), string(ev.Rune()); actual != expected {
		t.Errorf("Unexpected current event rune %s, expected %s\n", actual, expected)
	}
}

func TestScanIntegerLeadingZero(t *testing.T) {
	scanner := &Scanner{}
	for _, ev := range StringToEvents("000123j") {
		scanner.Push(ev)
	}
	result := scanner.ScanZeroOrMore(scanner.ScanDigit)
	if expected := ScanFull; result != expected {
		t.Errorf("Unexpected integer scan result %+v, expected %+v\n", result, expected)
	}
	integer := EventKeysToInteger(scanner.Scanned())
	if expected := 123; integer != expected {
		t.Errorf("Unexpected scanned integer %d, expected %d\n", integer, expected)
	}
	if scanner.IsEnd() {
		t.Errorf("Did not expected scanner state to be at the end")
	}
	ev := scanner.Peek()
	if expected, actual := string('j'), string(ev.Rune()); actual != expected {
		t.Errorf("Unexpected current event rune %s, expected %s\n", actual, expected)
	}
}

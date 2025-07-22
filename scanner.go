package main

import (
	"fmt"
	"log"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var ErrNotAnEventKey = fmt.Errorf("Accepting only event keys")
var ErrAmbiguous = fmt.Errorf("Ambiguous sequence")
var ErrNoMatch = fmt.Errorf("No match for sequence")
var ErrNoKey = fmt.Errorf("No more keys to scan")

type KeyTable map[tcell.Key]Operation
type RuneTable map[rune]Operation

type ScannerState struct {
	history []*tcell.EventKey
	curr    int
}

func (self *ScannerState) Clear() {
	self.history = self.history[self.curr:]
}

func (self *ScannerState) Advance() (*tcell.EventKey, error) {
	if self.curr < len(self.history) {
		self.curr++
	} else {
		return nil, ErrNoKey
	}
	return self.Curr()
}

func (self *ScannerState) Curr() (*tcell.EventKey, error) {
	if self.curr >= len(self.history) {
		return nil, ErrNoKey
	}
	return self.history[self.curr], nil
}

func (self *ScannerState) Push(ev tcell.Event) error {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return ErrNotAnEventKey
	}
	self.history = append(self.history, key_event)
	return nil
}

func (self *ScannerState) Reset() {
	self.curr = 0
}

type Scanner interface {
	Scan(ev tcell.Event) (Operation, error)
}

func IsDigit(key_event *tcell.EventKey) bool {
	return key_event.Key() == tcell.KeyRune && unicode.IsDigit(key_event.Rune())
}

func ScanCount(self *ScannerState) (int, error) {
	count := 0
	ev, err := self.Curr()
	if err != nil {
		return count, err
	}
	if !IsDigit(ev) || ev.Rune() == '0' {
		return count, ErrNoMatch
	}
	for ; err == nil && IsDigit(ev); ev, err = self.Advance() {
		count = count*10 + int(ev.Rune()) - int('0')
	}
	return count, nil
}

type GlobalScanner struct{}

func (self GlobalScanner) Scan(ev tcell.Event) (Operation, error) {
	if paste_event, ok := ev.(*tcell.EventClipboard); ok {
		log.Println("Got clipboard event")
		log.Println(string(paste_event.Data()))
		// return PasteClipboardOperation{data: paste_event.Data()}, nil
	}
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	if key_event.Key() == tcell.KeyCtrlC {
		return QuitOperation{}, nil
	}
	return nil, ErrNoMatch
}

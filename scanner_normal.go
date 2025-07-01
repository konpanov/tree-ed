package main

import (
	"fmt"
	"log"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var ErrNoKey = fmt.Errorf("No more keys to parse")

type NormalScannerTokenType int
const (
	DigitToken NormalScannerTokenType = iota
	KeyToken
)

type NormalParser struct{
	history []*tcell.EventKey
	curr    int
}

func (self *NormalParser) Parse(ev tcell.Event) (Operation, error) {
	key_event, ok := ev.(*tcell.EventKey)
	if !ok {
		return nil, ErrNotAnEventKey
	}
	self.history = append(self.history, key_event)
	log.Println("history: ", self.history)
	for i := 0; i < len(self.history); i++{
		log.Println(string(self.history[i].Rune()))
	}
	var op Operation
	self.curr = 0
	key_event, err := self.Curr()
	if key_event.Key() == tcell.KeyRune && unicode.IsDigit(key_event.Rune()){
		self.curr = 0
		log.Println("Trying to parse digit")
		op, err = self.ParseDigit()
	} else {
		self.curr = 0
		log.Println("Trying to operation")
		op, err = self.ParseOperation()
	}
	if err == ErrNoMatch {
		self.history = self.history[self.curr+1:]
	}
	if err == nil {
		self.history = self.history[self.curr:]
	}
	return op, err
}

func (self *NormalParser) ParseDigit() (Operation, error){
	key_event, err := self.Curr()
	if err != nil {
		return nil, err
	}
	count := 0
	for key_event.Key() == tcell.KeyRune && unicode.IsDigit(key_event.Rune()) {
		count = count * 10 + int(key_event.Rune()) - int('0')
		key_event, err = self.Advance()
		if err != nil {
			log.Println("Failed to parse number: ", err)
			return nil, err
		}
	}
	log.Println("Parsed number: ", count)
	op, err := self.ParseOperation()
	if err != nil {
		log.Println("Failed to parse operation after digit: ", err)
		return op, err
	}
	return CountOperation{count: count, op: op}, nil
}

func (self *NormalParser) Clear() {
	self.history = self.history[self.curr:]
}

func (self *NormalParser) Advance() (*tcell.EventKey, error) {
	self.curr++
	return self.Curr()
}

func (self *NormalParser) Curr() (*tcell.EventKey, error) {
	if self.curr >= len(self.history) {
		return nil, ErrNoKey
	}
	return self.history[self.curr], nil
}

func (self *NormalParser) ParseOperation() (Operation, error) {
	key_event, err := self.Curr()
	if err != nil {
		return nil, err
	}
	if key_event.Key() == tcell.KeyRune {
		switch key_event.Rune() {
		// Navigation
		case 'j':
			self.Advance()
			return NormalCursorDown{}, nil
		case 'k':
			self.Advance()
			return NormalCursorUp{}, nil
		case 'h':
			self.Advance()
			return NormalCursorLeft{}, nil
		case 'l':
			self.Advance()
			return NormalCursorRight{}, nil
		case 'w':
			self.Advance()
			return WordForwardOperation{}, nil
		case 'b':
			self.Advance()
			return WordBackwardOperation{}, nil
		// Modification
		case 'd':
			self.Advance()
			return EraseLineAtCursor{}, nil
		case 'x':
			self.Advance()
			return EraseCharNormalMode{}, nil
		// Modes
		case 'a':
			self.Advance()
			return SwitchToInsertModeAsAppend{}, nil
		case 'i':
			self.Advance()
			return SwitchToInsertMode{}, nil
		case 'v':
			self.Advance()
			return SwitchToVisualmode{}, nil
		case 't':
			self.Advance()
			return SwitchToTreeMode{}, nil
		case 'u':
			self.Advance()
			return UndoChangeOperation{}, nil
		}
	}
	switch key_event.Key() {
	case tcell.KeyCtrlR:
		self.Advance()
		return RedoChangeOperation{}, nil
	}
	return nil, ErrNoMatch
}

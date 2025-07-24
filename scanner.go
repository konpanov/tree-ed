package main

import (
	"fmt"
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

// Returns EventKey at position curr or ErrNoKey if curr is further than history
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
	Scan() (Operation, error)
	Push(ev tcell.Event)
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

type GlobalScanner struct {
	state *ScannerState
}

func (self *GlobalScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *GlobalScanner) Scan() (Operation, error) {
	var op Operation
	ek, err := self.state.Curr()
	if err == nil {
		switch {
		case ek.Key() == tcell.KeyCtrlC:
			self.state.Advance()
			op = QuitOperation{}
		default:
			err = ErrNoMatch
		}
	}
	return op, err
}

type OmniScanner struct {
	state          *ScannerState
	global_scanner *GlobalScanner
	normal_scanner *NormalScanner
	insert_scanner *InsertScanner
	visual_scanner *VisualScanner
	tree_scanner   *TreeScanner
	mode           WindowMode
}

func NewOmniScanner() *OmniScanner {
	state := &ScannerState{}
	return &OmniScanner{
		state:          state,
		global_scanner: &GlobalScanner{state: state},
		normal_scanner: &NormalScanner{state: state},
		insert_scanner: &InsertScanner{state: state},
		visual_scanner: &VisualScanner{state: state},
		tree_scanner:   &TreeScanner{state: state},
		mode:           NormalMode,
	}
}

func (self *OmniScanner) Push(ev tcell.Event) {
	self.state.Push(ev)
}

func (self *OmniScanner) Scan() (Operation, error) {
	op, _ := self.global_scanner.Scan()
	if op != nil {
		return op, nil
	}
	switch self.mode {
	case NormalMode:
		return self.normal_scanner.Scan()
	case InsertMode:
		return self.insert_scanner.Scan()
	case VisualMode:
		return self.visual_scanner.Scan()
	case TreeMode:
		return self.tree_scanner.Scan()
	}
	panic("Unkown mode")
	return nil, nil
}

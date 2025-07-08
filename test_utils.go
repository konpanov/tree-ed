package main

import (
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


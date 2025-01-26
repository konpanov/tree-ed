package main

import (
	"runtime"
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

func order(a int, b int) (int, int) {
	return min(a,b), max(a,b)
}

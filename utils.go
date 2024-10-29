package main

import (
	"os"
)

func readFile(filename string) string {
	dat, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(dat[:len(dat)-2])
}

func matchBytes(a []byte, b []byte) bool {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

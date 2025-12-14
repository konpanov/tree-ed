//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func LoadLanguageDynamicly(dll_path string, func_name string) (*sitter.Language, error) {
	var language func() uintptr
	handle, err := syscall.LoadLibrary(dll_path)
	if err != nil {
		return nil, err
	}
	lib, err := uintptr(handle), err
	purego.RegisterLibFunc(&language, lib, func_name)
	return sitter.NewLanguage(unsafe.Pointer(language())), nil
}

//go:build !windows

package main

import (
	"unsafe"

	"github.com/ebitengine/purego"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func LoadLanguageDynamicly(path string, func_name string) (*sitter.Language, error) {
	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}

	var language func() uintptr
	purego.RegisterLibFunc(&language, lib, func_name)
	return sitter.NewLanguage(unsafe.Pointer(language())), nil
}

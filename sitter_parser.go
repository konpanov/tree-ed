package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	sitter_c_sharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	sitter_erb "github.com/tree-sitter/tree-sitter-embedded-template/bindings/go"
	sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	sitter_hs "github.com/tree-sitter/tree-sitter-haskell/bindings/go"
	sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	sitter_js "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	sitter_json "github.com/tree-sitter/tree-sitter-json/bindings/go"
	sitter_julia "github.com/tree-sitter/tree-sitter-julia/bindings/go"
	sitter_ocaml "github.com/tree-sitter/tree-sitter-ocaml/bindings/go"
	sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	sitter_scala "github.com/tree-sitter/tree-sitter-scala/bindings/go"
	sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func GetFiletype(filename string) string {
	return last(strings.Split(filename, "."))
}

type LanguageConfigEntry struct {
	Path       string   `json:"path"`
	FuncName   string   `json:"func_name"`
	Extensions []string `json:"extensions"`
}

func LoadLanguageFromConfig(filetype string) (*sitter.Language, error) {
	config_filename := "config.json"
	content, err := os.ReadFile(config_filename)
	if err != nil {
		return nil, fmt.Errorf("Failed find config file named %s", config_filename)
	}
	var config []LanguageConfigEntry
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config file %s", config_filename)
	}

	for _, entry := range config {
		if slices.Contains(entry.Extensions, filetype) {
			return LoadLanguageDynamicly(entry.Path, entry.FuncName)
		}
	}
	return nil, fmt.Errorf("Failed to match filetype %s to extensions from config file %s", filetype, config_filename)

}

func ParserLanguageByFileType(filetype string) *sitter.Language {
	lang, err := LoadLanguageFromConfig(filetype)
	if err == nil {
		return lang
	}

	switch filetype {
	case "go":
		return sitter.NewLanguage(sitter_go.Language())
	case "js":
		return sitter.NewLanguage(sitter_js.Language())
	case "bash", "sh":
		return sitter.NewLanguage(sitter_bash.Language())
	case "cpp", "cc", "cxx", "C", "hpp", "hh", "hxx":
		return sitter.NewLanguage(sitter_cpp.Language())
	case "c", "h", "i", "o", "a", "so":
		return sitter.NewLanguage(sitter_c.Language())
	case "cs", "csx":
		return sitter.NewLanguage(sitter_c_sharp.Language())
	case "erb", "ejs":
		return sitter.NewLanguage(sitter_erb.Language())
	case "hs", "lhs", "hs-boot":
		return sitter.NewLanguage(sitter_hs.Language())
	case "html", "htm":
		return sitter.NewLanguage(sitter_html.Language())
	case "java", "class", "jar":
		return sitter.NewLanguage(sitter_java.Language())
	case "json":
		return sitter.NewLanguage(sitter_json.Language())
	case "jl", "jmd":
		return sitter.NewLanguage(sitter_julia.Language())
	case "ml":
		return sitter.NewLanguage(sitter_ocaml.LanguageOCaml())
	case "mli":
		return sitter.NewLanguage(sitter_ocaml.LanguageOCamlInterface())
	case "mlt":
		return sitter.NewLanguage(sitter_ocaml.LanguageOCamlType())
	case "php":
		return sitter.NewLanguage(sitter_php.LanguagePHP())
	case "py":
		return sitter.NewLanguage(sitter_python.Language())
	case "ruby":
		return sitter.NewLanguage(sitter_ruby.Language())
	case "rs":
		return sitter.NewLanguage(sitter_rust.Language())
	case "scala", "sc":
		return sitter.NewLanguage(sitter_scala.Language())
	case "ts":
		return sitter.NewLanguage(sitter_typescript.LanguageTypescript())
	case "tsx":
		return sitter.NewLanguage(sitter_typescript.LanguageTSX())
	}
	return nil
}

package wrap

import (
	"path/filepath"
	"strings"
)

// Language defines comment syntax for a programming language.
type Language struct {
	Name        string
	Extensions  []string
	LineMarkers []string // e.g., "//", "#"
	BlockStart  []string // e.g., "/*"
	BlockEnd    []string // e.g., "*/"
	BlockPrefix string   // e.g., " * " for JavaDoc-style
	Directives  []string // prefixes (after line marker) that indicate a directive, not a comment
}

var languages = []Language{
	{
		Name:        "go",
		Extensions:  []string{".go"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
		Directives:  []string{"go:", "line ", "export ", "nolint"},
	},
	{
		Name:        "c",
		Extensions:  []string{".c", ".h"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
	},
	{
		Name:        "cpp",
		Extensions:  []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
	},
	{
		Name:        "java",
		Extensions:  []string{".java"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
		BlockPrefix: " * ",
	},
	{
		Name:        "javascript",
		Extensions:  []string{".js", ".jsx", ".mjs", ".cjs"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
	},
	{
		Name:        "typescript",
		Extensions:  []string{".ts", ".tsx", ".mts", ".cts"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
	},
	{
		Name:        "python",
		Extensions:  []string{".py"},
		LineMarkers: []string{"#"},
	},
	{
		Name:        "shell",
		Extensions:  []string{".sh", ".bash", ".zsh"},
		LineMarkers: []string{"#"},
	},
	{
		Name:        "ruby",
		Extensions:  []string{".rb"},
		LineMarkers: []string{"#"},
	},
	{
		Name:        "rust",
		Extensions:  []string{".rs"},
		LineMarkers: []string{"//"},
		BlockStart:  []string{"/*"},
		BlockEnd:    []string{"*/"},
	},
	{
		Name:       "markdown",
		Extensions: []string{".md", ".markdown"},
	},
}

// extensionMap is built at init time for fast lookup.
var extensionMap map[string]*Language

func init() {
	extensionMap = make(map[string]*Language)
	for i := range languages {
		for _, ext := range languages[i].Extensions {
			extensionMap[ext] = &languages[i]
		}
	}
}

// LanguageFromExtension returns the language for the given file extension (including the dot).
// Returns nil if no language matches.
func LanguageFromExtension(ext string) *Language {
	return extensionMap[strings.ToLower(ext)]
}

// LanguageFromFilename returns the language for the given filename.
func LanguageFromFilename(filename string) *Language {
	return LanguageFromExtension(filepath.Ext(filename))
}

// LanguageFromName returns the language by its name (case-insensitive).
func LanguageFromName(name string) *Language {
	lower := strings.ToLower(name)
	for i := range languages {
		if languages[i].Name == lower {
			return &languages[i]
		}
	}
	return nil
}

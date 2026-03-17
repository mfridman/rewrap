package wrap

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// filenamePattern extracts the column width from filenames like "go_comments_c60.go".
var filenamePattern = regexp.MustCompile(`_c(\d+)\.`)

func TestGolden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*_c[0-9]*.*")
	require.NoError(t, err)
	// Filter out golden files.
	var testFiles []string
	for _, f := range inputs {
		if !isGoldenFile(f) {
			testFiles = append(testFiles, f)
		}
	}
	require.NotEmpty(t, testFiles, "no test input files found in testdata/")

	for _, inputPath := range testFiles {
		name := filepath.Base(inputPath)
		t.Run(name, func(t *testing.T) {
			// Extract column width from filename.
			matches := filenamePattern.FindStringSubmatch(name)
			require.GreaterOrEqual(t, len(matches), 2, "cannot extract column width from filename: %s", name)
			column, err := strconv.Atoi(matches[1])
			require.NoError(t, err, "invalid column width in filename: %s", matches[1])

			src, err := os.ReadFile(inputPath)
			require.NoError(t, err)

			// Determine language from extension.
			ext := filepath.Ext(inputPath)
			var lang *Language
			if ext != ".txt" {
				lang = LanguageFromExtension(ext)
			}

			got := Source(src, lang, column, 4)

			goldenPath := goldenFilePath(inputPath)
			if *update {
				require.NoError(t, os.WriteFile(goldenPath, got, 0o644))
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			require.NoError(t, err, "golden file not found (run with -update to create)")
			assert.Equal(t, string(want), string(got), "output does not match golden file %s", goldenPath)
		})
	}
}

func TestIdempotent(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*_c[0-9]*.*")
	require.NoError(t, err)
	var testFiles []string
	for _, f := range inputs {
		if !isGoldenFile(f) {
			testFiles = append(testFiles, f)
		}
	}
	require.NotEmpty(t, testFiles, "no test input files found in testdata/")

	for _, inputPath := range testFiles {
		name := filepath.Base(inputPath)
		t.Run(name, func(t *testing.T) {
			matches := filenamePattern.FindStringSubmatch(name)
			require.GreaterOrEqual(t, len(matches), 2, "cannot extract column width from filename: %s", name)
			column, err := strconv.Atoi(matches[1])
			require.NoError(t, err, "invalid column width in filename: %s", matches[1])

			src, err := os.ReadFile(inputPath)
			require.NoError(t, err)

			ext := filepath.Ext(inputPath)
			var lang *Language
			if ext != ".txt" {
				lang = LanguageFromExtension(ext)
			}

			pass1 := Source(src, lang, column, 4)
			pass2 := Source(pass1, lang, column, 4)
			assert.Equal(t, string(pass1), string(pass2), "output is not idempotent")
		})
	}
}

// TestGofmtCompatible verifies that rewrap's Go output is already gofmt-formatted, meaning running
// gofmt on a golden file produces no changes.
func TestGofmtCompatible(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not found in PATH")
	}
	goldens, err := filepath.Glob("testdata/*.go.golden")
	require.NoError(t, err)
	require.NotEmpty(t, goldens, "no .go.golden files found in testdata/")

	for _, goldenPath := range goldens {
		name := filepath.Base(goldenPath)
		t.Run(name, func(t *testing.T) {
			golden, err := os.ReadFile(goldenPath)
			require.NoError(t, err)

			cmd := exec.Command("gofmt", goldenPath)
			formatted, err := cmd.Output()
			require.NoError(t, err, "gofmt failed")

			assert.Equal(t, string(golden), string(formatted),
				"golden file is not gofmt-formatted; run gofmt on it or fix the rewrap output",
			)
		})
	}
}

func isGoldenFile(path string) bool {
	name := filepath.Base(path)
	return filepath.Ext(name) == ".golden" || strings.Contains(name, ".golden.")
}

// goldenFilePath returns the golden file path for the given input path. Go files use
// "<name>.go.golden" to avoid being compiled, all others use "<name>.golden.<ext>" so
// editors can preview them with proper syntax highlighting.
func goldenFilePath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == ".go" {
		return inputPath + ".golden"
	}
	return strings.TrimSuffix(inputPath, ext) + ".golden" + ext
}

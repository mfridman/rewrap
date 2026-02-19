package wrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource_PlainText(t *testing.T) {
	input := "this is a long line of text that should be wrapped at a narrow column width for testing purposes\n"
	got := string(Source([]byte(input), nil, 40, 4))
	// Should wrap and preserve trailing newline.
	require.NotEmpty(t, got)
	assert.Equal(t, byte('\n'), got[len(got)-1], "trailing newline not preserved")
	// No line should exceed 40 characters.
	for i, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		assert.LessOrEqual(t, displayWidth(line, 4), 40, "line %d exceeds column width: %q", i, line)
	}
}

func TestSource_GoComments(t *testing.T) {
	goLang := LanguageFromName("go")
	input := `package main

// This is a very long comment that should be rewrapped because it exceeds the typical column width of eighty characters.

func main() {}
`
	got := string(Source([]byte(input), goLang, 60, 4))
	lines := strings.Split(got, "\n")
	// The comment should now be multiple lines.
	commentCount := 0
	for _, line := range lines {
		if len(line) > 2 && line[:2] == "//" {
			commentCount++
		}
	}
	assert.GreaterOrEqual(t, commentCount, 2,
		"expected comment to be wrapped into multiple lines, got %d comment lines\noutput:\n%s", commentCount, got)
}

package wrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSegments(t *testing.T) {
	goLang := LanguageFromName("go")
	require.NotNil(t, goLang, "go language not found")

	t.Run("line comments", func(t *testing.T) {
		input := strings.Split("// hello\n// world\nfunc main() {}", "\n")
		segs := parseSegments(input, goLang)
		require.Len(t, segs, 2)
		assert.Equal(t, segmentComment, segs[0].typ)
		assert.Len(t, segs[0].lines, 2)
		assert.Equal(t, segmentCode, segs[1].typ)
	})

	t.Run("block comment", func(t *testing.T) {
		input := strings.Split("/*\n * hello\n */\nfunc main() {}", "\n")
		segs := parseSegments(input, goLang)
		require.Len(t, segs, 2)
		assert.Equal(t, segmentBlock, segs[0].typ)
	})

	t.Run("indented comments", func(t *testing.T) {
		input := strings.Split("\t// hello\n\t// world", "\n")
		segs := parseSegments(input, goLang)
		require.Len(t, segs, 1)
		assert.Equal(t, "\t", segs[0].indent)
	})

	t.Run("mixed code and comments", func(t *testing.T) {
		input := strings.Split("package main\n\n// Comment\nfunc foo() {}\n\n// Another\nfunc bar() {}", "\n")
		segs := parseSegments(input, goLang)
		// Should have: code, comment, code, comment, code
		wantTypes := []segmentType{segmentCode, segmentComment, segmentCode, segmentComment, segmentCode}
		require.Len(t, segs, len(wantTypes))
		for i, seg := range segs {
			assert.Equal(t, wantTypes[i], seg.typ, "segment %d", i)
		}
	})
}

func TestIsDecorationLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"========", true},
		{"--------", true},
		{"***", true},
		{"hello", false},
		{"== hello ==", false},
		{"", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isDecorationLine(tt.input), "isDecorationLine(%q)", tt.input)
	}
}

package wrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		prefix           string
		subsequentPrefix string
		columnWidth      int
		tabWidth         int
		want             []string
	}{
		{
			name:             "empty",
			text:             "",
			prefix:           "// ",
			subsequentPrefix: "// ",
			columnWidth:      40,
			tabWidth:         4,
			want:             []string{"// "},
		},
		{
			name:             "short line no wrap",
			text:             "hello world",
			prefix:           "// ",
			subsequentPrefix: "// ",
			columnWidth:      40,
			tabWidth:         4,
			want:             []string{"// hello world"},
		},
		{
			name:             "wraps at column width",
			text:             "one two three four five six seven eight nine ten",
			prefix:           "// ",
			subsequentPrefix: "// ",
			columnWidth:      20,
			tabWidth:         4,
			want: []string{
				"// one two three",
				"// four five six",
				"// seven eight nine",
				"// ten",
			},
		},
		{
			name:             "preserves paragraph breaks",
			text:             "first paragraph words\n\nsecond paragraph words",
			prefix:           "// ",
			subsequentPrefix: "// ",
			columnWidth:      40,
			tabWidth:         4,
			want: []string{
				"// first paragraph words",
				"//",
				"// second paragraph words",
			},
		},
		{
			name:             "no prefix plain text",
			text:             "some words that need to be wrapped at a narrow width",
			prefix:           "",
			subsequentPrefix: "",
			columnWidth:      20,
			tabWidth:         4,
			want: []string{
				"some words that need",
				"to be wrapped at a",
				"narrow width",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.prefix, tt.subsequentPrefix, tt.columnWidth, tt.tabWidth)
			require.Len(t, got, len(tt.want), "got:\n%s\nwant:\n%s",
				strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			for i := range got {
				assert.Equal(t, tt.want[i], got[i], "line %d", i)
			}
		})
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		s        string
		tabWidth int
		want     int
	}{
		{"hello", 4, 5},
		{"\t", 4, 4},
		{"a\tb", 4, 5}, // a at col 0, tab to col 4, b at col 4
		{"", 4, 0},
	}
	for _, tt := range tests {
		got := displayWidth(tt.s, tt.tabWidth)
		assert.Equal(t, tt.want, got, "displayWidth(%q, %d)", tt.s, tt.tabWidth)
	}
}

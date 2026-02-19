package wrap

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// processMarkdown rewraps paragraph text in Markdown source while preserving all structural
// elements (headings, code blocks, lists, blockquotes, tables, thematic breaks, HTML) verbatim.
func processMarkdown(src []byte, column, tabWidth int) []byte {
	// Normalize line endings.
	normalized := bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	reader := text.NewReader(normalized)
	md := goldmark.New(goldmark.WithExtensions(extension.Table))
	doc := md.Parser().Parse(reader)

	lines := strings.Split(string(normalized), "\n")

	// Build a set of line indices that belong to top-level paragraphs (0-indexed).
	type lineRange struct {
		start int // inclusive
		end   int // exclusive
	}
	var paragraphs []lineRange

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != ast.KindParagraph {
			continue
		}
		// Get the line range from the paragraph's text segments.
		segs := child.Lines()
		if segs.Len() == 0 {
			continue
		}
		firstSeg := segs.At(0)
		lastSeg := segs.At(segs.Len() - 1)
		startLine := byteOffsetToLine(normalized, firstSeg.Start)
		endLine := byteOffsetToLine(normalized, lastSeg.Stop-1) + 1
		paragraphs = append(paragraphs, lineRange{start: startLine, end: endLine})
	}

	// Build output by processing line ranges.
	var out []string
	i := 0
	for _, p := range paragraphs {
		// Pass through lines before this paragraph.
		for i < p.start && i < len(lines) {
			out = append(out, lines[i])
			i++
		}
		// Extract paragraph text and rewrap.
		paraLines := lines[p.start:p.end]
		joined := strings.Join(paraLines, "\n")
		wrapped := wrapText(joined, "", "", column, tabWidth)
		out = append(out, wrapped...)
		i = p.end
	}
	// Pass through remaining lines.
	for i < len(lines) {
		out = append(out, lines[i])
		i++
	}

	result := strings.Join(out, "\n")
	// Preserve trailing newline if original had one.
	if len(normalized) > 0 && normalized[len(normalized)-1] == '\n' && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return []byte(result)
}

// byteOffsetToLine converts a byte offset in src to a 0-indexed line number.
func byteOffsetToLine(src []byte, offset int) int {
	line := 0
	for i := 0; i < offset && i < len(src); i++ {
		if src[i] == '\n' {
			line++
		}
	}
	return line
}

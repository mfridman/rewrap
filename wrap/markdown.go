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
// elements (headings, code blocks, blockquotes, tables, thematic breaks, HTML) verbatim.
// Paragraphs inside list items are rewrapped with their marker/indentation preserved.
func processMarkdown(src []byte, column, tabWidth int) []byte {
	// Normalize line endings.
	normalized := bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	reader := text.NewReader(normalized)
	md := goldmark.New(goldmark.WithExtensions(extension.Table))
	doc := md.Parser().Parse(reader)

	lines := strings.Split(string(normalized), "\n")

	type paragraphInfo struct {
		start       int    // inclusive line number (0-indexed)
		end         int    // exclusive line number
		firstPrefix string // prefix for first wrapped line
		contPrefix  string // prefix for continuation wrapped lines
		text        string // text content from segments (markers stripped)
	}
	var paragraphs []paragraphInfo

	// Walk the full AST to find paragraphs at any nesting depth.
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || (node.Kind() != ast.KindParagraph && node.Kind() != ast.KindTextBlock) {
			return ast.WalkContinue, nil
		}
		segs := node.Lines()
		if segs.Len() == 0 {
			return ast.WalkContinue, nil
		}

		parent := node.Parent()
		if parent == nil {
			return ast.WalkContinue, nil
		}

		// Only process paragraphs in contexts we understand.
		firstSeg := segs.At(0)
		lastSeg := segs.At(segs.Len() - 1)

		// For non-document parents, derive the prefix from source.
		lineStart := lineStartOffset(normalized, firstSeg.Start)
		srcPrefix := string(normalized[lineStart:firstSeg.Start])

		var firstPrefix, contPrefix string
		switch parent.Kind() {
		case ast.KindDocument:
			// Top-level paragraph, no prefix.
		case ast.KindBlockquote:
			firstPrefix = srcPrefix
			contPrefix = srcPrefix
		case ast.KindListItem:
			firstPrefix = srcPrefix
			// Preserve blockquote markers ("> ") in continuation prefix, replacing
			// only the list-marker portion with spaces.
			bqPrefix := blockquotePrefix(srcPrefix)
			contPrefix = bqPrefix + strings.Repeat(" ", displayWidth(srcPrefix, tabWidth)-displayWidth(bqPrefix, tabWidth))
		default:
			// Inside other structure - skip.
			return ast.WalkContinue, nil
		}
		startLine := byteOffsetToLine(normalized, firstSeg.Start)
		endLine := byteOffsetToLine(normalized, lastSeg.Stop-1) + 1

		// Extract text content from segments (markers already stripped by parser).
		var segTexts []string
		for i := 0; i < segs.Len(); i++ {
			seg := segs.At(i)
			segTexts = append(segTexts, strings.TrimRight(string(normalized[seg.Start:seg.Stop]), "\n\r"))
		}

		paragraphs = append(paragraphs, paragraphInfo{
			start:       startLine,
			end:         endLine,
			firstPrefix: firstPrefix,
			contPrefix:  contPrefix,
			text:        strings.Join(segTexts, "\n"),
		})
		return ast.WalkContinue, nil
	})

	// Build output by processing line ranges.
	var out []string
	i := 0
	for _, p := range paragraphs {
		// Pass through lines before this paragraph.
		for i < p.start && i < len(lines) {
			out = append(out, lines[i])
			i++
		}
		wrapped := wrapText(p.text, p.firstPrefix, p.contPrefix, column, tabWidth)
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

// blockquotePrefix returns the blockquote marker portion of a line prefix.
// For "> - " it returns "> ", for "> > - " it returns "> > ", and for "- " it returns "".
func blockquotePrefix(prefix string) string {
	end := 0
	for i := 0; i < len(prefix); i++ {
		if prefix[i] == '>' {
			if i+1 < len(prefix) && prefix[i+1] == ' ' {
				end = i + 2
			} else {
				end = i + 1
			}
		}
	}
	return prefix[:end]
}

// lineStartOffset returns the byte offset of the start of the line containing the given offset.
func lineStartOffset(src []byte, offset int) int {
	for i := offset - 1; i >= 0; i-- {
		if src[i] == '\n' {
			return i + 1
		}
	}
	return 0
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

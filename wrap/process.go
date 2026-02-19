package wrap

import (
	"bytes"
	"go/doc/comment"
	"strings"
)

// Source rewraps comment blocks in src according to the given language and column width. If lang is
// nil, the entire input is treated as plain text.
func Source(src []byte, lang *Language, column int, tabWidth int) []byte {
	text := string(src)
	// Normalize line endings.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")

	// Plain text mode: no language, wrap everything.
	if lang == nil {
		return []byte(wrapPlainText(lines, column, tabWidth))
	}

	// Markdown mode: use AST-based processing.
	if lang.Name == "markdown" {
		return processMarkdown(src, column, tabWidth)
	}

	segments := parseSegments(lines, lang)
	var out []string
	for _, seg := range segments {
		switch seg.typ {
		case segmentCode:
			out = append(out, seg.lines...)
		case segmentComment:
			out = append(out, rewrapLineComments(seg, lang, column, tabWidth)...)
		case segmentBlock:
			out = append(out, rewrapBlockComment(seg, lang, column, tabWidth)...)
		}
	}
	result := strings.Join(out, "\n")
	// Preserve trailing newline if original had one.
	if len(src) > 0 && src[len(src)-1] == '\n' && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return []byte(result)
}

// rewrapLineComments rewraps a block of consecutive line comments. Decoration lines (lines
// consisting entirely of repeated punctuation like //========) are preserved verbatim and act as
// boundaries between wrappable runs of text.
func rewrapLineComments(seg segment, lang *Language, column, tabWidth int) []string {
	// Extract comment text, stripping indent and marker.
	type commentLine struct {
		raw     string // original source line
		content string // text after stripping indent + marker
	}
	var lines []commentLine
	for _, line := range seg.lines {
		stripped := strings.TrimLeft(line, " \t")
		if len(seg.marker) <= len(stripped) {
			lines = append(lines, commentLine{raw: line, content: stripped[len(seg.marker):]})
		} else {
			// Marker-only line (e.g., bare "//"), treat as blank.
			lines = append(lines, commentLine{raw: line, content: ""})
		}
	}

	// Split into runs separated by decoration lines. Decoration lines are emitted verbatim.
	var out []string
	runStart := -1
	flush := func(end int) {
		if runStart < 0 || runStart >= end {
			return
		}
		var textLines []string
		for _, cl := range lines[runStart:end] {
			textLines = append(textLines, cl.content)
		}
		// For Go line comments using "//", use go/doc/comment for proper doc comment handling.
		if lang.Name == "go" && strings.TrimSpace(seg.marker) == "//" {
			out = append(out, rewrapGoDocComment(textLines, seg.indent, column, tabWidth)...)
		} else {
			joined := strings.Join(textLines, "\n")
			prefix := seg.indent + seg.marker
			out = append(out, wrapText(joined, prefix, prefix, column, tabWidth)...)
		}
		runStart = -1
	}
	for i, cl := range lines {
		if isDecorationLine(cl.content) {
			flush(i)
			out = append(out, seg.indent+seg.marker+cl.content)
		} else {
			if runStart < 0 {
				runStart = i
			}
		}
	}
	flush(len(lines))
	return out
}

// rewrapGoDocComment uses go/doc/comment to rewrap Go doc comments with proper understanding of doc
// comment syntax (links, lists, code blocks, headings).
func rewrapGoDocComment(textLines []string, indent string, column, tabWidth int) []string {
	// go/doc/comment expects the raw text without comment markers.
	docText := strings.Join(textLines, "\n")

	var p comment.Parser
	doc := p.Parse(docText)

	// TextWidth is the width available for text content (excluding the prefix).
	prefix := "// "
	prefixWidth := displayWidth(indent+prefix, tabWidth)
	pr := comment.Printer{
		TextPrefix: prefix,
		TextWidth:  column - prefixWidth,
	}
	result := pr.Text(doc)
	result = bytes.TrimRight(result, "\n")

	lines := strings.Split(string(result), "\n")
	// Add indentation if the comment was indented.
	if indent != "" {
		for i, line := range lines {
			if line != "" {
				lines[i] = indent + line
			}
		}
	}
	return lines
}

// rewrapBlockComment rewraps a block comment (/* ... */).
func rewrapBlockComment(seg segment, lang *Language, column, tabWidth int) []string {
	if len(seg.lines) == 0 {
		return seg.lines
	}

	// Single-line block comments: pass through.
	if len(seg.lines) == 1 {
		return seg.lines
	}

	startMarker := lang.BlockStart[0]
	endMarker := lang.BlockEnd[0]

	// Extract content lines between start and end markers.
	var textLines []string
	for i, line := range seg.lines {
		stripped := strings.TrimLeft(line, " \t")
		if i == 0 {
			// Remove start marker.
			after := strings.TrimPrefix(stripped, startMarker)
			after = strings.TrimSpace(after)
			if after != "" {
				textLines = append(textLines, after)
			}
			continue
		}
		if strings.Contains(line, endMarker) {
			// Last line - remove end marker.
			before, _, _ := strings.Cut(stripped, endMarker)
			before = strings.TrimSpace(before)
			// Remove leading * if present.
			before = strings.TrimPrefix(before, "*")
			before = strings.TrimSpace(before)
			if before != "" {
				textLines = append(textLines, before)
			}
			continue
		}
		// Middle lines - strip leading " * " or " *" prefix.
		content := stripped
		content = strings.TrimPrefix(content, "* ")
		if content == stripped {
			content = strings.TrimPrefix(content, "*")
		}
		textLines = append(textLines, content)
	}

	// Determine the prefix for wrapped lines.
	blockPrefix := lang.BlockPrefix
	if blockPrefix == "" {
		blockPrefix = " * "
	}
	innerPrefix := seg.indent + blockPrefix

	joined := strings.Join(textLines, "\n")
	wrapped := wrapText(joined, innerPrefix, innerPrefix, column, tabWidth)

	// Reconstruct block comment.
	var result []string
	result = append(result, seg.indent+startMarker)
	result = append(result, wrapped...)
	result = append(result, seg.indent+" "+endMarker)
	return result
}

// wrapPlainText wraps plain text (no comment markers) preserving paragraph breaks.
func wrapPlainText(lines []string, column, tabWidth int) string {
	joined := strings.Join(lines, "\n")
	wrapped := wrapText(joined, "", "", column, tabWidth)
	result := strings.Join(wrapped, "\n")
	// Preserve trailing newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		if !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
	}
	return result
}

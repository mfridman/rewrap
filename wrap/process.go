package wrap

import (
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
		if lang.Name == "go" && strings.TrimSpace(seg.marker) == "//" {
			var textLines []string
			for _, cl := range lines[runStart:end] {
				stripped := strings.TrimLeft(cl.raw, " \t")
				if strings.HasPrefix(stripped, "// ") {
					textLines = append(textLines, stripped[3:]) // strip "// "
				} else if strings.HasPrefix(stripped, "//\t") {
					textLines = append(textLines, stripped[2:]) // strip "//", keep tab
				} else {
					textLines = append(textLines, "")
				}
			}
			out = append(out, rewrapGoDocComment(textLines, seg.indent, column, tabWidth)...)
			runStart = -1
			return
		}
		var textLines []string
		for _, cl := range lines[runStart:end] {
			textLines = append(textLines, cl.content)
		}
		{
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

// rewrapGoDocComment rewraps Go doc comments using comment.Parser for structure detection, then
// renders each block directly to preserve original text content (whitespace, doc link brackets).
// The textLines parameter contains lines with "//" stripped (preserving leading space or tab).
func rewrapGoDocComment(textLines []string, indent string, column, tabWidth int) []string {
	prefix := indent + "// "
	bareMarker := indent + "//"

	// Count leading and trailing blank lines that comment.Parser would strip.
	leadingBlanks := 0
	for _, l := range textLines {
		if strings.TrimSpace(l) == "" {
			leadingBlanks++
		} else {
			break
		}
	}
	trailingBlanks := 0
	for i := len(textLines) - 1; i >= leadingBlanks; i-- {
		if strings.TrimSpace(textLines[i]) == "" {
			trailingBlanks++
		} else {
			break
		}
	}

	docText := strings.Join(textLines, "\n")

	var p comment.Parser
	doc := p.Parse(docText)

	if len(doc.Content) == 0 {
		var result []string
		for range leadingBlanks + trailingBlanks {
			result = append(result, bareMarker)
		}
		return result
	}

	var result []string
	for range leadingBlanks {
		result = append(result, bareMarker)
	}

	for i, block := range doc.Content {
		if i > 0 {
			// A list directly following a paragraph (no blank line) omits the separator
			// unless the list's ForceBlankBefore flag is set.
			addBlank := true
			if list, ok := block.(*comment.List); ok {
				if _, prevIsPara := doc.Content[i-1].(*comment.Paragraph); prevIsPara {
					addBlank = list.ForceBlankBefore
				}
			}
			if addBlank {
				result = append(result, bareMarker)
			}
		}
		switch b := block.(type) {
		case *comment.Paragraph:
			text := docInlineText(b.Text)
			result = append(result, wrapText(text, prefix, prefix, column, tabWidth)...)
		case *comment.Code:
			lines := strings.Split(strings.TrimRight(b.Text, "\n"), "\n")
			for _, line := range lines {
				if line == "" {
					result = append(result, bareMarker)
				} else {
					result = append(result, bareMarker+"\t"+line)
				}
			}
		case *comment.Heading:
			result = append(result, prefix+"# "+docInlineText(b.Text))
		case *comment.List:
			result = append(result, renderDocList(b, prefix, bareMarker, column, tabWidth)...)
		}
	}

	for range trailingBlanks {
		result = append(result, bareMarker)
	}

	return result
}

// docInlineText extracts the text content from a slice of comment.Text nodes, preserving original
// whitespace and rendering doc links with their [bracket] syntax.
func docInlineText(texts []comment.Text) string {
	var b strings.Builder
	for _, t := range texts {
		switch t := t.(type) {
		case comment.Plain:
			b.WriteString(string(t))
		case comment.Italic:
			b.WriteString(string(t))
		case *comment.Link:
			b.WriteString(docInlineText(t.Text))
		case *comment.DocLink:
			b.WriteByte('[')
			b.WriteString(docInlineText(t.Text))
			b.WriteByte(']')
		}
	}
	return b.String()
}

// renderDocList renders a comment.List using appropriate bullet/number prefixes and wrapText.
func renderDocList(list *comment.List, prefix, bareMarker string, column, tabWidth int) []string {
	var result []string
	for i, item := range list.Items {
		if i > 0 && list.ForceBlankBetween {
			result = append(result, bareMarker)
		}
		var bullet string
		if item.Number != "" {
			bullet = item.Number + ". "
		} else {
			bullet = "- "
		}
		firstPrefix := prefix + "  " + bullet
		contPrefix := prefix + "  " + strings.Repeat(" ", len(bullet))

		for j, block := range item.Content {
			if j > 0 {
				result = append(result, bareMarker)
			}
			if para, ok := block.(*comment.Paragraph); ok {
				text := docInlineText(para.Text)
				if j == 0 {
					result = append(result, wrapText(text, firstPrefix, contPrefix, column, tabWidth)...)
				} else {
					result = append(result, wrapText(text, contPrefix, contPrefix, column, tabWidth)...)
				}
			}
		}
	}
	return result
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

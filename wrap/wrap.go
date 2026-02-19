package wrap

import (
	"strings"
	"unicode/utf8"
)

// wrapText wraps the given text to fit within columnWidth, accounting for the prefix added to each
// line. The first line uses prefix, subsequent lines use subsequentPrefix. Paragraph breaks (blank
// lines) are preserved.
func wrapText(text string, prefix string, subsequentPrefix string, columnWidth int, tabWidth int) []string {
	if text == "" {
		return []string{prefix}
	}

	paragraphs := splitParagraphs(text)
	var result []string
	for i, para := range paragraphs {
		if i > 0 {
			// Blank line between paragraphs, using the subsequent prefix trimmed of trailing space.
			result = append(result, strings.TrimRight(subsequentPrefix, " "))
		}
		lines := wrapParagraph(para, prefix, subsequentPrefix, columnWidth, tabWidth, i == 0)
		result = append(result, lines...)
	}
	return result
}

// splitParagraphs splits text into paragraphs separated by blank lines.
func splitParagraphs(text string) []string {
	lines := strings.Split(text, "\n")
	var paragraphs []string
	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(current) > 0 {
				paragraphs = append(paragraphs, strings.Join(current, " "))
				current = nil
			}
		} else {
			current = append(current, trimmed)
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, " "))
	}
	return paragraphs
}

// wrapParagraph wraps a single paragraph of text using greedy line breaking.
func wrapParagraph(text string, prefix, subsequentPrefix string, columnWidth, tabWidth int, isFirst bool) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currentPrefix := prefix
	if !isFirst {
		currentPrefix = subsequentPrefix
	}

	available := max(columnWidth-displayWidth(currentPrefix, tabWidth), 1)

	var line strings.Builder
	lineWidth := 0

	for _, word := range words {
		wordWidth := displayWidth(word, tabWidth)
		if line.Len() > 0 {
			// Check if adding this word (with space) would exceed the available width.
			if lineWidth+1+wordWidth > available {
				// Emit current line.
				lines = append(lines, currentPrefix+line.String())
				line.Reset()
				lineWidth = 0
				// Switch to subsequent prefix for remaining lines.
				currentPrefix = subsequentPrefix
				available = max(columnWidth-displayWidth(currentPrefix, tabWidth), 1)
			} else {
				line.WriteByte(' ')
				lineWidth++
			}
		}
		line.WriteString(word)
		lineWidth += wordWidth
	}
	if line.Len() > 0 {
		lines = append(lines, currentPrefix+line.String())
	}
	return lines
}

// displayWidth calculates the display width of a string, expanding tabs to tabWidth columns.
func displayWidth(s string, tabWidth int) int {
	col := 0
	for i := 0; i < len(s); {
		if s[i] == '\t' {
			col += tabWidth - (col % tabWidth)
			i++
		} else {
			_, size := utf8.DecodeRuneInString(s[i:])
			col++
			i += size
		}
	}
	return col
}

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
	// Split into tokens that preserve the original inter-word spacing. Each token has the
	// whitespace that preceded it (empty for the first token) and the word text.
	type token struct {
		gap  string // whitespace before this word in the original text
		word string
	}
	var tokens []token
	i := 0
	for i < len(text) {
		gapStart := i
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) {
			break
		}
		gap := text[gapStart:i]
		wordStart := i
		for i < len(text) && text[i] != ' ' && text[i] != '\t' {
			i++
		}
		tokens = append(tokens, token{gap: gap, word: text[wordStart:i]})
	}
	if len(tokens) == 0 {
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

	for idx, tok := range tokens {
		wordWidth := displayWidth(tok.word, tabWidth)
		if line.Len() > 0 {
			gapWidth := displayWidth(tok.gap, tabWidth)
			if idx == 0 {
				gapWidth = 0
			}
			// Use a single space as the minimum gap for wrapping decisions.
			breakWidth := max(gapWidth, 1)
			if lineWidth+breakWidth+wordWidth > available {
				lines = append(lines, currentPrefix+line.String())
				line.Reset()
				lineWidth = 0
				currentPrefix = subsequentPrefix
				available = max(columnWidth-displayWidth(currentPrefix, tabWidth), 1)
			} else {
				// Preserve original spacing within a line.
				if gapWidth > 0 {
					line.WriteString(tok.gap)
				} else {
					line.WriteByte(' ')
					gapWidth = 1
				}
				lineWidth += gapWidth
			}
		}
		line.WriteString(tok.word)
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

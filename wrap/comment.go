package wrap

import (
	"strings"
)

// segmentType identifies whether a segment is code or a comment.
type segmentType int

const (
	segmentCode    segmentType = iota
	segmentComment             // line comment block
	segmentBlock               // block comment (/* ... */)
)

// segment represents a contiguous block of either code or comments in source text.
type segment struct {
	typ    segmentType
	lines  []string
	indent string // leading whitespace of the comment block
	marker string // comment marker including trailing space, e.g., "// "
}

// parseSegments splits source lines into code and comment segments for the given language.
func parseSegments(lines []string, lang *Language) []segment {
	var segments []segment
	i := 0
	for i < len(lines) {
		// Try block comment first.
		if lang != nil && len(lang.BlockStart) > 0 {
			if seg, end := tryBlockComment(lines, i, lang); end > i {
				segments = append(segments, seg)
				i = end
				continue
			}
		}
		// Try line comment.
		if lang != nil && len(lang.LineMarkers) > 0 {
			if seg, end := tryLineCommentBlock(lines, i, lang); end > i {
				segments = append(segments, seg)
				i = end
				continue
			}
		}
		// Code line - accumulate consecutive code lines.
		start := i
		for i < len(lines) {
			if lang != nil {
				if _, end := tryLineCommentBlock(lines, i, lang); end > i {
					break
				}
				if len(lang.BlockStart) > 0 {
					if _, end := tryBlockComment(lines, i, lang); end > i {
						break
					}
				}
			}
			i++
		}
		segments = append(segments, segment{
			typ:   segmentCode,
			lines: lines[start:i],
		})
	}
	return segments
}

// tryLineCommentBlock tries to parse a block of consecutive line comments starting at line index i.
// Returns the segment and the index after the last comment line.
func tryLineCommentBlock(lines []string, i int, lang *Language) (segment, int) {
	indent, marker, ok := matchLineComment(lines[i], lang)
	if !ok {
		return segment{}, i
	}

	// Use the base marker token (without trailing space) for grouping, so that bare "//"-only lines
	// stay grouped with "// " content lines in the same comment block.
	baseMarker := strings.TrimRight(marker, " ")

	start := i
	for i < len(lines) {
		ind, mk, ok := matchLineComment(lines[i], lang)
		if !ok || ind != indent || strings.TrimRight(mk, " ") != baseMarker {
			break
		}
		// Prefer the longer marker (with space) for the segment, since that's the content marker.
		if len(mk) > len(marker) {
			marker = mk
		}
		i++
	}
	if i == start {
		return segment{}, start
	}
	return segment{
		typ:    segmentComment,
		lines:  lines[start:i],
		indent: indent,
		marker: marker,
	}, i
}

// matchLineComment checks if a line is a line comment and returns the indent and marker.
func matchLineComment(line string, lang *Language) (indent, marker string, ok bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return "", "", false
	}
	indent = line[:len(line)-len(trimmed)]
	for _, m := range lang.LineMarkers {
		if strings.HasPrefix(trimmed, m) {
			rest := trimmed[len(m):]
			// Check if the remaining text is a directive -- if so, treat the line as code.
			for _, d := range lang.Directives {
				if strings.HasPrefix(rest, d) {
					return "", "", false
				}
			}
			// The marker is the comment token plus one trailing space if present.
			if len(rest) > 0 && rest[0] == ' ' {
				marker = m + " "
			} else {
				marker = m
			}
			return indent, marker, true
		}
	}
	return "", "", false
}

// tryBlockComment tries to parse a block comment (/* ... */) starting at line index i.
func tryBlockComment(lines []string, i int, lang *Language) (segment, int) {
	trimmed := strings.TrimLeft(lines[i], " \t")
	indent := lines[i][:len(lines[i])-len(trimmed)]

	// Check if line starts with a block start marker.
	startMarker := ""
	for _, bs := range lang.BlockStart {
		if strings.HasPrefix(trimmed, bs) {
			startMarker = bs
			break
		}
	}
	if startMarker == "" {
		return segment{}, i
	}

	// Find the matching block end.
	endMarker := lang.BlockEnd[0] // use first block end marker
	start := i
	for i < len(lines) {
		if strings.Contains(lines[i], endMarker) {
			i++ // include the line with the end marker
			return segment{
				typ:    segmentBlock,
				lines:  lines[start:i],
				indent: indent,
			}, i
		}
		i++
	}
	// Unterminated block comment - treat as code.
	return segment{
		typ:   segmentCode,
		lines: lines[start:i],
	}, i
}

// isDecorationLine returns true if the comment content (after stripping the marker) consists
// entirely of repeated punctuation/symbols (e.g., "//========" or "//------").
func isDecorationLine(content string) bool {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) == 0 {
		return false
	}
	for _, r := range trimmed {
		switch r {
		case '=', '-', '*', '#', '~', '+', '_', '.':
			// decoration characters
		default:
			return false
		}
	}
	return true
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// sentenceCase uppercases first unicode letter and keeps the rest unchanged.
func sentenceCase(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	runes := []rune(trimmed)
	for index := range runes {
		if !unicode.IsLetter(runes[index]) {
			continue
		}

		runes[index] = unicode.ToUpper(runes[index])
		break
	}

	return string(runes)
}

// normalizeWrapWidth validates wrap width and falls back to default.
func normalizeWrapWidth(value int) int {
	if value <= 0 {
		return 80
	}

	return value
}

// normalizeTOCMode validates TOC mode and falls back to auto.
func normalizeTOCMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "always":
		return "always"
	case "off":
		return "off"
	default:
		return "auto"
	}
}

// wrapText wraps plain paragraphs at target width.
func wrapText(value string, width int) string {
	width = normalizeWrapWidth(width)

	normalized := strings.TrimSpace(normalizeLineEndings(value))
	if normalized == "" {
		return ""
	}

	paragraphs := strings.Split(normalized, "\n\n")
	out := make([]string, 0, len(paragraphs))

	for index := range paragraphs {
		line := strings.Join(strings.Fields(paragraphs[index]), " ")
		if line == "" {
			continue
		}

		out = append(out, wrapParagraph(line, width)...)
	}

	return strings.Join(out, "\n")
}

// wrapParagraph wraps one plain paragraph using rune-length width.
func wrapParagraph(value string, width int) []string {
	words := strings.Fields(value)
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, 2)
	current := ""
	currentLen := 0

	for index := range words {
		word := words[index]
		wordLen := utf8.RuneCountInString(word)
		if wordLen > width {
			if current != "" {
				lines = append(lines, current)
			}

			chunks := chunkByRuneWidth(word, width)
			lines = append(lines, chunks[:len(chunks)-1]...)
			current = chunks[len(chunks)-1]
			currentLen = utf8.RuneCountInString(current)
			continue
		}

		if current == "" {
			current = word
			currentLen = wordLen
			continue
		}

		if currentLen+1+wordLen <= width {
			current += " " + word
			currentLen += 1 + wordLen
			continue
		}

		lines = append(lines, current)
		current = word
		currentLen = wordLen
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// chunkByRuneWidth splits one long token into fixed-width rune chunks.
func chunkByRuneWidth(value string, width int) []string {
	if width <= 0 || utf8.RuneCountInString(value) <= width {
		return []string{value}
	}

	runes := []rune(value)
	out := make([]string, 0, (len(runes)/width)+1)
	for start := 0; start < len(runes); start += width {
		end := start + width
		end = min(end, len(runes))
		out = append(out, string(runes[start:end]))
	}

	return out
}

// normalizeLineEndings converts CRLF/CR to LF.
func normalizeLineEndings(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}

// compactMarkdownBlankLines collapses repeated blank lines for markdownlint.
func compactMarkdownBlankLines(value string) string {
	text := normalizeLineEndings(value)
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(text)
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import (
	"html"
	"strings"
)

// templateHTMLRichText renders limited markdown-like text to safe HTML.
//
// Supported syntax:
// * inline code with backticks: `code`
// * list lines prefixed by: "-", "*", or "1."
func templateHTMLRichText(value string) string {
	text := strings.TrimSpace(normalizeLineEndings(value))
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var out strings.Builder
	inList := false
	paragraphLines := make([]string, 0, 2)

	flushParagraph := func() {
		if len(paragraphLines) == 0 {
			return
		}

		content := strings.Join(paragraphLines, " ")
		content = strings.Join(strings.Fields(content), " ")
		if content != "" {
			out.WriteString("<p>")
			out.WriteString(renderInlineCodeHTML(content))
			out.WriteString("</p>")
		}

		paragraphLines = paragraphLines[:0]
	}

	for index := range lines {
		line := strings.TrimSpace(lines[index])
		if line == "" {
			flushParagraph()
			if inList {
				out.WriteString("</ul>")
				inList = false
			}

			continue
		}

		itemText, isListItem := parseSimpleListItem(line)
		if isListItem {
			flushParagraph()
			if !inList {
				out.WriteString("<ul>")
				inList = true
			}

			out.WriteString("<li>")
			out.WriteString(renderInlineCodeHTML(itemText))
			out.WriteString("</li>")
			continue
		}

		if inList {
			out.WriteString("</ul>")
			inList = false
		}

		paragraphLines = append(paragraphLines, line)
	}

	flushParagraph()
	if inList {
		out.WriteString("</ul>")
	}

	return out.String()
}

// parseSimpleListItem parses one simple markdown list marker line.
func parseSimpleListItem(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:]), true
	}

	digits := 0
	for digits < len(line) && line[digits] >= '0' && line[digits] <= '9' {
		digits++
	}

	if digits > 0 && digits+1 < len(line) && line[digits] == '.' && line[digits+1] == ' ' {
		return strings.TrimSpace(line[digits+2:]), true
	}

	return "", false
}

// renderInlineCodeHTML escapes text and renders paired backticks as <code>.
func renderInlineCodeHTML(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.Count(value, "`")%2 != 0 {
		return html.EscapeString(value)
	}

	parts := strings.Split(value, "`")
	if len(parts) == 1 {
		return html.EscapeString(value)
	}

	var out strings.Builder
	for index := range parts {
		if index%2 == 0 {
			out.WriteString(html.EscapeString(parts[index]))
			continue
		}

		out.WriteString("<code>")
		out.WriteString(html.EscapeString(parts[index]))
		out.WriteString("</code>")
	}

	return out.String()
}

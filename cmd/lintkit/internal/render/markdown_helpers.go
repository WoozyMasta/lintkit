// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"text/template"
	"unicode"

	"github.com/woozymasta/lintkit/lint"
)

// markdownTemplateFuncs returns helper functions used by markdown templates.
func markdownTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"anchor":             markdownHeadingAnchor,
		"blockquote":         markdownBlockQuote,
		"defaultEnabled":     markdownDefaultEnabled,
		"defaultEnabledText": markdownDefaultEnabledText,
		"descriptionText":    markdownDescriptionText,
		"messageText":        markdownMessageText,
		"escape":             markdownEscape,
		"fileKinds":          markdownFileKinds,
		"hasDefaultEnabled":  markdownHasDefaultEnabled,
		"headingAnchor":      markdownHeadingAnchor,
		"html":               templateHTMLEscape,
		"json":               templateJSON,
		"jsonPretty":         templatePrettyJSON,
		"md":                 markdownEscape,
		"richText":           templateHTMLRichText,
		"ruleHeading":        markdownRuleHeading,
		"sentenceCase":       sentenceCase,
		"tocList":            markdownTOCList,
		"wrap":               wrapText,
		"wrapText":           wrapText,
	}
}

// markdownBlockQuote formats plain text as markdown blockquote.
func markdownBlockQuote(value string) string {
	normalized := strings.TrimSpace(normalizeLineEndings(value))
	if normalized == "" {
		return ""
	}

	lines := strings.Split(normalized, "\n")
	for index := range lines {
		lines[index] = "> " + strings.TrimSpace(lines[index])
	}

	return strings.Join(lines, "\n")
}

// markdownMessageText returns one rule message for docs.
func markdownMessageText(rule lint.RuleSpec) string {
	return strings.TrimSpace(normalizeLineEndings(rule.Message))
}

// markdownDescriptionText returns one detailed rule description for docs.
func markdownDescriptionText(rule lint.RuleSpec) string {
	description := normalizeLineEndings(rule.Description)
	if strings.TrimSpace(description) == "" {
		return ""
	}

	return description
}

// markdownFileKinds joins file kinds for compact markdown output.
func markdownFileKinds(values []lint.FileKind) string {
	normalized := lint.NormalizeFileKinds(values)
	if len(normalized) == 0 {
		return ""
	}

	out := make([]string, 0, len(normalized))
	for index := range normalized {
		out = append(out, string(normalized[index]))
	}

	return strings.Join(out, ", ")
}

// markdownDefaultEnabled returns effective enabled value with default fallback.
func markdownDefaultEnabled(value *bool) bool {
	if value == nil {
		return true
	}

	return *value
}

// markdownHasDefaultEnabled reports whether explicit default enabled is set.
func markdownHasDefaultEnabled(value *bool) bool {
	return value != nil
}

// markdownDefaultEnabledText returns effective enabled text for markdown.
func markdownDefaultEnabledText(value *bool) string {
	if value == nil {
		return "true"
	}

	if *value {
		return "true"
	}

	return "false"
}

// markdownEscape escapes markdown-sensitive characters in inline text.
func markdownEscape(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"|", "\\|",
		"`", "\\`",
	)

	return replacer.Replace(value)
}

// markdownHeadingAnchor converts heading text to stable anchor token.
func markdownHeadingAnchor(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))
	lastDash := false

	for _, item := range trimmed {
		switch {
		case unicode.IsLetter(item), unicode.IsDigit(item), item == '_':
			builder.WriteRune(item)
			lastDash = false
		case unicode.IsSpace(item), item == '-':
			if lastDash || builder.Len() == 0 {
				continue
			}

			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

// markdownRuleHeading builds concise rule heading label.
func markdownRuleHeading(rule lint.RuleSpec) string {
	code := strings.TrimSpace(rule.Code)
	if code != "" {
		return code
	}

	id := strings.TrimSpace(rule.ID)
	if id != "" {
		return id
	}

	message := strings.TrimSpace(rule.Message)
	if message != "" {
		return message
	}

	return "rule"
}

// templateJSON marshals arbitrary value to compact JSON text.
func templateJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}

	return string(data)
}

// templatePrettyJSON marshals arbitrary value to indented JSON text.
func templatePrettyJSON(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(data)
}

// templateHTMLEscape escapes plain text for safe HTML rendering.
func templateHTMLEscape(value any) string {
	return html.EscapeString(fmt.Sprint(value))
}

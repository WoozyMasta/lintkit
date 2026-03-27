// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import (
	"strings"
	"testing"
)

func TestMarkdownTOCListCompact(t *testing.T) {
	t.Parallel()

	text := markdownTOCList([]Module{
		{
			ID:        "module_alpha",
			RuleCount: 3,
			Scopes: []Scope{
				{ID: "parse", RuleCount: 2},
				{ID: "lint", RuleCount: 1},
			},
		},
	})

	if strings.Contains(text, "\n\n") {
		t.Fatalf("markdownTOCList() contains blank lines:\n%s", text)
	}
	if !strings.Contains(text, "* [module_alpha](#module_alpha) (3)") {
		t.Fatalf("markdownTOCList() missing module row:\n%s", text)
	}
	if !strings.Contains(text, "  * [parse](#parse) (2)") {
		t.Fatalf("markdownTOCList() missing scope row:\n%s", text)
	}
}

func TestTemplateHTMLRichText(t *testing.T) {
	t.Parallel()

	rendered := templateHTMLRichText(
		"Line with `inline` code\n\n* first\n* second\n",
	)

	if !strings.Contains(rendered, "<p>Line with <code>inline</code> code</p>") {
		t.Fatalf("templateHTMLRichText() no inline code:\n%s", rendered)
	}
	if !strings.Contains(rendered, "<ul><li>first</li><li>second</li></ul>") {
		t.Fatalf("templateHTMLRichText() no list:\n%s", rendered)
	}
}

func TestWrapText(t *testing.T) {
	t.Parallel()

	text := wrapText("one two three four five", 7)
	lines := strings.Split(text, "\n")
	for index := range lines {
		if len([]rune(lines[index])) > 7 {
			t.Fatalf("wrapText() line too long: %q", lines[index])
		}
	}
}

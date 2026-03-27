// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"strings"
)

const (
	// minRuleIDDescriptionTokenLength is minimal meaningful description token.
	minRuleIDDescriptionTokenLength = 3
)

// BuildRuleID builds human-readable rule id from module, stage, and description.
func BuildRuleID(
	module string,
	stage Stage,
	description string,
	fallback Code,
) string {
	moduleToken := strings.TrimSpace(module)
	if moduleToken == "" {
		moduleToken = "module"
	}

	stageToken := ruleIDToken(string(stage))
	descriptionToken := ruleIDToken(description)
	if len(descriptionToken) < minRuleIDDescriptionTokenLength {
		descriptionToken = ""
	}

	if descriptionToken == "" {
		descriptionToken = ruleIDToken(FormatCode(fallback))
	}
	if descriptionToken == "" {
		descriptionToken = "rule"
	}

	if stageToken == "" {
		return moduleToken + "." + descriptionToken
	}

	return moduleToken + "." + stageToken + "." + descriptionToken
}

// ruleIDToken normalizes one rule id segment to stable lower-kebab form.
func ruleIDToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))
	lastDash := false
	for _, r := range trimmed {
		if isASCIIIDLetter(r) || isASCIIIDDigit(r) {
			builder.WriteRune(asciiLower(r))
			lastDash = false

			continue
		}

		if builder.Len() == 0 || lastDash {
			continue
		}

		builder.WriteByte('-')
		lastDash = true
	}

	out := strings.Trim(builder.String(), "-")
	if out == "" {
		return ""
	}

	first := rune(out[0])
	if !isASCIIIDLetter(first) {
		return "rule-" + out
	}

	return out
}

// isASCIIIDLetter reports whether rune is ASCII latin letter.
func isASCIIIDLetter(item rune) bool {
	return (item >= 'A' && item <= 'Z') || (item >= 'a' && item <= 'z')
}

// isASCIIIDDigit reports whether rune is ASCII digit.
func isASCIIIDDigit(item rune) bool {
	return item >= '0' && item <= '9'
}

// asciiLower converts ASCII uppercase letter to lowercase.
func asciiLower(item rune) rune {
	if item >= 'A' && item <= 'Z' {
		return item + ('a' - 'A')
	}

	return item
}

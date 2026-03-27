// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
	"strings"
)

// IsValidModuleToken reports whether one module token is stable.
func IsValidModuleToken(value string) bool {
	return isValidModuleID(strings.TrimSpace(value))
}

// IsValidScopeToken reports whether one scope token has stable shape.
func IsValidScopeToken(value string) bool {
	return isValidScopeToken(strings.TrimSpace(value))
}

// IsValidCodePrefixToken reports whether one code prefix token is stable.
func IsValidCodePrefixToken(value string) bool {
	return isValidCodePrefix(strings.TrimSpace(value))
}

// IsSupportedSeverity reports whether severity value is supported.
func IsSupportedSeverity(value Severity) bool {
	switch value {
	case SeverityError, SeverityWarning, SeverityInfo, SeverityNotice:
		return true
	default:
		return false
	}
}

// SeverityRank returns comparable severity rank where higher is more severe.
func SeverityRank(value Severity) int {
	switch value {
	case SeverityError:
		return 4
	case SeverityWarning:
		return 3
	case SeverityInfo:
		return 2
	case SeverityNotice:
		return 1
	default:
		return 0
	}
}

// ApplyCodePrefix applies exported prefix to code token.
func ApplyCodePrefix(prefix string, code Code) string {
	return applyCodePrefix(normalizeCodePrefix(prefix), code)
}

// RebaseCodePrefix rewrites one code token to target prefix.
//
// If source code already has "<prefix><number>" shape, the old prefix is replaced.
func RebaseCodePrefix(prefix string, code string) string {
	normalized := normalizeRuleCodeToken(code)
	if normalized == "" {
		return ""
	}

	targetPrefix := normalizeCodePrefix(prefix)
	if targetPrefix == "" {
		return normalized
	}

	digits, ok := splitPublicCodeToken(normalized)
	if !ok {
		return normalized
	}

	return targetPrefix + digits
}

// normalizeCodePrefix normalizes one exported code prefix token.
func normalizeCodePrefix(prefix string) string {
	value := strings.TrimSpace(prefix)
	return value
}

// ValidateCodePrefix validates one exported code prefix token.
func ValidateCodePrefix(prefix string) error {
	normalized := normalizeCodePrefix(prefix)
	if normalized == "" {
		return nil
	}

	if !isValidCodePrefix(normalized) {
		return fmt.Errorf("%w: %q", ErrInvalidCodePrefix, prefix)
	}

	return nil
}

// applyCodePrefix applies normalized prefix to normalized code.
func applyCodePrefix(prefix string, code Code) string {
	if code == 0 {
		return ""
	}

	codeText := FormatCode(code)
	if prefix == "" {
		return codeText
	}

	return prefix + codeText
}

// normalizeRuleCodeToken normalizes exported rule code token.
func normalizeRuleCodeToken(value string) string {
	return strings.TrimSpace(value)
}

// splitPublicCodeToken returns trailing numeric token from public code value.
func splitPublicCodeToken(value string) (string, bool) {
	normalized := normalizeRuleCodeToken(value)
	if normalized == "" {
		return "", false
	}

	digitStart := len(normalized)
	for digitStart > 0 {
		item := normalized[digitStart-1]
		if item < '0' || item > '9' {
			break
		}

		digitStart--
	}

	if digitStart == len(normalized) {
		return "", false
	}

	prefix := normalized[:digitStart]
	digits := normalized[digitStart:]
	if prefix == "" {
		return digits, true
	}

	if !isValidCodePrefix(prefix) {
		return "", false
	}

	return digits, true
}

// isValidModuleID reports whether one module token is stable.
func isValidModuleID(value string) bool {
	if value == "" {
		return false
	}

	first := value[0]
	if !isASCIILetter(first) {
		return false
	}

	for index := 1; index < len(value); index++ {
		item := value[index]
		if isASCIILetter(item) || (item >= '0' && item <= '9') || item == '_' || item == '-' {
			continue
		}

		return false
	}

	return true
}

// isValidCodePrefix reports whether code prefix token has stable shape.
func isValidCodePrefix(value string) bool {
	if value == "" {
		return false
	}

	first := value[0]
	if !isASCIILetter(first) {
		return false
	}

	for index := 1; index < len(value); index++ {
		item := value[index]
		if isASCIILetter(item) || (item >= '0' && item <= '9') || item == '_' || item == '-' {
			continue
		}

		return false
	}

	return true
}

// isASCIILetter reports whether one byte is ASCII latin letter.
func isASCIILetter(value byte) bool {
	return (value >= 'A' && value <= 'Z') || (value >= 'a' && value <= 'z')
}

// isValidScopeToken reports whether scope token has stable shape.
func isValidScopeToken(value string) bool {
	if value == "" {
		return false
	}

	first := value[0]
	if !isASCIILetter(first) {
		return false
	}

	for index := 1; index < len(value); index++ {
		item := value[index]
		if isASCIILetter(item) || (item >= '0' && item <= '9') || item == '_' || item == '-' {
			continue
		}

		return false
	}

	return true
}

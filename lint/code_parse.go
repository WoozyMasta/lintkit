// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"strconv"
	"strings"
)

// ParseCode parses base-10 numeric code token.
func ParseCode(value string) (Code, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, false
	}

	parsed, err := strconv.ParseUint(normalized, 10, 32)
	if err != nil {
		return 0, false
	}

	return Code(parsed), true
}

// ParsePublicCode parses exported code token "<PREFIX><NUMBER>" or "<NUMBER>".
func ParsePublicCode(value string) (Code, bool) {
	normalized := normalizeRuleCodeToken(value)
	if normalized == "" {
		return 0, false
	}

	digits, ok := splitPublicCodeToken(normalized)
	if !ok {
		return 0, false
	}

	return ParseCode(digits)
}

// FormatCode formats numeric code token as base-10 string.
func FormatCode(code Code) string {
	if code == 0 {
		return ""
	}

	return strconv.FormatUint(uint64(code), 10)
}

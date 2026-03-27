// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"slices"
	"strings"
)

const (
	// FileKindAny marks wildcard rule support for any target kind.
	FileKindAny FileKind = "*"
)

// FileKind defines normalized lint target/rule kind token.
type FileKind string

// NormalizeFileKind returns normalized file kind token.
func NormalizeFileKind(kind FileKind) FileKind {
	return FileKind(strings.ToLower(strings.TrimSpace(string(kind))))
}

// NormalizeFileKinds trims, deduplicates, and sorts file kind list.
func NormalizeFileKinds(kinds []FileKind) []FileKind {
	if len(kinds) == 0 {
		return nil
	}

	if len(kinds) == 1 {
		normalized := NormalizeFileKind(kinds[0])
		if normalized == "" {
			return nil
		}

		return []FileKind{normalized}
	}

	// For short lists avoid map allocation and deduplicate linearly.
	if len(kinds) <= 8 {
		out := make([]FileKind, 0, len(kinds))
		for index := range kinds {
			normalized := NormalizeFileKind(kinds[index])
			if normalized == "" {
				continue
			}

			duplicate := slices.Contains(out, normalized)

			if duplicate {
				continue
			}

			out = append(out, normalized)
		}

		if len(out) == 0 {
			return nil
		}

		slices.Sort(out)
		return out
	}

	seen := make(map[FileKind]struct{}, len(kinds))
	out := make([]FileKind, 0, len(kinds))
	for index := range kinds {
		normalized := NormalizeFileKind(kinds[index])
		if normalized == "" {
			continue
		}

		if _, exists := seen[normalized]; exists {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	slices.Sort(out)
	return out
}

// SupportsFileKind reports whether one rule supports current target kind.
func SupportsFileKind(supported []FileKind, target FileKind) bool {
	if len(supported) == 0 {
		return true
	}

	normalizedTarget := NormalizeFileKind(target)
	return SupportsNormalizedFileKind(supported, normalizedTarget)
}

// SupportsNormalizedFileKind reports whether one rule supports normalized kind.
//
// Use this helper in hot paths where target is already normalized and
// supported file kinds come from normalized rule specs.
func SupportsNormalizedFileKind(supported []FileKind, normalizedTarget FileKind) bool {
	if len(supported) == 0 {
		return true
	}

	if normalizedTarget == "" {
		return true
	}

	for index := range supported {
		supportedKind := supported[index]
		if supportedKind == "" {
			continue
		}

		if supportedKind == FileKindAny || supportedKind == normalizedTarget {
			return true
		}
	}

	return false
}

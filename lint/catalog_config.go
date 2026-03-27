// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
	"strings"
)

// validateCodeCatalogSpecs validates per-row code metadata against config.
func validateCodeCatalogSpecs(
	config CodeCatalogConfig,
	specs []CodeSpec,
) error {
	seenCodes := make(map[Code]int, len(specs))

	for index := range specs {
		if specs[index].Code == 0 {
			return fmt.Errorf(
				"%w: spec[%d] code is required",
				ErrInvalidCodeCatalogConfig,
				index,
			)
		}

		if previousIndex, exists := seenCodes[specs[index].Code]; exists {
			return fmt.Errorf(
				"%w: duplicate code %d in spec[%d] and spec[%d]",
				ErrInvalidCodeCatalogConfig,
				specs[index].Code,
				previousIndex,
				index,
			)
		}

		seenCodes[specs[index].Code] = index

		if strings.TrimSpace(specs[index].Message) == "" {
			return fmt.Errorf(
				"%w: spec[%d] message is required",
				ErrInvalidCodeCatalogConfig,
				index,
			)
		}

		scope := strings.TrimSpace(string(specs[index].Stage))
		if scope == "" {
			continue
		}

		if !isValidScopeToken(scope) {
			return fmt.Errorf(
				"%w: spec[%d] has invalid scope %q",
				ErrInvalidCodeCatalogConfig,
				index,
				scope,
			)
		}

		description, ok := config.ScopeDescriptions[Stage(scope)]
		if !ok || strings.TrimSpace(description) == "" {
			return fmt.Errorf(
				"%w: spec[%d] scope %q requires scope description",
				ErrInvalidCodeCatalogConfig,
				index,
				scope,
			)
		}
	}

	for scope := range config.ScopeDescriptions {
		normalized := strings.TrimSpace(string(scope))
		if normalized == "" {
			return fmt.Errorf(
				"%w: empty scope description key",
				ErrInvalidCodeCatalogConfig,
			)
		}

		if !isValidScopeToken(normalized) {
			return fmt.Errorf(
				"%w: invalid scope description key %q",
				ErrInvalidCodeCatalogConfig,
				normalized,
			)
		}
	}

	return nil
}

// normalizeCodeCatalogConfig validates and normalizes catalog config.
func normalizeCodeCatalogConfig(config CodeCatalogConfig) (CodeCatalogConfig, error) {
	out := config
	out.Module = strings.TrimSpace(out.Module)
	out.CodePrefix = normalizeCodePrefix(out.CodePrefix)
	out.ModuleName = strings.TrimSpace(out.ModuleName)
	out.ModuleDescription = strings.TrimSpace(out.ModuleDescription)
	out.ScopeDescriptions = normalizeScopeDescriptions(out.ScopeDescriptions)

	if out.Module == "" {
		return CodeCatalogConfig{}, fmt.Errorf(
			"%w: module is required",
			ErrInvalidCodeCatalogConfig,
		)
	}

	if !isValidModuleID(out.Module) {
		return CodeCatalogConfig{}, fmt.Errorf(
			"%w: invalid module %q",
			ErrInvalidCodeCatalogConfig,
			out.Module,
		)
	}

	if out.CodePrefix == "" {
		return CodeCatalogConfig{}, fmt.Errorf(
			"%w: code prefix is required",
			ErrInvalidCodeCatalogConfig,
		)
	}

	if err := ValidateCodePrefix(out.CodePrefix); err != nil {
		return CodeCatalogConfig{}, fmt.Errorf(
			"%w: %w",
			ErrInvalidCodeCatalogConfig,
			err,
		)
	}

	return out, nil
}

// normalizeScopeDescriptions validates and normalizes scope descriptions map.
func normalizeScopeDescriptions(
	value map[Stage]string,
) map[Stage]string {
	if len(value) == 0 {
		return nil
	}

	out := make(map[Stage]string, len(value))
	for key, description := range value {
		normalizedKey := Stage(strings.TrimSpace(string(key)))
		if normalizedKey == "" {
			continue
		}

		out[normalizedKey] = strings.TrimSpace(description)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// scopeDescription returns scope documentation text for stage token.
func (catalog CodeCatalog) scopeDescription(stage Stage) string {
	normalized := Stage(strings.TrimSpace(string(stage)))
	if normalized == "" || len(catalog.scopeDescriptions) == 0 {
		return ""
	}

	return strings.TrimSpace(catalog.scopeDescriptions[normalized])
}

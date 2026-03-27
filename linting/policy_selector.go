// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// selectorKindAll matches all registered rules.
	selectorKindAll selectorKind = iota + 1

	// selectorKindModule matches all rules in one module.
	selectorKindModule

	// selectorKindScope matches all rules in one module scope.
	selectorKindScope

	// selectorKindRule matches one exact rule id.
	selectorKindRule

	// selectorKindCode matches one short lint code across modules.
	selectorKindCode
)

// selectorKind stores parsed selector discriminator.
type selectorKind uint8

// parsedRuleSelector stores parsed selector fields.
type parsedRuleSelector struct {
	// raw stores normalized selector input.
	raw string

	// module stores parsed module token for module/rule selectors.
	module string

	// scope stores parsed scope token for scope selectors.
	scope string

	// ruleID stores parsed exact rule id for rule selectors.
	ruleID string

	// code stores parsed short lint code for code selectors.
	code lint.Code

	// kind stores parsed selector kind.
	kind selectorKind
}

// parseRuleSelector parses and validates selector into typed structure.
func parseRuleSelector(selector string) (parsedRuleSelector, error) {
	normalized := strings.TrimSpace(selector)
	if normalized == "" {
		return parsedRuleSelector{}, fmt.Errorf("%w: empty selector", ErrInvalidRunPolicy)
	}

	if normalized == RuleSelectorAll {
		return parsedRuleSelector{
			kind: selectorKindAll,
			raw:  normalized,
		}, nil
	}

	if module, ok := strings.CutSuffix(normalized, ".*"); ok {
		if strings.Count(module, ".") == 0 {
			if !isValidModule(module) {
				return parsedRuleSelector{}, fmt.Errorf(
					"%w: invalid selector %q",
					ErrInvalidRunPolicy,
					normalized,
				)
			}

			return parsedRuleSelector{
				kind:   selectorKindModule,
				module: module,
				raw:    normalized,
			}, nil
		}

		moduleToken, scopeToken, ok := strings.Cut(module, ".")
		if !ok || strings.Contains(scopeToken, ".") ||
			!isValidModule(moduleToken) || !isValidScope(scopeToken) {
			return parsedRuleSelector{}, fmt.Errorf(
				"%w: invalid selector %q",
				ErrInvalidRunPolicy,
				normalized,
			)
		}

		return parsedRuleSelector{
			kind:   selectorKindScope,
			module: moduleToken,
			scope:  scopeToken,
			raw:    normalized,
		}, nil
	}

	parsedCode, ok := lint.ParsePublicCode(normalized)
	if ok && parsedCode != 0 && isCodeSelectorToken(normalized) {
		return parsedRuleSelector{
			kind: selectorKindCode,
			code: parsedCode,
			raw:  strings.TrimSpace(normalized),
		}, nil
	}

	if !isValidRuleID(normalized) {
		return parsedRuleSelector{}, fmt.Errorf(
			"%w: invalid selector %q",
			ErrInvalidRunPolicy,
			normalized,
		)
	}

	dot := strings.IndexByte(normalized, '.')
	module := normalized
	if dot > 0 {
		module = normalized[:dot]
	}

	return parsedRuleSelector{
		kind:   selectorKindRule,
		module: module,
		ruleID: normalized,
		raw:    normalized,
	}, nil
}

// isCodeSelectorToken reports whether value is public code with non-empty prefix.
func isCodeSelectorToken(value string) bool {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return false
	}

	for _, item := range normalized {
		if (item >= 'a' && item <= 'z') || (item >= 'A' && item <= 'Z') {
			return true
		}
	}

	return false
}

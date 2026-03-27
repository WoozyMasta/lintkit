// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "strings"

const (
	// RunValueRuleOptionsKey stores current rule options payload in run values.
	RunValueRuleOptionsKey = "lint.rule.options"
)

// SetRunValue stores one typed run value under normalized key.
func SetRunValue[T any](run *RunContext, key string, value T) bool {
	if run == nil {
		return false
	}

	normalizedKey, ok := normalizeRunValueKey(key)
	if !ok {
		return false
	}

	if run.Values == nil {
		run.Values = make(map[string]any)
	}

	run.Values[normalizedKey] = value
	return true
}

// GetRunValue returns one typed run value by normalized key.
func GetRunValue[T any](run *RunContext, key string) (T, bool) {
	var zero T

	if run == nil || run.Values == nil {
		return zero, false
	}

	normalizedKey, ok := normalizeRunValueKey(key)
	if !ok {
		return zero, false
	}

	value, exists := run.Values[normalizedKey]
	if !exists {
		return zero, false
	}

	typed, ok := value.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// DeleteRunValue removes one run value by normalized key.
func DeleteRunValue(run *RunContext, key string) bool {
	if run == nil || run.Values == nil {
		return false
	}

	normalizedKey, ok := normalizeRunValueKey(key)
	if !ok {
		return false
	}

	_, exists := run.Values[normalizedKey]
	if !exists {
		return false
	}

	delete(run.Values, normalizedKey)
	return true
}

// SetCurrentRuleOptions stores current rule options payload in run context.
func SetCurrentRuleOptions(run *RunContext, options any) bool {
	if run == nil {
		return false
	}

	if run.Values == nil {
		run.Values = make(map[string]any)
	}

	run.Values[RunValueRuleOptionsKey] = options
	return true
}

// ClearCurrentRuleOptions removes current rule options payload from run context.
func ClearCurrentRuleOptions(run *RunContext) bool {
	if run == nil || run.Values == nil {
		return false
	}

	_, exists := run.Values[RunValueRuleOptionsKey]
	if !exists {
		return false
	}

	delete(run.Values, RunValueRuleOptionsKey)
	return true
}

// GetCurrentRuleOptions returns current rule options payload by target type.
func GetCurrentRuleOptions[T any](run *RunContext) (T, bool) {
	return GetRunValue[T](run, RunValueRuleOptionsKey)
}

// IndexByCode builds deterministic map index grouped by extracted code key.
func IndexByCode[T any, Code comparable](
	items []T,
	codeOf func(item T) Code,
) map[Code][]T {
	if len(items) == 0 || codeOf == nil {
		return nil
	}

	index := make(map[Code][]T)
	for itemIndex := range items {
		item := items[itemIndex]
		code := codeOf(item)
		index[code] = append(index[code], item)
	}

	return index
}

// SetIndexedByCode builds grouped index and stores it in run values.
func SetIndexedByCode[T any, Code comparable](
	run *RunContext,
	key string,
	items []T,
	codeOf func(item T) Code,
) bool {
	return SetRunValue(run, key, IndexByCode(items, codeOf))
}

// GetIndexedByCode returns grouped index map from run values.
func GetIndexedByCode[T any, Code comparable](
	run *RunContext,
	key string,
) (map[Code][]T, bool) {
	return GetRunValue[map[Code][]T](run, key)
}

// ValidateRunValueKey reports whether one run value key matches "module.key".
func ValidateRunValueKey(key string) bool {
	_, ok := normalizeRunValueKey(key)
	return ok
}

// normalizeRunValueKey returns normalized run value key in "module.key" shape.
func normalizeRunValueKey(key string) (string, bool) {
	normalized := strings.TrimSpace(key)
	if normalized == "" {
		return "", false
	}

	module, tail, ok := strings.Cut(normalized, ".")
	if !ok {
		return "", false
	}

	module = strings.TrimSpace(module)
	tail = strings.TrimSpace(tail)
	if module == "" || tail == "" {
		return "", false
	}

	if !isValidScopeToken(module) {
		return "", false
	}

	if !isValidRunValueTail(tail) {
		return "", false
	}

	if module == normalized[:len(module)] &&
		len(normalized) == len(module)+1+len(tail) &&
		normalized[len(module)] == '.' {
		return normalized, true
	}

	return module + "." + tail, true
}

// isValidRunValueTail reports whether tail tokens are valid ("a.b.c").
func isValidRunValueTail(tail string) bool {
	start := 0
	for index := 0; index <= len(tail); index++ {
		if index != len(tail) && tail[index] != '.' {
			continue
		}

		if index == start {
			return false
		}

		token := tail[start:index]
		if !isValidScopeToken(token) {
			return false
		}

		start = index + 1
	}

	return true
}

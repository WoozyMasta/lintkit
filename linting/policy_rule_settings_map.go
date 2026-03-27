// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import "fmt"

// NormalizeRuleSettingsMap validates and normalizes selector settings map.
//
// Returned map is detached from source and uses canonical selector keys:
//   - "*"
//   - "<module>.*"
//   - "<module>.<scope>.*"
//   - "<module>.<rule_name>"
//   - "<CODE>"
func NormalizeRuleSettingsMap(
	source map[string]RuleSettings,
) (map[string]RuleSettings, error) {
	if len(source) == 0 {
		return nil, nil
	}

	rawByCanonical := make(map[string]string, len(source))
	out := make(map[string]RuleSettings, len(source))

	for rawSelector, setting := range source {
		selector, err := parseRuleSelector(rawSelector)
		if err != nil {
			return nil, err
		}

		if setting.Severity != "" && !isSupportedSeverity(setting.Severity) {
			return nil, fmt.Errorf(
				"%w: selector %q has invalid severity %q",
				ErrInvalidRunPolicy,
				selector.raw,
				setting.Severity,
			)
		}

		if prevRaw, exists := rawByCanonical[selector.raw]; exists && prevRaw != rawSelector {
			return nil, fmt.Errorf(
				"%w: duplicate normalized selector %q (from %q and %q)",
				ErrInvalidRunPolicy,
				selector.raw,
				prevRaw,
				rawSelector,
			)
		}

		rawByCanonical[selector.raw] = rawSelector
		out[selector.raw] = cloneRuleSettings(setting)
	}

	return out, nil
}

// MergeRuleSettingsMaps merges selector settings with overlay precedence.
//
// Both maps are normalized before merge. Overlay entries replace base
// entries for the same normalized selector.
func MergeRuleSettingsMaps(
	base map[string]RuleSettings,
	overlay map[string]RuleSettings,
) (map[string]RuleSettings, error) {
	normalizedBase, err := NormalizeRuleSettingsMap(base)
	if err != nil {
		return nil, err
	}

	normalizedOverlay, err := NormalizeRuleSettingsMap(overlay)
	if err != nil {
		return nil, err
	}

	if len(normalizedBase) == 0 && len(normalizedOverlay) == 0 {
		return nil, nil
	}

	out := make(map[string]RuleSettings, len(normalizedBase)+len(normalizedOverlay))
	for selector, setting := range normalizedBase {
		out[selector] = cloneRuleSettings(setting)
	}

	for selector, setting := range normalizedOverlay {
		out[selector] = cloneRuleSettings(setting)
	}

	return out, nil
}

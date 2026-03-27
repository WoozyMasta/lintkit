// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"sort"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

// buildSelectorEnum builds ordered selector enum and reports whether enum is enabled.
func buildSelectorEnum(
	snapshot lint.RegistrySnapshot,
	rawModes []string,
) ([]string, bool) {
	modes := normalizeSelectorModes(rawModes)
	if modes["none"] {
		return nil, false
	}

	if modes["all"] || hasAllExplicitSelectorModes(modes) {
		return collectAllSelectorEnum(snapshot), true
	}

	return collectExplicitSelectorEnum(snapshot, modes), true
}

// normalizeSelectorModes normalizes repeated selector mode tokens.
func normalizeSelectorModes(rawModes []string) map[string]bool {
	if len(rawModes) == 0 {
		return map[string]bool{"all": true}
	}

	modes := make(map[string]bool, len(rawModes))
	for index := range rawModes {
		mode := strings.ToLower(strings.TrimSpace(rawModes[index]))
		if mode == "" {
			continue
		}

		modes[mode] = true
	}

	if len(modes) == 0 {
		modes["all"] = true
	}

	return modes
}

// hasAllExplicitSelectorModes reports whether module+id+code are all requested.
func hasAllExplicitSelectorModes(modes map[string]bool) bool {
	return modes["module"] && modes["id"] && modes["code"]
}

// collectAllSelectorEnum returns all known selectors including wildcard.
func collectAllSelectorEnum(snapshot lint.RegistrySnapshot) []string {
	values := make(map[string]struct{}, len(snapshot.Rules)*2+len(snapshot.Modules)+1)
	values["*"] = struct{}{}

	for index := range snapshot.Modules {
		module := strings.TrimSpace(snapshot.Modules[index].ID)
		if module == "" {
			continue
		}

		values[module+".*"] = struct{}{}
	}

	for index := range snapshot.Rules {
		rule := snapshot.Rules[index]
		id := strings.TrimSpace(rule.ID)
		if id != "" {
			values[id] = struct{}{}
		}

		module := strings.TrimSpace(rule.Module)
		if module != "" {
			values[module+".*"] = struct{}{}
		}

		scope := strings.TrimSpace(rule.Scope)
		if module != "" && scope != "" {
			values[module+"."+scope+".*"] = struct{}{}
		}

		code := strings.TrimSpace(rule.Code)
		if code != "" {
			values[code] = struct{}{}
		}
	}

	return sortedMapKeys(values)
}

// collectExplicitSelectorEnum returns selectors for requested kinds only.
func collectExplicitSelectorEnum(
	snapshot lint.RegistrySnapshot,
	modes map[string]bool,
) []string {
	values := make(map[string]struct{}, len(snapshot.Rules)*2+len(snapshot.Modules)+1)

	if hasAllExplicitSelectorModes(modes) {
		values["*"] = struct{}{}
	}

	if modes["module"] {
		for index := range snapshot.Modules {
			module := strings.TrimSpace(snapshot.Modules[index].ID)
			if module == "" {
				continue
			}

			values[module+".*"] = struct{}{}
		}

		for index := range snapshot.Rules {
			module := strings.TrimSpace(snapshot.Rules[index].Module)
			if module == "" {
				continue
			}

			values[module+".*"] = struct{}{}

			scope := strings.TrimSpace(snapshot.Rules[index].Scope)
			if scope != "" {
				values[module+"."+scope+".*"] = struct{}{}
			}
		}
	}

	if modes["id"] {
		for index := range snapshot.Rules {
			id := strings.TrimSpace(snapshot.Rules[index].ID)
			if id == "" {
				continue
			}

			values[id] = struct{}{}
		}
	}

	if modes["code"] {
		for index := range snapshot.Rules {
			code := strings.TrimSpace(snapshot.Rules[index].Code)
			if code == "" {
				continue
			}

			values[code] = struct{}{}
		}
	}

	return sortedMapKeys(values)
}

// sortedMapKeys returns sorted string keys from set-like map.
func sortedMapKeys(values map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

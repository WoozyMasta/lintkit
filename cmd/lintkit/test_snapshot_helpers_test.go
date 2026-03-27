// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"encoding/json"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

// testSnapshotJSON returns deterministic test registry snapshot JSON bytes.
func testSnapshotJSON(t *testing.T) []byte {
	t.Helper()

	disabled := false
	return testSnapshotJSONWithRules(t, []lint.RuleSpec{
		{
			ID:               "module_alpha.parse.trailing-comma",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "RVCFG2020",
			Message:          "trailing comma",
			Description:      "Warn when trailing comma is used.",
			DefaultSeverity:  lint.SeverityWarning,
			DefaultEnabled:   &disabled,
			FileKinds:        []lint.FileKind{"module_alpha.config"},
		},
	})
}

// testSnapshotJSONTwoModules returns deterministic two-module snapshot bytes.
func testSnapshotJSONTwoModules(t *testing.T) []byte {
	t.Helper()

	return testSnapshotJSONWithRules(t, []lint.RuleSpec{
		{
			ID:               "module_alpha.parse.trailing-comma",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "RVCFG2020",
			Message:          "trailing comma",
			Description:      "Warn when trailing comma is used.",
			DefaultSeverity:  lint.SeverityWarning,
		},
		{
			ID:               "module_beta.validate.rule",
			Module:           "module_beta",
			Scope:            "validate",
			ScopeDescription: "Validation diagnostics.",
			Code:             "RVMAT1001",
			Message:          "beta validation rule",
			Description:      "Validate module_beta payload.",
			DefaultSeverity:  lint.SeverityError,
		},
	})
}

// testSnapshotJSONWithRules encodes deterministic snapshot from provided rules.
func testSnapshotJSONWithRules(t *testing.T, rules []lint.RuleSpec) []byte {
	t.Helper()

	snapshot := lint.RegistrySnapshot{
		Rules: rules,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal test snapshot: %v", err)
	}

	return data
}

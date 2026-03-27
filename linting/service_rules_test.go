// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestAppendServiceRules(t *testing.T) {
	t.Parallel()

	input := lint.RegistrySnapshot{
		Modules: []lint.ModuleSpec{
			{ID: "module_alpha"},
		},
		Rules: []lint.RuleSpec{
			{
				ID:      "module_alpha.parse.rule-a",
				Module:  "module_alpha",
				Message: "rule a",
			},
		},
	}

	output := AppendServiceRules(input)
	if len(output.Modules) <= len(input.Modules) {
		t.Fatalf("len(output.Modules)=%d, want > %d", len(output.Modules), len(input.Modules))
	}

	if len(output.Rules) <= len(input.Rules) {
		t.Fatalf("len(output.Rules)=%d, want > %d", len(output.Rules), len(input.Rules))
	}
}

func TestAppendServiceRulesIdempotent(t *testing.T) {
	t.Parallel()

	first := AppendServiceRules(lint.RegistrySnapshot{})
	second := AppendServiceRules(first)

	if len(second.Modules) != len(first.Modules) {
		t.Fatalf("len(second.Modules)=%d, want %d", len(second.Modules), len(first.Modules))
	}

	if len(second.Rules) != len(first.Rules) {
		t.Fatalf("len(second.Rules)=%d, want %d", len(second.Rules), len(first.Rules))
	}
}

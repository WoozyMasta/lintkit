// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestCollectPolicyServiceDiagnosticsDeprecatedSelector(t *testing.T) {
	t.Parallel()

	registered := []lint.RuleSpec{
		{
			ID:               "module_alpha.parse.rule",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule",
			DefaultSeverity:  lint.SeverityWarning,
			Deprecated:       true,
		},
	}
	policy := &RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.parse.rule": {},
		},
	}

	diagnostics := collectPolicyServiceDiagnostics(policy, registered)
	if len(diagnostics) != 1 {
		t.Fatalf("len(diagnostics)=%d, want 1", len(diagnostics))
	}

	if diagnostics[0].RuleID != ruleIDPolicyDeprecatedSelector {
		t.Fatalf("RuleID=%q, want %q", diagnostics[0].RuleID, ruleIDPolicyDeprecatedSelector)
	}

	if diagnostics[0].Code != codePolicyDeprecatedSelector {
		t.Fatalf("Code=%q, want %q", diagnostics[0].Code, codePolicyDeprecatedSelector)
	}
}

func TestCollectPolicyServiceDiagnosticsSoftUnknownAndAmbiguous(t *testing.T) {
	t.Parallel()

	registered := []lint.RuleSpec{
		{
			ID:               "module_alpha.parse.rule_one",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule one",
			DefaultSeverity:  lint.SeverityWarning,
		},
		{
			ID:               "module_beta.parse.rule_two",
			Module:           "module_beta",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule two",
			DefaultSeverity:  lint.SeverityWarning,
		},
	}
	policy := &RunPolicy{
		Strict: false,
		Rules: map[string]RuleSettings{
			"UNKNOWN9999": {},
			"ALPHA2001":   {},
		},
	}

	diagnostics := collectPolicyServiceDiagnostics(policy, registered)
	if len(diagnostics) != 2 {
		t.Fatalf("len(diagnostics)=%d, want 2", len(diagnostics))
	}

	hasUnknown := false
	hasAmbiguous := false
	for index := range diagnostics {
		switch diagnostics[index].RuleID {
		case ruleIDPolicyUnknownSelector:
			hasUnknown = true
		case ruleIDPolicyAmbiguousSelector:
			hasAmbiguous = true
		}
	}

	if !hasUnknown || !hasAmbiguous {
		t.Fatalf("diagnostics mismatch, unknown=%v ambiguous=%v", hasUnknown, hasAmbiguous)
	}
}

func TestCollectPolicyServiceDiagnosticsShadowedAndNeverMatches(t *testing.T) {
	t.Parallel()

	policy := &RunPolicy{
		RuleEntries: []PolicyRuleEntry{
			{Selector: "*"},
			{Selector: "*"},
		},
		Overrides: []PolicyOverride{
			{
				Name:    "dead override",
				Matcher: PathMatcherFunc(nil),
				Rules: map[string]RuleSettings{
					"*": {},
				},
			},
		},
	}

	diagnostics := collectPolicyServiceDiagnostics(policy, nil)
	if len(diagnostics) != 2 {
		t.Fatalf("len(diagnostics)=%d, want 2", len(diagnostics))
	}

	hasShadowed := false
	hasNever := false
	for index := range diagnostics {
		if diagnostics[index].RuleID == ruleIDPolicyShadowedEntry {
			hasShadowed = true
		}

		if diagnostics[index].RuleID == ruleIDPolicyNeverMatchesOverride {
			hasNever = true
		}
	}

	if !hasShadowed || !hasNever {
		t.Fatalf("diagnostics mismatch, shadowed=%v never=%v", hasShadowed, hasNever)
	}
}

func TestAppendUniquePolicyDiagnostics(t *testing.T) {
	t.Parallel()

	run := lint.RunContext{
		Values: make(map[string]any),
	}
	items := []lint.Diagnostic{
		{
			RuleID:  ruleIDPolicyUnknownSelector,
			Message: "unknown policy selector",
		},
		{
			RuleID:  ruleIDPolicyUnknownSelector,
			Message: "unknown policy selector",
		},
	}

	target := make([]lint.Diagnostic, 0, 2)
	appendUniquePolicyDiagnostics(&run, &target, items)
	if len(target) != 1 {
		t.Fatalf("len(target)=%d, want 1", len(target))
	}

	appendUniquePolicyDiagnostics(&run, &target, items)
	if len(target) != 1 {
		t.Fatalf("len(target second append)=%d, want 1", len(target))
	}

	value, ok := run.Values[policyDiagnosticKey].(map[string]struct{})
	if !ok || len(value) != 1 {
		t.Fatalf("policy diagnostic state invalid: %#v", run.Values[policyDiagnosticKey])
	}

	if !strings.Contains(target[0].Message, "unknown") {
		t.Fatalf("unexpected message: %q", target[0].Message)
	}
}

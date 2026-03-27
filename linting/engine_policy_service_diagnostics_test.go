// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestEngineRunEmitsPolicyServiceDiagnostics(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	if err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:               "module_alpha.parse.rule",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule",
			DefaultSeverity:  lint.SeverityWarning,
			Deprecated:       true,
		},
		run: func(
			_ context.Context,
			_ *lint.RunContext,
			_ lint.DiagnosticEmit,
		) error {
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	runContext := lint.RunContext{
		Values: make(map[string]any),
	}
	policy := &RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.parse.rule": {},
		},
	}

	result, err := engine.RunDefault(runContext, &RunOptions{Policy: policy})
	if err != nil {
		t.Fatalf("RunDefault() error: %v", err)
	}

	found := false
	for index := range result.Diagnostics {
		if result.Diagnostics[index].RuleID == ruleIDPolicyDeprecatedSelector {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("policy deprecated service diagnostic was not emitted")
	}
}

func TestEngineRunEmitsCompiledPolicyServiceDiagnostics(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	if err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:               "module_alpha.parse.rule",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule",
			DefaultSeverity:  lint.SeverityWarning,
			Deprecated:       true,
		},
	}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	specs := engine.Rules()
	compiled, err := CompileRunPolicy(&RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.parse.rule": {},
		},
	}, specs)
	if err != nil {
		t.Fatalf("CompileRunPolicy() error: %v", err)
	}

	result, err := engine.RunDefault(
		lint.RunContext{Values: make(map[string]any)},
		&RunOptions{CompiledPolicy: compiled},
	)
	if err != nil {
		t.Fatalf("RunDefault() error: %v", err)
	}

	found := false
	for index := range result.Diagnostics {
		if result.Diagnostics[index].RuleID == ruleIDPolicyDeprecatedSelector {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("compiled policy deprecated service diagnostic was not emitted")
	}
}

func TestEngineRunCanDisablePolicyServiceDiagnostics(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	if err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:               "module_alpha.parse.rule",
			Module:           "module_alpha",
			Scope:            "parse",
			ScopeDescription: "Parser diagnostics.",
			Code:             "ALPHA2001",
			Message:          "rule",
			DefaultSeverity:  lint.SeverityWarning,
			Deprecated:       true,
		},
	}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	disabled := false
	result, err := engine.RunDefault(
		lint.RunContext{Values: make(map[string]any)},
		&RunOptions{
			EnableServiceDiagnostics: &disabled,
			Policy: &RunPolicy{
				Rules: map[string]RuleSettings{
					"module_alpha.parse.rule": {},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("RunDefault() error: %v", err)
	}

	for index := range result.Diagnostics {
		if result.Diagnostics[index].RuleID == ruleIDPolicyDeprecatedSelector {
			t.Fatal("policy deprecated service diagnostic must be disabled")
		}
	}
}

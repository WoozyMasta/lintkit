// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"errors"
	"github.com/woozymasta/lintkit/lint"
	"reflect"
	"strings"
	"testing"
)

func TestCompileRunPolicyNil(t *testing.T) {
	t.Parallel()

	compiled, err := CompileRunPolicy(nil, nil)
	if err != nil {
		t.Fatalf("CompileRunPolicy(nil) error: %v", err)
	}

	if compiled != nil {
		t.Fatalf("CompileRunPolicy(nil)=%v, want nil", compiled)
	}
}

func TestCompileRunPolicyValidation(t *testing.T) {
	t.Parallel()

	registered := []lint.RuleSpec{
		{
			ID:              "module_alpha.R001",
			Module:          "module_alpha",
			Scope:           "parse",
			Code:            "RVCFG2001",
			Message:         "rule",
			DefaultSeverity: lint.SeverityWarning,
		},
		{
			ID:              "module_beta.R001",
			Module:          "module_beta",
			Scope:           "lint",
			Code:            "RVMAT2001",
			Message:         "rule duplicate code",
			DefaultSeverity: lint.SeverityWarning,
		},
	}

	_, err := CompileRunPolicy(&RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.R001": {
				Severity: lint.Severity("fatal"),
			},
		},
	}, registered)
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("CompileRunPolicy(invalid severity) error=%v, want ErrInvalidRunPolicy", err)
	}

	_, err = CompileRunPolicy(&RunPolicy{
		Strict: true,
		Rules: map[string]RuleSettings{
			"module_alpha.UNKNOWN": {
				Enabled: BoolPtr(true),
			},
		},
	}, registered)
	if !errors.Is(err, ErrUnknownRuleSelector) {
		t.Fatalf("CompileRunPolicy(unknown strict selector) error=%v, want ErrUnknownRuleSelector", err)
	}

	_, err = CompileRunPolicy(&RunPolicy{
		Strict: true,
		Rules: map[string]RuleSettings{
			"RVCFG9999": {
				Enabled: BoolPtr(true),
			},
		},
	}, registered)
	if !errors.Is(err, ErrUnknownRuleSelector) {
		t.Fatalf("CompileRunPolicy(unknown code selector) error=%v, want ErrUnknownRuleSelector", err)
	}

	_, err = CompileRunPolicy(&RunPolicy{
		Strict: true,
		Rules: map[string]RuleSettings{
			"module_alpha.unknown.*": {
				Enabled: BoolPtr(true),
			},
		},
	}, registered)
	if !errors.Is(err, ErrUnknownRuleSelector) {
		t.Fatalf("CompileRunPolicy(unknown scope selector) error=%v, want ErrUnknownRuleSelector", err)
	}
}

func TestCompiledRunPolicyResolveParity(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_beta.missing_resource",
		Module:          "module_beta",
		Message:         "missing resource",
		DefaultSeverity: lint.SeverityWarning,
	}
	base := RunPolicy{
		Rules: map[string]RuleSettings{
			RuleSelectorAll: {
				Severity: lint.SeverityInfo,
			},
			"module_beta.*": {
				Severity: lint.SeverityError,
			},
			"module_beta.missing_resource": {
				Enabled: BoolPtr(false),
			},
		},
		Overrides: []PolicyOverride{
			{
				Matcher: PathMatcherFunc(func(path string, _ bool) bool {
					return strings.Contains(path, "/strict/")
				}),
				Rules: map[string]RuleSettings{
					"module_beta.missing_resource": {
						Enabled:  BoolPtr(true),
						Severity: lint.SeverityNotice,
					},
				},
			},
		},
	}

	compiled, err := base.Compile(nil)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}

	paths := []string{
		"workspace/base/material.module_beta",
		"workspace/strict/material.module_beta",
	}
	for index := range paths {
		got := compiled.Resolve(spec, paths[index], false)
		want := base.Resolve(spec, paths[index], false)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Resolve(%q)=%+v, want %+v", paths[index], got, want)
		}
	}
}

func TestCompileRunPolicySelectorNormalization(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_alpha.R001",
		Module:          "module_alpha",
		Scope:           "parse",
		Code:            "RVCFG2001",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}
	policy := RunPolicy{
		Strict: true,
		Rules: map[string]RuleSettings{
			" * ": {
				Severity: lint.SeverityInfo,
			},
			" module_alpha.* ": {
				Severity: lint.SeverityError,
			},
			" module_alpha.parse.* ": {
				Severity: lint.SeverityNotice,
			},
			" module_alpha.R001 ": {
				Enabled: BoolPtr(false),
			},
			" RVCFG2001 ": {
				Severity: lint.SeverityNotice,
			},
		},
	}

	compiled, err := policy.Compile([]lint.RuleSpec{spec})
	if err != nil {
		t.Fatalf("Compile(selector normalization) error: %v", err)
	}

	decision := compiled.Resolve(spec, "workspace/a/source.cfg", false)
	if decision.Enabled {
		t.Fatal("Resolve(selector normalization).Enabled=true, want false")
	}

	if decision.Severity != lint.SeverityNotice {
		t.Fatalf(
			"Resolve(selector normalization).Severity=%q, want %q",
			decision.Severity,
			lint.SeverityNotice,
		)
	}
}

func TestEngineRunCompiledPolicy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Message:         "macro redefinition",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{
				Message: "redefined macro",
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	policy := RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.R015": {
				Severity: lint.SeverityError,
			},
		},
	}
	compiled, err := policy.Compile(engine.Rules())
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}

	result, err := engine.Run(context.Background(), lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, &RunOptions{
		CompiledPolicy: compiled,
	})
	if err != nil {
		t.Fatalf("Run(compiled policy) error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].Severity != lint.SeverityError {
		t.Fatalf("Diagnostics[0].Severity=%q, want %q", result.Diagnostics[0].Severity, lint.SeverityError)
	}
}

func TestEngineRunRejectsPolicyAndCompiledPolicy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	_, err := engine.Run(context.Background(), lint.RunContext{}, &RunOptions{
		Policy:         &RunPolicy{},
		CompiledPolicy: &CompiledRunPolicy{},
	})
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("Run(policy+compiled) error=%v, want ErrInvalidRunPolicy", err)
	}
}

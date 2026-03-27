// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"github.com/woozymasta/lintkit/lint"
	"strings"
	"testing"
)

func TestRunPolicyResolvePrecedence(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_beta.missing_resource",
		Module:          "module_beta",
		Message:         "missing resource",
		DefaultSeverity: lint.SeverityWarning,
	}
	policy := RunPolicy{
		Rules: map[string]RuleSettings{
			RuleSelectorAll: {
				Severity: lint.SeverityInfo,
				Options: map[string]any{
					"source": "all",
				},
			},
			"module_beta.*": {
				Severity: lint.SeverityError,
				Options: map[string]any{
					"source": "module",
				},
			},
			"module_beta.missing_resource": {
				Enabled: BoolPtr(false),
				Options: map[string]any{
					"source": "rule",
				},
			},
		},
		Overrides: []PolicyOverride{
			{
				Name: "strict_path",
				Matcher: PathMatcherFunc(func(path string, _ bool) bool {
					return strings.Contains(path, "/strict/")
				}),
				Rules: map[string]RuleSettings{
					"module_beta.missing_resource": {
						Enabled:  BoolPtr(true),
						Severity: lint.SeverityNotice,
						Options: map[string]any{
							"source": "override",
						},
					},
				},
			},
		},
	}

	regular := policy.Resolve(spec, "workspace/base/material.module_beta", false)
	if regular.Enabled {
		t.Fatal("Resolve(regular).Enabled=true, want false")
	}

	if regular.Severity != lint.SeverityError {
		t.Fatalf("Resolve(regular).Severity=%q, want %q", regular.Severity, lint.SeverityError)
	}

	strict := policy.Resolve(spec, "workspace/strict/material.module_beta", false)
	if !strict.Enabled {
		t.Fatal("Resolve(strict).Enabled=false, want true")
	}

	if strict.Severity != lint.SeverityNotice {
		t.Fatalf("Resolve(strict).Severity=%q, want %q", strict.Severity, lint.SeverityNotice)
	}

	strictOptions, ok := strict.Options.(map[string]any)
	if !ok {
		t.Fatalf("Resolve(strict).Options type=%T, want map[string]any", strict.Options)
	}

	if strictOptions["source"] != "override" {
		t.Fatalf("Resolve(strict).Options[source]=%v, want override", strictOptions["source"])
	}
}

func TestRunPolicyPathEnabled(t *testing.T) {
	t.Parallel()

	policy := RunPolicy{
		Include: PathMatcherFunc(func(path string, _ bool) bool {
			return strings.HasSuffix(path, ".cfg")
		}),
		Exclude: PathMatcherFunc(func(path string, _ bool) bool {
			return strings.Contains(path, "/vendor/")
		}),
	}

	if !policy.PathEnabled("workspace/a/source.cfg", false) {
		t.Fatal("PathEnabled(source.cfg)=false, want true")
	}

	if policy.PathEnabled("workspace/a/source.bin", false) {
		t.Fatal("PathEnabled(source.bin)=true, want false")
	}

	if policy.PathEnabled("workspace/vendor/source.cfg", false) {
		t.Fatal("PathEnabled(vendor)=true, want false")
	}
}

func TestRunPolicyResolveDefaultEnabledFromRuleSpec(t *testing.T) {
	t.Parallel()

	disabled := false
	spec := lint.RuleSpec{
		ID:              "module_alpha.rule_disabled",
		Module:          "module_alpha",
		Message:         "disabled by default",
		DefaultSeverity: lint.SeverityWarning,
		DefaultEnabled:  &disabled,
	}

	decision := (*RunPolicy)(nil).Resolve(spec, "workspace/a/source.cfg", false)
	if decision.Enabled {
		t.Fatal("Resolve(nil policy).Enabled=true, want false")
	}

	enabledPolicy := RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.rule_disabled": {
				Enabled: BoolPtr(true),
			},
		},
	}
	decision = enabledPolicy.Resolve(spec, "workspace/a/source.cfg", false)
	if !decision.Enabled {
		t.Fatal("Resolve(policy override).Enabled=false, want true")
	}
}

func TestRunPolicyValidate(t *testing.T) {
	t.Parallel()

	registered := []lint.RuleSpec{
		{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Scope:           "parse",
			Code:            "RVCFG2015",
			Message:         "macro redefinition",
			DefaultSeverity: lint.SeverityWarning,
		},
		{
			ID:              "module_gamma.R015",
			Module:          "module_gamma",
			Scope:           "lint",
			Code:            "RVMAT2015",
			Message:         "same code in other module",
			DefaultSeverity: lint.SeverityWarning,
		},
	}

	tests := []struct {
		name    string
		policy  RunPolicy
		wantErr error
	}{
		{
			name: "invalid_severity",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"module_alpha.R015": {
						Severity: lint.Severity("fatal"),
					},
				},
			},
			wantErr: ErrInvalidRunPolicy,
		},
		{
			name: "unknown_rule_selector",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"module_alpha.UNKNOWN": {
						Enabled: BoolPtr(true),
					},
				},
			},
			wantErr: ErrUnknownRuleSelector,
		},
		{
			name: "unknown_module_selector",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"module_beta.*": {
						Enabled: BoolPtr(true),
					},
				},
			},
			wantErr: ErrUnknownRuleSelector,
		},
		{
			name: "unknown_scope_selector",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"module_alpha.unknown.*": {
						Enabled: BoolPtr(true),
					},
				},
			},
			wantErr: ErrUnknownRuleSelector,
		},
		{
			name: "invalid_module_code_selector",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"module_alpha:2015": {
						Enabled: BoolPtr(false),
					},
				},
			},
			wantErr: ErrInvalidRunPolicy,
		},
		{
			name: "unknown_code_selector",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					"RVCFG9999": {
						Enabled: BoolPtr(true),
					},
				},
			},
			wantErr: ErrUnknownRuleSelector,
		},
		{
			name: "valid_selectors",
			policy: RunPolicy{
				Rules: map[string]RuleSettings{
					RuleSelectorAll: {
						Severity: lint.SeverityInfo,
					},
					"module_alpha.*": {
						Enabled: BoolPtr(true),
					},
					"module_alpha.parse.*": {
						Severity: lint.SeverityNotice,
					},
					"module_alpha.R015": {
						Severity: lint.SeverityError,
					},
					"RVCFG2015": {
						Enabled: BoolPtr(true),
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.policy.Validate(registered)
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error=%v, want nil", err)
				}

				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error=%v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestPolicySetters(t *testing.T) {
	t.Parallel()

	policy := &RunPolicy{}
	if err := policy.SetAll(RuleSettings{
		Severity: lint.SeverityInfo,
	}); err != nil {
		t.Fatalf("SetAll() error: %v", err)
	}

	if err := policy.SetModule("module_alpha", RuleSettings{
		Enabled: BoolPtr(true),
	}); err != nil {
		t.Fatalf("SetModule() error: %v", err)
	}

	if err := policy.SetRule("module_alpha.R015", RuleSettings{
		Severity: lint.SeverityError,
	}); err != nil {
		t.Fatalf("SetRule() error: %v", err)
	}

	if err := policy.SetScope("module_alpha", "parse", RuleSettings{
		Severity: lint.SeverityNotice,
	}); err != nil {
		t.Fatalf("SetScope() error: %v", err)
	}

	if err := policy.SetCode("RVCFG2015", RuleSettings{
		Enabled: BoolPtr(false),
	}); err != nil {
		t.Fatalf("SetCode() error: %v", err)
	}

	override := PolicyOverride{
		Name: "config_only",
		Matcher: PathMatcherFunc(func(path string, _ bool) bool {
			return strings.HasSuffix(path, ".cfg")
		}),
	}
	if err := override.SetModule("module_alpha", RuleSettings{
		Enabled: BoolPtr(false),
	}); err != nil {
		t.Fatalf("override.SetModule() error: %v", err)
	}

	if err := override.SetScope("module_alpha", "parse", RuleSettings{
		Enabled: BoolPtr(true),
	}); err != nil {
		t.Fatalf("override.SetScope() error: %v", err)
	}

	if err := policy.AddOverride(override); err != nil {
		t.Fatalf("AddOverride() error: %v", err)
	}

	if len(policy.Rules) != 5 {
		t.Fatalf("len(policy.Rules)=%d, want 5", len(policy.Rules))
	}

	if len(policy.Overrides) != 1 {
		t.Fatalf("len(policy.Overrides)=%d, want 1", len(policy.Overrides))
	}
}

func TestPolicySettersValidation(t *testing.T) {
	t.Parallel()

	policy := &RunPolicy{}
	if err := policy.SetRule("1bad.id", RuleSettings{}); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("SetRule() error=%v, want ErrInvalidRunPolicy", err)
	}

	if err := policy.SetModule("1bad", RuleSettings{}); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("SetModule() error=%v, want ErrInvalidRunPolicy", err)
	}

	if err := policy.SetScope("module_alpha", "bad.scope", RuleSettings{}); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("SetScope() error=%v, want ErrInvalidRunPolicy", err)
	}

	if err := policy.SetCode("2015", RuleSettings{}); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("SetCode() error=%v, want ErrInvalidRunPolicy", err)
	}

	if err := policy.SetAll(RuleSettings{
		Severity: lint.Severity("fatal"),
	}); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("SetAll() error=%v, want ErrInvalidRunPolicy", err)
	}
}

func TestSelectorHelpers(t *testing.T) {
	t.Parallel()

	ruleSelector, err := RuleSelector("module_alpha.R015")
	if err != nil {
		t.Fatalf("RuleSelector() error: %v", err)
	}

	if ruleSelector != "module_alpha.R015" {
		t.Fatalf("RuleSelector()=%q, want %q", ruleSelector, "module_alpha.R015")
	}

	moduleSelector, err := ModuleSelector("module_alpha")
	if err != nil {
		t.Fatalf("ModuleSelector() error: %v", err)
	}

	if moduleSelector != "module_alpha.*" {
		t.Fatalf("ModuleSelector()=%q, want %q", moduleSelector, "module_alpha.*")
	}

	scopeSelector, err := ScopeSelector("module_alpha", "parse")
	if err != nil {
		t.Fatalf("ScopeSelector() error: %v", err)
	}

	if scopeSelector != "module_alpha.parse.*" {
		t.Fatalf("ScopeSelector()=%q, want %q", scopeSelector, "module_alpha.parse.*")
	}

	if _, err := RuleSelector("module_alpha"); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("RuleSelector(invalid) error=%v, want ErrInvalidRunPolicy", err)
	}

	if _, err := ModuleSelector("module_alpha.core"); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("ModuleSelector(invalid) error=%v, want ErrInvalidRunPolicy", err)
	}

	if _, err := ScopeSelector("module_alpha", "bad.scope"); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("ScopeSelector(invalid) error=%v, want ErrInvalidRunPolicy", err)
	}

	codeSelector, err := CodeSelector("RVCFG2015")
	if err != nil {
		t.Fatalf("CodeSelector() error: %v", err)
	}
	if codeSelector != "RVCFG2015" {
		t.Fatalf("CodeSelector()=%q, want %q", codeSelector, "RVCFG2015")
	}
}

func TestParseRuleSelector(t *testing.T) {
	t.Parallel()

	all, err := parseRuleSelector(" * ")
	if err != nil {
		t.Fatalf("parseRuleSelector(all) error: %v", err)
	}
	if all.kind != selectorKindAll {
		t.Fatalf("all.kind=%d, want %d", all.kind, selectorKindAll)
	}

	module, err := parseRuleSelector(" module_alpha.* ")
	if err != nil {
		t.Fatalf("parseRuleSelector(module) error: %v", err)
	}
	if module.kind != selectorKindModule || module.module != "module_alpha" {
		t.Fatalf("module=%+v", module)
	}

	scope, err := parseRuleSelector(" module_alpha.parse.* ")
	if err != nil {
		t.Fatalf("parseRuleSelector(scope) error: %v", err)
	}
	if scope.kind != selectorKindScope || scope.module != "module_alpha" || scope.scope != "parse" {
		t.Fatalf("scope=%+v", scope)
	}

	rule, err := parseRuleSelector(" module_alpha.R015 ")
	if err != nil {
		t.Fatalf("parseRuleSelector(rule) error: %v", err)
	}
	if rule.kind != selectorKindRule ||
		rule.module != "module_alpha" ||
		rule.ruleID != "module_alpha.R015" {
		t.Fatalf("rule=%+v", rule)
	}

	code, err := parseRuleSelector(" RVCFG2015 ")
	if err != nil {
		t.Fatalf("parseRuleSelector(code) error: %v", err)
	}
	if code.kind != selectorKindCode || code.code != 2015 {
		t.Fatalf("code=%+v", code)
	}

	if _, err := parseRuleSelector(" module_alpha:2015 "); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("parseRuleSelector(module code) error=%v, want ErrInvalidRunPolicy", err)
	}

	if _, err := parseRuleSelector("1bad"); !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("parseRuleSelector(invalid) error=%v, want ErrInvalidRunPolicy", err)
	}
}

func TestRunPolicyResolveCodeSelectors(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_alpha.R001",
		Module:          "module_alpha",
		Code:            "RVCFG2001",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}
	policy := RunPolicy{
		Rules: map[string]RuleSettings{
			"RVCFG2001": {
				Severity: lint.SeverityError,
			},
		},
	}

	decision := policy.Resolve(spec, "workspace/a/source.cfg", false)
	if !decision.Enabled {
		t.Fatal("Resolve(code selectors).Enabled=false, want true")
	}

	if decision.Severity != lint.SeverityError {
		t.Fatalf(
			"Resolve(code selectors).Severity=%q, want %q",
			decision.Severity,
			lint.SeverityError,
		)
	}
}

func TestRunPolicyResolveScopeSelectors(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_alpha.rule-id",
		Module:          "module_alpha",
		Scope:           "parse",
		Code:            "RVCFG2001",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}
	policy := RunPolicy{
		Rules: map[string]RuleSettings{
			"module_alpha.*": {
				Severity: lint.SeverityError,
			},
			"module_alpha.parse.*": {
				Severity: lint.SeverityNotice,
			},
		},
	}

	decision := policy.Resolve(spec, "workspace/a/source.cfg", false)
	if decision.Severity != lint.SeverityNotice {
		t.Fatalf(
			"Resolve(scope selectors).Severity=%q, want %q",
			decision.Severity,
			lint.SeverityNotice,
		)
	}
}

func TestRunPolicyResolveRuleEntries(t *testing.T) {
	t.Parallel()

	spec := lint.RuleSpec{
		ID:              "module_alpha.rule-id",
		Module:          "module_alpha",
		Scope:           "parse",
		Code:            "RVCFG2001",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}
	policy := RunPolicy{
		RuleEntries: []PolicyRuleEntry{
			{
				Selector: RuleSelectorAll,
				Settings: RuleSettings{
					Severity: lint.SeverityError,
				},
			},
			{
				Selector: "module_alpha.*",
				Settings: RuleSettings{
					Severity: lint.SeverityNotice,
				},
				Matcher: PathMatcherFunc(func(path string, _ bool) bool {
					return strings.Contains(path, "/strict/")
				}),
			},
		},
	}

	regular := policy.Resolve(spec, "workspace/base/source.cfg", false)
	if regular.Severity != lint.SeverityError {
		t.Fatalf("Resolve(regular).Severity=%q, want %q", regular.Severity, lint.SeverityError)
	}

	strict := policy.Resolve(spec, "workspace/strict/source.cfg", false)
	if strict.Severity != lint.SeverityNotice {
		t.Fatalf("Resolve(strict).Severity=%q, want %q", strict.Severity, lint.SeverityNotice)
	}
}

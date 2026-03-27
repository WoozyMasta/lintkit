// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestRunPolicyConfigBuildWithoutPatterns(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Rules: []RunPolicyRuleConfig{
			{
				Rule:     "module_alpha.R015",
				Enabled:  BoolPtr(false),
				Severity: lint.SeverityError,
			},
		},
	}

	policy, err := config.Build(nil)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if len(policy.RuleEntries) != 1 {
		t.Fatalf("len(Build().RuleEntries)=%d, want 1", len(policy.RuleEntries))
	}

	if !policy.Strict {
		t.Fatal("Build().Strict=false, want true")
	}
}

func TestRunPolicyConfigBuildNormalizesRules(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Rules: []RunPolicyRuleConfig{
			{Rule: " * ", Severity: lint.SeverityInfo},
			{Rule: " module_alpha.* ", Severity: lint.SeverityError},
			{Rule: " module_alpha.R015 ", Enabled: BoolPtr(false)},
		},
	}

	policy, err := config.Build(nil)
	if err != nil {
		t.Fatalf("Build(normalize rules) error: %v", err)
	}

	decision := policy.Resolve(lint.RuleSpec{
		ID:              "module_alpha.R015",
		Module:          "module_alpha",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}, "workspace/a/source.cfg", false)

	if decision.Enabled {
		t.Fatal("Resolve(normalized rules).Enabled=true, want false")
	}

	if decision.Severity != lint.SeverityError {
		t.Fatalf(
			"Resolve(normalized rules).Severity=%q, want %q",
			decision.Severity,
			lint.SeverityError,
		)
	}
}

func TestRunPolicyConfigBuildNeedsCompilerForPatterns(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Exclude: []string{"workspace/**"},
	}

	_, err := config.Build(nil)
	if !errors.Is(err, ErrNilPatternMatcherCompiler) {
		t.Fatalf("Build() error=%v, want ErrNilPatternMatcherCompiler", err)
	}
}

func TestRunPolicyConfigBuildRejectsInvalidRule(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Rules: []RunPolicyRuleConfig{
			{Rule: "   "},
		},
	}

	_, err := config.Build(nil)
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("Build(invalid rule) error=%v, want ErrInvalidRunPolicy", err)
	}
}

func TestRunPolicyConfigBuildCompilesMatchers(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Exclude: []string{"vendor/"},
		Rules: []RunPolicyRuleConfig{
			{
				Rule:    "module_alpha.R015",
				Exclude: []string{"generated/"},
				Enabled: BoolPtr(false),
			},
		},
	}

	compiler := func(patterns []string) (PathMatcher, error) {
		copied := append([]string(nil), patterns...)
		return PathMatcherFunc(func(path string, _ bool) bool {
			for index := range copied {
				if strings.Contains(path, copied[index]) {
					return true
				}
			}

			return false
		}), nil
	}

	policy, err := config.Build(compiler)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if policy.Exclude == nil {
		t.Fatal("Build().Exclude=nil, want matcher")
	}

	if len(policy.RuleEntries) != 1 {
		t.Fatalf("len(Build().RuleEntries)=%d, want 1", len(policy.RuleEntries))
	}

	if policy.RuleEntries[0].Matcher == nil {
		t.Fatal("Build().RuleEntries[0].Matcher=nil, want matcher")
	}

	if !policy.Exclude.Match("workspace/vendor/source.cfg", false) {
		t.Fatal("Build().Exclude matcher mismatch")
	}

	if policy.RuleEntries[0].Matcher.Match("workspace/generated/source.cfg", false) {
		t.Fatal("Build().RuleEntries[0].Matcher(generated)=true, want false")
	}
}

func TestRunPolicyConfigBuildSoftUnknownSelectors(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		SoftUnknownSelectors: true,
	}

	policy, err := config.Build(nil)
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if policy.Strict {
		t.Fatal("Build().Strict=true, want false")
	}
}

func TestRunPolicyConfigBuildRejectsInvalidFailOn(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		FailOn: lint.Severity("fatal"),
	}

	_, err := config.Build(nil)
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf("Build(invalid fail_on) error=%v, want ErrInvalidRunPolicy", err)
	}
}

func TestRunPolicyConfigShouldFail(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		FailOn: lint.SeverityWarning,
	}

	result := RunResult{
		Diagnostics: []lint.Diagnostic{
			{Severity: lint.SeverityInfo},
			{Severity: lint.SeverityWarning},
		},
	}

	fail, err := config.ShouldFail(result)
	if err != nil {
		t.Fatalf("ShouldFail() error: %v", err)
	}

	if !fail {
		t.Fatal("ShouldFail()=false, want true")
	}
}

func TestRunPolicyConfigShouldFailAlwaysIncludesRuleErrors(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{}
	result := RunResult{
		RuleErrors: []RuleError{{Cause: errors.New("boom")}},
	}

	fail, err := config.ShouldFail(result)
	if err != nil {
		t.Fatalf("ShouldFail(rule error) error: %v", err)
	}

	if !fail {
		t.Fatal("ShouldFail(rule error)=false, want true")
	}
}

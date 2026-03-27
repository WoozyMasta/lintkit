// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestMergeRunPolicyConfig(t *testing.T) {
	t.Parallel()

	base := RunPolicyConfig{
		Exclude: []string{"vendor/**", "vendor/**"},
		FailOn:  lint.SeverityError,
		Rules: []RunPolicyRuleConfig{
			{
				Rule:    " module_alpha.* ",
				Enabled: BoolPtr(true),
			},
		},
	}

	overlay := RunPolicyConfig{
		Exclude: []string{"generated/**"},
		FailOn:  lint.SeverityWarning,
		Rules: []RunPolicyRuleConfig{
			{
				Rule:     "module_alpha.R001",
				Severity: lint.SeverityError,
			},
		},
		SoftUnknownSelectors: true,
	}

	merged, err := MergeRunPolicyConfig(base, overlay)
	if err != nil {
		t.Fatalf("MergeRunPolicyConfig() error: %v", err)
	}

	if len(merged.Exclude) != 2 ||
		merged.Exclude[0] != "vendor/**" ||
		merged.Exclude[1] != "generated/**" {
		t.Fatalf("merged.Exclude=%v, want [vendor/** generated/**]", merged.Exclude)
	}

	if !merged.SoftUnknownSelectors {
		t.Fatal("merged.SoftUnknownSelectors=false, want true")
	}

	if merged.FailOn != lint.SeverityWarning {
		t.Fatalf("merged.FailOn=%q, want %q", merged.FailOn, lint.SeverityWarning)
	}

	if len(merged.Rules) != 2 {
		t.Fatalf("len(merged.Rules)=%d, want 2", len(merged.Rules))
	}

	if merged.Rules[0].Rule != "module_alpha.*" {
		t.Fatalf("merged.Rules[0].Rule=%q, want %q", merged.Rules[0].Rule, "module_alpha.*")
	}

	if merged.Rules[1].Rule != "module_alpha.R001" {
		t.Fatalf("merged.Rules[1].Rule=%q, want %q", merged.Rules[1].Rule, "module_alpha.R001")
	}
}

func TestMergeRunPolicyConfigRejectsInvalidRules(t *testing.T) {
	t.Parallel()

	_, err := MergeRunPolicyConfig(
		RunPolicyConfig{},
		RunPolicyConfig{
			Rules: []RunPolicyRuleConfig{
				{
					Rule:     "module_alpha.R001",
					Severity: lint.Severity("fatal"),
				},
			},
		},
	)
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf(
			"MergeRunPolicyConfig(invalid severity) error=%v, want ErrInvalidRunPolicy",
			err,
		)
	}
}

func TestMergeRunPolicyConfigRejectsInvalidFailOn(t *testing.T) {
	t.Parallel()

	_, err := MergeRunPolicyConfig(
		RunPolicyConfig{FailOn: lint.Severity("fatal")},
		RunPolicyConfig{},
	)
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf(
			"MergeRunPolicyConfig(invalid fail_on) error=%v, want ErrInvalidRunPolicy",
			err,
		)
	}
}

func TestRunPolicyConfigMergeMethod(t *testing.T) {
	t.Parallel()

	base := RunPolicyConfig{
		Rules: []RunPolicyRuleConfig{
			{
				Rule:     "module_alpha.R001",
				Severity: lint.SeverityWarning,
			},
		},
	}
	overlay := RunPolicyConfig{
		Rules: []RunPolicyRuleConfig{
			{
				Rule:     "module_alpha.R001",
				Severity: lint.SeverityError,
			},
		},
	}

	merged, err := base.Merge(overlay)
	if err != nil {
		t.Fatalf("RunPolicyConfig.Merge() error: %v", err)
	}

	if merged.Rules[1].Severity != lint.SeverityError {
		t.Fatalf("merged.Rules[1].Severity=%q, want %q", merged.Rules[1].Severity, lint.SeverityError)
	}
}

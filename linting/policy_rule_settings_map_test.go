// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"github.com/woozymasta/lintkit/lint"
	"testing"
)

func TestNormalizeRuleSettingsMap(t *testing.T) {
	t.Parallel()

	enabled := true
	source := map[string]RuleSettings{
		" * ": {
			Severity: lint.SeverityInfo,
		},
		" module_alpha.* ": {
			Enabled: &enabled,
		},
		" module_alpha.R001 ": {
			Severity: lint.SeverityError,
		},
		" RVCFG2001 ": {
			Enabled: BoolPtr(false),
		},
	}

	normalized, err := NormalizeRuleSettingsMap(source)
	if err != nil {
		t.Fatalf("NormalizeRuleSettingsMap() error: %v", err)
	}

	if len(normalized) != 4 {
		t.Fatalf("len(NormalizeRuleSettingsMap())=%d, want 4", len(normalized))
	}

	if _, exists := normalized[RuleSelectorAll]; !exists {
		t.Fatal("normalized map missing wildcard selector")
	}

	if _, exists := normalized["module_alpha.*"]; !exists {
		t.Fatal("normalized map missing module selector")
	}

	if _, exists := normalized["module_alpha.R001"]; !exists {
		t.Fatal("normalized map missing rule selector")
	}

	if _, exists := normalized["RVCFG2001"]; !exists {
		t.Fatal("normalized map missing code selector")
	}

	enabled = false
	if !*normalized["module_alpha.*"].Enabled {
		t.Fatal("NormalizeRuleSettingsMap() did not detach pointer values")
	}
}

func TestNormalizeRuleSettingsMapRejectsDuplicateNormalizedSelector(t *testing.T) {
	t.Parallel()

	_, err := NormalizeRuleSettingsMap(map[string]RuleSettings{
		"module_alpha.R001":   {},
		" module_alpha.R001 ": {},
	})
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf(
			"NormalizeRuleSettingsMap(duplicate) error=%v, want ErrInvalidRunPolicy",
			err,
		)
	}
}

func TestNormalizeRuleSettingsMapRejectsDuplicateNormalizedCodeSelector(t *testing.T) {
	t.Parallel()

	_, err := NormalizeRuleSettingsMap(map[string]RuleSettings{
		"RVCFG2001":   {},
		" RVCFG2001 ": {},
	})
	if !errors.Is(err, ErrInvalidRunPolicy) {
		t.Fatalf(
			"NormalizeRuleSettingsMap(duplicate code) error=%v, want ErrInvalidRunPolicy",
			err,
		)
	}
}

func TestMergeRuleSettingsMaps(t *testing.T) {
	t.Parallel()

	base := map[string]RuleSettings{
		RuleSelectorAll: {
			Severity: lint.SeverityWarning,
			Options: map[string]any{
				"profile": "base",
			},
		},
		"module_alpha.*": {
			Enabled: BoolPtr(true),
		},
	}
	overlay := map[string]RuleSettings{
		" module_alpha.* ": {
			Enabled: BoolPtr(false),
		},
		"module_alpha.R001": {
			Severity: lint.SeverityError,
			Options: map[string]any{
				"profile": "strict",
			},
		},
		"RVCFG2001": {
			Severity: lint.SeverityNotice,
		},
	}

	merged, err := MergeRuleSettingsMaps(base, overlay)
	if err != nil {
		t.Fatalf("MergeRuleSettingsMaps() error: %v", err)
	}

	if len(merged) != 4 {
		t.Fatalf("len(MergeRuleSettingsMaps())=%d, want 4", len(merged))
	}

	if *merged["module_alpha.*"].Enabled {
		t.Fatal("overlay did not override base selector settings")
	}

	if merged["module_alpha.R001"].Severity != lint.SeverityError {
		t.Fatalf(
			"merged rule severity=%q, want %q",
			merged["module_alpha.R001"].Severity,
			lint.SeverityError,
		)
	}

	options, ok := merged["module_alpha.R001"].Options.(map[string]any)
	if !ok {
		t.Fatalf(
			"merged rule options type=%T, want map[string]any",
			merged["module_alpha.R001"].Options,
		)
	}

	if options["profile"] != "strict" {
		t.Fatalf("merged rule option profile=%v, want strict", options["profile"])
	}
}

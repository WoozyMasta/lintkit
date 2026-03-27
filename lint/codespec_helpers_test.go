// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "testing"

func TestNewCodeSpec(t *testing.T) {
	t.Parallel()

	spec := NewCodeSpec(CodeSpec{
		Code:     1001,
		Stage:    "parse",
		Severity: SeverityWarning,
		Message:  "sample",
	})
	if spec.Code != 1001 {
		t.Fatalf("unexpected code: %d", spec.Code)
	}

	if spec.Stage != "parse" {
		t.Fatalf("unexpected stage: %q", spec.Stage)
	}

	if spec.Severity != SeverityWarning {
		t.Fatalf("unexpected severity: %q", spec.Severity)
	}

	if spec.Message != "sample" {
		t.Fatalf("unexpected message: %q", spec.Message)
	}
}

func TestNewCodeSpecWithOptions(t *testing.T) {
	t.Parallel()

	spec := NewCodeSpec(
		CodeSpec{
			Code:     1002,
			Stage:    "validate",
			Severity: SeverityError,
			Message:  "sample options",
			Rule: &CodeRuleOverride{
				FileKinds: []FileKind{"alpha.cfg", "alpha.cfg", "beta.cfg"},
			},
		},
	)
	spec = WithCodeEnabled(spec, false)
	spec = WithCodeSeverity(spec, SeverityWarning)
	spec = WithCodeOptions(spec, map[string]any{
		"allow_list": []string{"foo", "bar"},
	})
	spec = WithCodeRule(spec, CodeRuleOverride{})
	spec = WithCodeRule(spec, CodeRuleOverride{Deprecated: true})
	spec.Description = "Longer code description."
	if spec.Enabled == nil || *spec.Enabled {
		t.Fatal("unexpected enabled flag, want false")
	}

	options, ok := spec.Options.(map[string]any)
	if !ok {
		t.Fatalf("unexpected options type: %T", spec.Options)
	}

	if len(options) != 1 {
		t.Fatalf("unexpected options size: %d", len(options))
	}

	if spec.Description != "Longer code description." {
		t.Fatalf("unexpected description: %q", spec.Description)
	}

	if spec.Rule == nil {
		t.Fatalf("unexpected code rule override: %#v", spec.Rule)
	}

	if len(spec.Rule.FileKinds) != 3 {
		t.Fatalf("unexpected rule override file kinds size: %d", len(spec.Rule.FileKinds))
	}

	if !spec.Rule.Deprecated {
		t.Fatal("unexpected rule override deprecated flag, want true")
	}
}

func TestSeverityCodeSpecHelpers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		got      CodeSpec
		severity Severity
	}{
		{
			name:     "error",
			got:      ErrorCodeSpec(1010, "lex", "err"),
			severity: SeverityError,
		},
		{
			name:     "warning",
			got:      WarningCodeSpec(1011, "parse", "warn"),
			severity: SeverityWarning,
		},
		{
			name:     "info",
			got:      InfoCodeSpec(1012, "analyze", "info"),
			severity: SeverityInfo,
		},
		{
			name:     "notice",
			got:      NoticeCodeSpec(1013, "post", "note"),
			severity: SeverityNotice,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got.Severity != tc.severity {
				t.Fatalf(
					"unexpected severity: got %q want %q",
					tc.got.Severity,
					tc.severity,
				)
			}
		})
	}

	cases[0].got.Enabled = boolPtr(false)
	if cases[0].got.Enabled == nil || *cases[0].got.Enabled {
		t.Fatal("CodeSpec enabled setter fallback failed")
	}
}

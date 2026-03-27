// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "testing"

func TestSuppressionSetFunc(t *testing.T) {
	t.Parallel()

	var nilFn SuppressionSetFunc
	if nilFn.DecideSuppression("rule", "path", Position{}, Position{}).Suppressed {
		t.Fatal("nil SuppressionSetFunc.DecideSuppression().Suppressed=true, want false")
	}

	called := false
	fn := SuppressionSetFunc(func(
		ruleID string,
		path string,
		start Position,
		end Position,
	) bool {
		called = true
		return ruleID == "module_alpha.R001" &&
			path == "workspace/main/source.cfg" &&
			start.Line == 10 &&
			end.Line == 10
	})
	got := fn.DecideSuppression(
		"module_alpha.R001",
		"workspace/main/source.cfg",
		Position{Line: 10},
		Position{Line: 10},
	)
	if !called {
		t.Fatal("SuppressionSetFunc callback was not called")
	}

	if !got.Suppressed {
		t.Fatal("SuppressionSetFunc.DecideSuppression().Suppressed=false, want true")
	}

	if got.Scope != SuppressionScopeLine {
		t.Fatalf("Scope=%q, want %q", got.Scope, SuppressionScopeLine)
	}
}

func TestSuppressionDecisionFunc(t *testing.T) {
	t.Parallel()

	var nilFn SuppressionDecisionFunc
	decision := nilFn.DecideSuppression("rule", "path", Position{}, Position{})
	if decision.Suppressed {
		t.Fatal("nil SuppressionDecisionFunc.DecideSuppression().Suppressed=true, want false")
	}

	fn := SuppressionDecisionFunc(func(
		_ string,
		_ string,
		_ Position,
		_ Position,
	) SuppressionDecision {
		return SuppressionDecision{
			Suppressed: true,
			Reason:     "temporary ignore",
			Source:     "inline",
			ExpiresAt:  "2026-12-31T23:59:59Z",
		}
	})

	decision = fn.DecideSuppression("rule", "path", Position{}, Position{})
	if !decision.Suppressed {
		t.Fatal("SuppressionDecisionFunc.DecideSuppression().Suppressed=false, want true")
	}

	if decision.Reason != "temporary ignore" {
		t.Fatalf("Reason=%q, want temporary ignore", decision.Reason)
	}

	if decision.Scope != SuppressionScopeLine {
		t.Fatalf("Scope=%q, want %q", decision.Scope, SuppressionScopeLine)
	}
}

func TestNormalizeSuppressionScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input SuppressionScope
		want  SuppressionScope
	}{
		{
			name:  "line",
			input: SuppressionScopeLine,
			want:  SuppressionScopeLine,
		},
		{
			name:  "block",
			input: SuppressionScopeBlock,
			want:  SuppressionScopeBlock,
		},
		{
			name:  "file",
			input: SuppressionScopeFile,
			want:  SuppressionScopeFile,
		},
		{
			name:  "empty_default",
			input: "",
			want:  SuppressionScopeLine,
		},
		{
			name:  "trim_upper",
			input: "  FILE ",
			want:  SuppressionScopeFile,
		},
		{
			name:  "invalid_default",
			input: "region",
			want:  SuppressionScopeLine,
		},
	}

	for index := range testCases {
		testCase := testCases[index]
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeSuppressionScope(testCase.input)
			if got != testCase.want {
				t.Fatalf("NormalizeSuppressionScope(%q)=%q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

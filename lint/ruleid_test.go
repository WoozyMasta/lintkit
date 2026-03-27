// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "testing"

func TestBuildRuleID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		module      string
		stage       Stage
		description string
		fallback    Code
		want        string
	}{
		{
			name:        "stage and description",
			module:      "rvcfg",
			stage:       "parse",
			description: "expected value",
			fallback:    2016,
			want:        "rvcfg.parse.expected-value",
		},
		{
			name:        "fallback code",
			module:      "pofile",
			stage:       "lint",
			description: "",
			fallback:    2001,
			want:        "pofile.lint.rule-2001",
		},
		{
			name:        "description starts digit",
			module:      "module_alpha",
			stage:       "stage",
			description: "123 value",
			fallback:    1001,
			want:        "module_alpha.stage.rule-123-value",
		},
		{
			name:        "non ascii description fallback",
			module:      "module_alpha",
			stage:       "parse",
			description: "неожиданная ошибка",
			fallback:    2002,
			want:        "module_alpha.parse.rule-2002",
		},
		{
			name:        "short description fallback",
			module:      "module_alpha",
			stage:       "parse",
			description: "ab",
			fallback:    2003,
			want:        "module_alpha.parse.rule-2003",
		},
		{
			name:        "separator collapse",
			module:      "module_alpha",
			stage:       "parse",
			description: "a   ---   b",
			fallback:    0,
			want:        "module_alpha.parse.a-b",
		},
		{
			name:        "empty inputs",
			module:      "",
			stage:       "",
			description: "",
			fallback:    0,
			want:        "module.rule",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := BuildRuleID(tc.module, tc.stage, tc.description, tc.fallback)
			if got != tc.want {
				t.Fatalf("BuildRuleID()=%q, want %q", got, tc.want)
			}
		})
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import "testing"

func TestCompilePathRulesMatcher(t *testing.T) {
	t.Parallel()

	matcher, err := CompilePathRulesMatcher(
		[]string{
			"workspace/**",
			"examples/**",
		},
		PathRulesCompilerOptions{},
	)
	if err != nil {
		t.Fatalf("CompilePathRulesMatcher() error: %v", err)
	}

	if !matcher.Match("workspace/main/source.cfg", false) {
		t.Fatal("matcher.Match(workspace)=false, want true")
	}

	if !matcher.Match("examples/demo/init.txt", false) {
		t.Fatal("matcher.Match(examples)=false, want true")
	}

	if matcher.Match("tools/scripts/gen.go", false) {
		t.Fatal("matcher.Match(tools)=true, want false")
	}
}

func TestRunPolicyConfigBuildWithPathRulesCompiler(t *testing.T) {
	t.Parallel()

	config := RunPolicyConfig{
		Exclude: []string{"workspace/vendor/**"},
		Rules: []RunPolicyRuleConfig{
			{
				Rule:    "module_alpha.R015",
				Exclude: []string{"workspace/**/generated/**"},
				Enabled: BoolPtr(false),
			},
		},
	}

	policy, err := config.Build(PathRulesCompiler(PathRulesCompilerOptions{}))
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if !policy.PathEnabled("workspace/main/source.cfg", false) {
		t.Fatal("PathEnabled(main)=false, want true")
	}

	if policy.PathEnabled("workspace/vendor/source.cfg", false) {
		t.Fatal("PathEnabled(vendor)=true, want false")
	}

	if len(policy.RuleEntries) != 1 {
		t.Fatalf("len(RuleEntries)=%d, want 1", len(policy.RuleEntries))
	}

	if policy.RuleEntries[0].Matcher == nil {
		t.Fatal("RuleEntries[0].Matcher=nil, want matcher")
	}

	if policy.RuleEntries[0].Matcher.Match("workspace/a/generated/source.cfg", false) {
		t.Fatal("override exclude matcher mismatch")
	}
}

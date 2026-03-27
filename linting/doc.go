// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package linting provides upstream lint runtime primitives.

It validates rule metadata, applies selector-based policy,
executes rule runners, and returns aggregated results.

Quick start: create engine and register providers

	engine, err := linting.NewEngineWithProviders(
		modulea.LintRulesProvider{},
		moduleb.LintRulesProvider{},
	)
	if err != nil {
		return err
	}

Quick start: build run profile from config and run engine

	config := linting.RunPolicyConfig{
		Exclude: []string{"vendor/**"},
		FailOn:  lint.SeverityError,
		Rules: []linting.RunPolicyRuleConfig{
			{Rule: "*", Severity: lint.SeverityWarning},
			{Rule: "MODULEA1001", Enabled: linting.BoolPtr(false)},
			{
				Rule:    "moduleb.parse.*",
				Exclude: []string{"generated/**"},
				Severity: lint.SeverityNotice,
			},
		},
	}

	profile, err := linting.BuildRunProfile(linting.BuildRunProfileOptions{
		Base:     config,
		Compiler: linting.PathRulesCompiler(linting.PathRulesCompilerOptions{}),
		Registered: engine.Rules(),
	})
	if err != nil {
		return err
	}

	result, err := engine.RunWithProfile(
		ctx,
		lint.RunContext{TargetPath: path, TargetKind: "source"},
		profile,
	)
	if err != nil {
		return err
	}

	_ = result

	shouldFail, err := profile.ShouldFail(result)
	if err != nil {
		return err
	}

	_ = shouldFail
*/
package linting

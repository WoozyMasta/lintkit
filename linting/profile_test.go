// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestBuildPolicyOverlay(t *testing.T) {
	t.Parallel()

	failOn := lint.SeverityWarning
	softUnknownSelectors := true
	overlay := BuildPolicyOverlay(OverlayInput{
		DisableRules: []string{
			" ",
			"mod.*",
			"mod.*",
			" MOD2001 ",
		},
		Exclude: []string{
			"",
			"**/vendor/**",
			"**/vendor/**",
		},
		FailOn:               &failOn,
		SoftUnknownSelectors: &softUnknownSelectors,
	})

	if !overlay.SoftUnknownSelectors {
		t.Fatal("SoftUnknownSelectors=false, want true")
	}

	if overlay.FailOn != lint.SeverityWarning {
		t.Fatalf("FailOn=%q, want %q", overlay.FailOn, lint.SeverityWarning)
	}

	if len(overlay.Exclude) != 1 || overlay.Exclude[0] != "**/vendor/**" {
		t.Fatalf("Exclude=%v, want [**/vendor/**]", overlay.Exclude)
	}

	if len(overlay.Rules) != 2 {
		t.Fatalf("len(Rules)=%d, want 2", len(overlay.Rules))
	}

	if overlay.Rules[0].Rule != "mod.*" {
		t.Fatalf("Rules[0].Rule=%q, want mod.*", overlay.Rules[0].Rule)
	}

	if overlay.Rules[0].Enabled == nil || *overlay.Rules[0].Enabled {
		t.Fatalf("Rules[0].Enabled=%v, want false", overlay.Rules[0].Enabled)
	}

	if overlay.Rules[1].Rule != "MOD2001" {
		t.Fatalf("Rules[1].Rule=%q, want MOD2001", overlay.Rules[1].Rule)
	}
}

func TestBuildRunProfile(t *testing.T) {
	t.Parallel()

	failOn := lint.SeverityInfo
	profile, err := BuildRunProfile(BuildRunProfileOptions{
		Base: RunPolicyConfig{
			FailOn: lint.SeverityError,
			Rules: []RunPolicyRuleConfig{
				{
					Rule:     "mod.*",
					Severity: lint.SeverityWarning,
				},
			},
		},
		Overlays: []RunPolicyConfig{
			{
				Rules: []RunPolicyRuleConfig{
					{
						Rule:    "MOD2001",
						Enabled: BoolPtr(false),
					},
				},
			},
		},
		Compiler: PathRulesCompiler(PathRulesCompilerOptions{}),
		Registered: []lint.RuleSpec{
			{
				ID:              "mod.rule_a",
				Module:          "mod",
				Message:         "rule a",
				DefaultSeverity: lint.SeverityWarning,
				Code:            "MOD2001",
			},
		},
		FailOn: failOn,
	})
	if err != nil {
		t.Fatalf("BuildRunProfile() error: %v", err)
	}

	if !profile.Enabled {
		t.Fatal("Enabled=false, want true")
	}

	if !profile.EnableServiceDiagnostics {
		t.Fatal("EnableServiceDiagnostics=false, want true")
	}

	if profile.FailOn != lint.SeverityInfo {
		t.Fatalf("FailOn=%q, want %q", profile.FailOn, lint.SeverityInfo)
	}

	if profile.CompiledPolicy == nil {
		t.Fatal("CompiledPolicy=nil")
	}
}

func TestBuildRunProfileWithoutPolicySkipsCompile(t *testing.T) {
	t.Parallel()

	profile, err := BuildRunProfile(BuildRunProfileOptions{
		Base: RunPolicyConfig{},
	})
	if err != nil {
		t.Fatalf("BuildRunProfile() error: %v", err)
	}

	if profile.CompiledPolicy != nil {
		t.Fatalf("CompiledPolicy=%v, want nil", profile.CompiledPolicy)
	}
}

func TestBuildRunProfileInvalidFailOn(t *testing.T) {
	t.Parallel()

	profile, err := BuildRunProfile(BuildRunProfileOptions{
		FailOn: lint.Severity("fatal"),
	})
	if !errors.Is(err, ErrInvalidFailSeverity) {
		t.Fatalf("BuildRunProfile() error=%v, want ErrInvalidFailSeverity", err)
	}

	if profile != (RunProfile{}) {
		t.Fatalf("profile=%+v, want zero", profile)
	}
}

func TestRunProfileShouldFailDisabled(t *testing.T) {
	t.Parallel()

	profile := RunProfile{
		Enabled: false,
		FailOn:  lint.SeverityWarning,
	}

	fail, err := profile.ShouldFail(RunResult{
		Diagnostics: []lint.Diagnostic{
			{
				Severity: lint.SeverityError,
				Message:  "error",
			},
		},
	})
	if err != nil {
		t.Fatalf("ShouldFail() error: %v", err)
	}

	if fail {
		t.Fatal("ShouldFail()=true, want false")
	}
}

func TestRunProfileOptions(t *testing.T) {
	t.Parallel()

	compiled := &CompiledRunPolicy{}
	options, err := (RunProfile{
		Enabled:                  true,
		FailOn:                   lint.SeverityWarning,
		CompiledPolicy:           compiled,
		EnableServiceDiagnostics: false,
	}).Options()
	if err != nil {
		t.Fatalf("Options() error: %v", err)
	}

	if options.CompiledPolicy != compiled {
		t.Fatalf("CompiledPolicy=%p, want %p", options.CompiledPolicy, compiled)
	}

	if options.EnableServiceDiagnostics == nil {
		t.Fatal("EnableServiceDiagnostics=nil, want non-nil")
	}

	if *options.EnableServiceDiagnostics {
		t.Fatal("EnableServiceDiagnostics=true, want false")
	}
}

func TestRunProfileOptionsDefaultServiceDiagnostics(t *testing.T) {
	t.Parallel()

	options, err := (RunProfile{
		Enabled:                  true,
		FailOn:                   lint.SeverityWarning,
		EnableServiceDiagnostics: true,
	}).Options()
	if err != nil {
		t.Fatalf("Options() error: %v", err)
	}

	if options.EnableServiceDiagnostics != nil {
		t.Fatalf(
			"EnableServiceDiagnostics=%v, want nil (default true)",
			*options.EnableServiceDiagnostics,
		)
	}
}

func TestEngineRunWithProfile(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "mod.rule_a",
			Module:          "mod",
			Message:         "rule a",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(
			_ context.Context,
			_ *lint.RunContext,
			emit lint.DiagnosticEmit,
		) error {
			emit(lint.Diagnostic{Message: "warning"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunWithProfile(context.Background(), lint.RunContext{
		TargetPath: "source.cfg",
	}, RunProfile{
		Enabled:                  true,
		FailOn:                   lint.SeverityWarning,
		EnableServiceDiagnostics: true,
	})
	if err != nil {
		t.Fatalf("RunWithProfile() error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}
}

func TestEngineRunWithProfileDisabled(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "mod.rule_a",
			Module:          "mod",
			Message:         "rule a",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(
			_ context.Context,
			_ *lint.RunContext,
			emit lint.DiagnosticEmit,
		) error {
			emit(lint.Diagnostic{Message: "warning"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunWithProfile(context.Background(), lint.RunContext{
		TargetPath: "source.cfg",
	}, RunProfile{
		Enabled: false,
		FailOn:  lint.SeverityWarning,
	})
	if err != nil {
		t.Fatalf("RunWithProfile() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

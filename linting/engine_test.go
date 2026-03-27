// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"errors"
	"github.com/woozymasta/lintkit/lint"
	"strings"
	"testing"
)

// testRunner stores configurable test behavior for engine tests.
type testRunner struct {
	spec lint.RuleSpec
	run  func(ctx context.Context, run *lint.RunContext, emit lint.DiagnosticEmit) error
}

// RuleSpec returns test rule metadata.
func (runner testRunner) RuleSpec() lint.RuleSpec {
	return runner.spec
}

// Check runs test callback behavior.
func (runner testRunner) Check(ctx context.Context, run *lint.RunContext, emit lint.DiagnosticEmit) error {
	if runner.run == nil {
		return nil
	}

	return runner.run(ctx, run, emit)
}

func TestEngineRegisterAndRun(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_a",
				Module:          "mod",
				Message:         "rule a",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{
					Message: "warning a",
				})
				return nil
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_b",
				Module:          "mod",
				Message:         "rule b",
				DefaultSeverity: lint.SeverityError,
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{
					Message: "error b",
				})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "a/source.cfg",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 2 {
		t.Fatalf("len(Diagnostics)=%d, want 2", len(result.Diagnostics))
	}

	if result.Diagnostics[0].RuleID != "mod.rule_a" {
		t.Fatalf("Diagnostics[0].RuleID=%q", result.Diagnostics[0].RuleID)
	}

	if result.Diagnostics[0].Severity != lint.SeverityWarning {
		t.Fatalf("Diagnostics[0].Severity=%q", result.Diagnostics[0].Severity)
	}

	if result.Diagnostics[0].Path != "a/source.cfg" {
		t.Fatalf("Diagnostics[0].Path=%q", result.Diagnostics[0].Path)
	}
}

func TestEngineRunFilteredRuleIDs(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_a",
				Module:          "mod",
				Message:         "rule a",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{Message: "a"})
				return nil
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_b",
				Module:          "mod",
				Message:         "rule b",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{Message: "b"})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{}, &RunOptions{
		RuleIDs: []string{"mod.rule_b"},
	})
	if err != nil {
		t.Fatalf("Run(filtered) error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].RuleID != "mod.rule_b" {
		t.Fatalf("Diagnostics[0].RuleID=%q, want mod.rule_b", result.Diagnostics[0].RuleID)
	}
}

func TestEngineRegisterUpdatesDefaultRunOrder(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_beta.R002",
			Module:          "module_beta",
			Message:         "beta",
			DefaultSeverity: lint.SeverityWarning,
		},
	})
	if err != nil {
		t.Fatalf("Register(first) error: %v", err)
	}

	err = engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R001",
			Module:          "module_alpha",
			Message:         "alpha",
			DefaultSeverity: lint.SeverityWarning,
		},
	})
	if err != nil {
		t.Fatalf("Register(second) error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{}, nil)
	if err != nil {
		t.Fatalf("RunDefault() error: %v", err)
	}

	if len(result.RuleErrors) != 0 {
		t.Fatalf("len(RuleErrors)=%d, want 0", len(result.RuleErrors))
	}

	registered := engine.Rules()
	ids := make([]string, 0, len(registered))
	for index := range registered {
		ids = append(ids, registered[index].ID)
	}

	want := []string{"module_alpha.R001", "module_beta.R002"}
	if len(ids) != len(want) {
		t.Fatalf("len(ruleIDs)=%d, want %d", len(ids), len(want))
	}

	for index := range want {
		if ids[index] != want[index] {
			t.Fatalf("ruleIDs[%d]=%q, want %q", index, ids[index], want[index])
		}
	}
}

func TestEngineRuleReturnsDetachedSpec(t *testing.T) {
	t.Parallel()

	enabled := true
	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Message:         "rule",
			DefaultSeverity: lint.SeverityWarning,
			DefaultEnabled:  &enabled,
			DefaultOptions: map[string]any{
				"mode": "strict",
			},
			FileKinds: []lint.FileKind{"source.cfg"},
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	spec, ok := engine.Rule("module_alpha.parse.rule")
	if !ok {
		t.Fatal("Rule() returned not found")
	}

	options, ok := spec.DefaultOptions.(map[string]any)
	if !ok {
		t.Fatalf("DefaultOptions type=%T, want map[string]any", spec.DefaultOptions)
	}

	options["mode"] = "changed"
	spec.FileKinds[0] = "changed.kind"
	*spec.DefaultEnabled = false

	readAgain, ok := engine.Rule("module_alpha.parse.rule")
	if !ok {
		t.Fatal("Rule() returned not found after mutation")
	}

	if readAgain.FileKinds[0] != "source.cfg" {
		t.Fatalf("readAgain.FileKinds[0]=%q, want source.cfg", readAgain.FileKinds[0])
	}

	readAgainOptions, ok := readAgain.DefaultOptions.(map[string]any)
	if !ok {
		t.Fatalf(
			"readAgain.DefaultOptions type=%T, want map[string]any",
			readAgain.DefaultOptions,
		)
	}

	if readAgainOptions["mode"] != "strict" {
		t.Fatalf("readAgain.DefaultOptions[mode]=%v, want strict", readAgainOptions["mode"])
	}

	if readAgain.DefaultEnabled == nil || !*readAgain.DefaultEnabled {
		t.Fatal("readAgain.DefaultEnabled mutated via returned pointer")
	}
}

func TestEngineRunUnknownRuleID(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "mod.rule_a",
			Module:          "mod",
			Message:         "rule a",
			DefaultSeverity: lint.SeverityWarning,
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	_, err = engine.RunDefault(lint.RunContext{}, &RunOptions{
		RuleIDs: []string{"mod.rule_x"},
	})
	if !errors.Is(err, ErrUnknownRuleID) {
		t.Fatalf("Run() error=%v, want ErrUnknownRuleID", err)
	}
}

func TestEngineStopOnRuleError(t *testing.T) {
	t.Parallel()

	const runnerErrText = "runner failed"
	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_a",
				Module:          "mod",
				Message:         "rule a",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(_ context.Context, _ *lint.RunContext, _ lint.DiagnosticEmit) error {
				return errors.New(runnerErrText)
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_b",
				Module:          "mod",
				Message:         "rule b",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{Message: "should not run"})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{}, &RunOptions{
		StopOnRuleError: true,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.RuleErrors) != 1 {
		t.Fatalf("len(RuleErrors)=%d, want 1", len(result.RuleErrors))
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

func TestEngineRunConvertsRunnerPanicToRuleError(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_a_panic",
				Module:          "mod",
				Message:         "panic rule",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(
				_ context.Context,
				_ *lint.RunContext,
				_ lint.DiagnosticEmit,
			) error {
				panic("panic payload")
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_z_ok",
				Module:          "mod",
				Message:         "ok rule",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(
				_ context.Context,
				_ *lint.RunContext,
				emit lint.DiagnosticEmit,
			) error {
				emit(lint.Diagnostic{Message: "still running"})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{}, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.RuleErrors) != 1 {
		t.Fatalf("len(RuleErrors)=%d, want 1", len(result.RuleErrors))
	}

	if result.RuleErrors[0].RuleID != "mod.rule_a_panic" {
		t.Fatalf(
			"RuleErrors[0].RuleID=%q, want mod.rule_a_panic",
			result.RuleErrors[0].RuleID,
		)
	}

	if !errors.Is(result.RuleErrors[0].Cause, ErrRuleRunnerPanic) {
		t.Fatalf("RuleErrors[0].Cause=%v, want ErrRuleRunnerPanic", result.RuleErrors[0].Cause)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].RuleID != "mod.rule_z_ok" {
		t.Fatalf("Diagnostics[0].RuleID=%q, want mod.rule_z_ok", result.Diagnostics[0].RuleID)
	}
}

func TestEngineRunStopOnRuleErrorStopsAfterRunnerPanic(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_a_panic",
				Module:          "mod",
				Message:         "panic rule",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(
				_ context.Context,
				_ *lint.RunContext,
				_ lint.DiagnosticEmit,
			) error {
				panic("panic payload")
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "mod.rule_z_after",
				Module:          "mod",
				Message:         "after panic rule",
				DefaultSeverity: lint.SeverityWarning,
			},
			run: func(
				_ context.Context,
				_ *lint.RunContext,
				emit lint.DiagnosticEmit,
			) error {
				emit(lint.Diagnostic{Message: "must not run"})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{}, &RunOptions{
		StopOnRuleError: true,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.RuleErrors) != 1 {
		t.Fatalf("len(RuleErrors)=%d, want 1", len(result.RuleErrors))
	}

	if !errors.Is(result.RuleErrors[0].Cause, ErrRuleRunnerPanic) {
		t.Fatalf("RuleErrors[0].Cause=%v, want ErrRuleRunnerPanic", result.RuleErrors[0].Cause)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

func TestRunResultJoinErrors(t *testing.T) {
	t.Parallel()

	result := RunResult{
		RuleErrors: []RuleError{
			{RuleID: "a", Cause: errors.New("err a")},
			{RuleID: "b", Cause: errors.New("err b")},
		},
	}

	joined := result.JoinErrors()
	if joined == nil {
		t.Fatal("JoinErrors() returned nil on non-empty errors")
	}

	if !strings.Contains(joined.Error(), "a: err a") {
		t.Fatalf("JoinErrors() output=%q", joined.Error())
	}
}

func TestEngineRunFiltersByTargetKind(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(
		testRunner{
			spec: lint.RuleSpec{
				ID:              "module_alpha.R001",
				Module:          "module_alpha",
				Message:         "config rule",
				DefaultSeverity: lint.SeverityWarning,
				FileKinds:       []lint.FileKind{"source.cfg"},
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{Message: "config"})
				return nil
			},
		},
		testRunner{
			spec: lint.RuleSpec{
				ID:              "module_beta.R001",
				Module:          "module_beta",
				Message:         "material rule",
				DefaultSeverity: lint.SeverityWarning,
				FileKinds:       []lint.FileKind{"module_beta"},
			},
			run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
				emit(lint.Diagnostic{Message: "module_beta"})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/material.module_beta",
		TargetKind: "module_beta",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].RuleID != "module_beta.R001" {
		t.Fatalf("Diagnostics[0].RuleID=%q, want module_beta.R001", result.Diagnostics[0].RuleID)
	}
}

func TestEngineRunPolicyOverrideSeverity(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Message:         "macro redefinition",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{
				Message: "redefined macro",
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, &RunOptions{
		Policy: &RunPolicy{
			Rules: map[string]RuleSettings{
				"module_alpha.R015": {
					Severity: lint.SeverityError,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].Severity != lint.SeverityError {
		t.Fatalf("Diagnostics[0].Severity=%q, want %q", result.Diagnostics[0].Severity, lint.SeverityError)
	}
}

func TestEngineRunPolicyPathSkip(t *testing.T) {
	t.Parallel()

	executed := false
	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Message:         "macro redefinition",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			executed = true
			emit(lint.Diagnostic{
				Message: "redefined macro",
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, &RunOptions{
		Policy: &RunPolicy{
			Include: PathMatcherFunc(func(path string, _ bool) bool {
				return strings.Contains(path, "/must-run/")
			}),
		},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}

	if executed {
		t.Fatal("rule runner executed on excluded path")
	}
}

func TestEngineRunPolicyStrictValidation(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Message:         "macro redefinition",
			DefaultSeverity: lint.SeverityWarning,
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	_, err = engine.RunDefault(lint.RunContext{}, &RunOptions{
		Policy: &RunPolicy{
			Strict: true,
			Rules: map[string]RuleSettings{
				"module_alpha.UNKNOWN": {
					Enabled: BoolPtr(true),
				},
			},
		},
	})
	if !errors.Is(err, ErrUnknownRuleSelector) {
		t.Fatalf("Run() error=%v, want ErrUnknownRuleSelector", err)
	}
}

func TestEngineRunNilContext(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R001",
			Module:          "module_alpha",
			Message:         "simple rule",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{Message: "ok"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.Run(context.TODO(), lint.RunContext{}, nil)
	if err != nil {
		t.Fatalf("Run(nil, ...) error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}
}

func TestEngineRunRuleDefaultEnabledFalse(t *testing.T) {
	t.Parallel()

	disabled := false
	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R100",
			Module:          "module_alpha",
			Message:         "disabled by default rule",
			DefaultSeverity: lint.SeverityWarning,
			DefaultEnabled:  &disabled,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{Message: "should not run by default"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}
}

func TestEngineRunRuleOptionsFromPolicy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R101",
			Module:          "module_alpha",
			Message:         "policy options rule",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(
			_ context.Context,
			run *lint.RunContext,
			emit lint.DiagnosticEmit,
		) error {
			options, ok := lint.GetCurrentRuleOptions[map[string]any](run)
			if !ok {
				emit(lint.Diagnostic{Message: "missing options"})
				return nil
			}

			if options["max_len"] != 12 {
				emit(lint.Diagnostic{Message: "invalid options"})
				return nil
			}

			emit(lint.Diagnostic{Message: "options passed"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, &RunOptions{
		Policy: &RunPolicy{
			Rules: map[string]RuleSettings{
				"module_alpha.R101": {
					Options: map[string]any{
						"max_len": 12,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("len(Diagnostics)=%d, want 1", len(result.Diagnostics))
	}

	if result.Diagnostics[0].Message != "options passed" {
		t.Fatalf("Diagnostics[0].Message=%q", result.Diagnostics[0].Message)
	}
}

func TestEngineRunRuleOptionsAreIsolatedPerRun(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R102",
			Module:          "module_alpha",
			Message:         "policy options isolation",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(
			_ context.Context,
			run *lint.RunContext,
			emit lint.DiagnosticEmit,
		) error {
			options, ok := lint.GetCurrentRuleOptions[map[string]any](run)
			if !ok {
				emit(lint.Diagnostic{Message: "missing options"})
				return nil
			}

			if options["max_len"] != 12 {
				emit(lint.Diagnostic{Message: "mutated options"})
				return nil
			}

			options["max_len"] = 99
			emit(lint.Diagnostic{Message: "options passed"})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	options := map[string]any{
		"max_len": 12,
	}
	runOptions := &RunOptions{
		Policy: &RunPolicy{
			Rules: map[string]RuleSettings{
				"module_alpha.R102": {
					Options: options,
				},
			},
		},
	}

	first, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, runOptions)
	if err != nil {
		t.Fatalf("Run(first) error: %v", err)
	}

	second, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/a/source.cfg",
	}, runOptions)
	if err != nil {
		t.Fatalf("Run(second) error: %v", err)
	}

	if len(first.Diagnostics) != 1 || first.Diagnostics[0].Message != "options passed" {
		t.Fatalf("first diagnostics=%v, want one 'options passed'", first.Diagnostics)
	}

	if len(second.Diagnostics) != 1 || second.Diagnostics[0].Message != "options passed" {
		t.Fatalf("second diagnostics=%v, want one 'options passed'", second.Diagnostics)
	}

	if options["max_len"] != 12 {
		t.Fatalf("source options mutated to %v, want 12", options["max_len"])
	}
}

func TestEngineRunSuppressions(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R020",
			Module:          "module_alpha",
			Message:         "inline suppression sample",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{
				Message: "should be suppressed",
				Start: lint.Position{
					Line:   12,
					Column: 4,
				},
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	suppressed := false
	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/main/source.cfg",
	}, &RunOptions{
		Suppressions: lint.SuppressionDecisionFunc(
			func(
				ruleID string,
				path string,
				start lint.Position,
				end lint.Position,
			) lint.SuppressionDecision {
				if ruleID == "module_alpha.R020" &&
					path == "workspace/main/source.cfg" &&
					start.Line == 12 &&
					end.Line == 12 {
					suppressed = true
					return lint.SuppressionDecision{
						Suppressed: true,
						Scope:      lint.SuppressionScopeBlock,
						Reason:     "legacy block",
						Source:     "inline",
						ExpiresAt:  "2026-12-31T23:59:59Z",
					}
				}

				return lint.SuppressionDecision{}
			},
		),
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !suppressed {
		t.Fatal("suppression callback was not triggered with normalized diagnostic")
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}

	if len(result.Suppressed) != 1 {
		t.Fatalf("len(Suppressed)=%d, want 1", len(result.Suppressed))
	}

	if result.Suppressed[0].Reason != "legacy block" {
		t.Fatalf("Suppressed[0].Reason=%q, want legacy block", result.Suppressed[0].Reason)
	}

	if result.Suppressed[0].Scope != lint.SuppressionScopeBlock {
		t.Fatalf("Suppressed[0].Scope=%q, want %q", result.Suppressed[0].Scope, lint.SuppressionScopeBlock)
	}
}

func TestEngineRunSuppressionsDefaultScope(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.Register(testRunner{
		spec: lint.RuleSpec{
			ID:              "module_alpha.R021",
			Module:          "module_alpha",
			Message:         "suppression default scope",
			DefaultSeverity: lint.SeverityWarning,
		},
		run: func(_ context.Context, _ *lint.RunContext, emit lint.DiagnosticEmit) error {
			emit(lint.Diagnostic{
				Message: "suppressed with bool set func",
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	result, err := engine.RunDefault(lint.RunContext{
		TargetPath: "workspace/main/source.cfg",
	}, &RunOptions{
		Suppressions: lint.SuppressionSetFunc(
			func(
				_ string,
				_ string,
				_ lint.Position,
				_ lint.Position,
			) bool {
				return true
			},
		),
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Diagnostics) != 0 {
		t.Fatalf("len(Diagnostics)=%d, want 0", len(result.Diagnostics))
	}

	if len(result.Suppressed) != 1 {
		t.Fatalf("len(Suppressed)=%d, want 1", len(result.Suppressed))
	}

	if result.Suppressed[0].Scope != lint.SuppressionScopeLine {
		t.Fatalf("Suppressed[0].Scope=%q, want %q", result.Suppressed[0].Scope, lint.SuppressionScopeLine)
	}
}

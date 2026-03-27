// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// benchmarkDiagnosticMessage is synthetic benchmark diagnostic payload.
	benchmarkDiagnosticMessage = "synthetic diagnostic"
)

// benchmarkRunner emits configurable amount of diagnostics per run.
type benchmarkRunner struct {
	// spec stores stable rule metadata.
	spec lint.RuleSpec

	// diagnosticsPerRun controls emission count.
	diagnosticsPerRun int
}

// RuleSpec returns runner rule metadata.
func (runner benchmarkRunner) RuleSpec() lint.RuleSpec {
	return runner.spec
}

// Check emits synthetic diagnostics.
func (runner benchmarkRunner) Check(
	_ context.Context,
	_ *lint.RunContext,
	emit lint.DiagnosticEmit,
) error {

	for index := 0; index < runner.diagnosticsPerRun; index++ {
		emit(lint.Diagnostic{
			Message: benchmarkDiagnosticMessage,
		})
	}

	return nil
}

// BenchmarkRegistryRegisterMany measures registry registration throughput.
func BenchmarkRegistryRegisterMany(b *testing.B) {
	ruleCounts := []int{128, 512, 2048}

	for index := range ruleCounts {
		ruleCount := ruleCounts[index]
		specs := benchmarkRuleSpecs(ruleCount)

		b.Run(fmt.Sprintf("rules_%d", ruleCount), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for iteration := 0; iteration < b.N; iteration++ {
				registry := NewRegistry()
				if err := registry.RegisterMany(specs...); err != nil {
					b.Fatalf("RegisterMany() error: %v", err)
				}
			}
		})
	}
}

// BenchmarkRunPolicyCompile measures precompilation cost by ruleset size.
func BenchmarkRunPolicyCompile(b *testing.B) {
	ruleCounts := []int{128, 512, 2048}

	for index := range ruleCounts {
		ruleCount := ruleCounts[index]
		specs := benchmarkRuleSpecs(ruleCount)
		policy := benchmarkRunPolicy(specs)

		b.Run(fmt.Sprintf("rules_%d", ruleCount), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for iteration := 0; iteration < b.N; iteration++ {
				if _, err := policy.Compile(specs); err != nil {
					b.Fatalf("Compile() error: %v", err)
				}
			}
		})
	}
}

// BenchmarkEngineRun measures end-to-end engine runtime flows.
func BenchmarkEngineRun(b *testing.B) {
	testCases := []struct {
		// name is benchmark case label.
		name string

		// ruleCount controls engine ruleset size.
		ruleCount int

		// diagnosticsPerRule controls emitted diagnostics count.
		diagnosticsPerRule int
	}{
		{
			name:               "rules_64_emit_1",
			ruleCount:          64,
			diagnosticsPerRule: 1,
		},
		{
			name:               "rules_512_emit_1",
			ruleCount:          512,
			diagnosticsPerRule: 1,
		},
	}

	for index := range testCases {
		testCase := testCases[index]
		b.Run(testCase.name, func(b *testing.B) {
			benchmarkEngineRunModes(b, testCase.ruleCount, testCase.diagnosticsPerRule)
		})
	}
}

// BenchmarkEngineRunScale isolates rule-count scaling with fixed emit=1.
func BenchmarkEngineRunScale(b *testing.B) {
	ruleCounts := []int{16, 64, 256, 1024, 4096}

	for index := range ruleCounts {
		ruleCount := ruleCounts[index]
		b.Run(fmt.Sprintf("rules_%d_emit_1", ruleCount), func(b *testing.B) {
			benchmarkEngineRunModes(b, ruleCount, 1)
		})
	}
}

// BenchmarkEngineRunScaleNoPolicy isolates rule-count scaling without policy.
func BenchmarkEngineRunScaleNoPolicy(b *testing.B) {
	benchmarkEngineRunSingleModeScale(b, "no_policy")
}

// BenchmarkEngineRunScalePolicy isolates rule-count scaling with policy.
func BenchmarkEngineRunScalePolicy(b *testing.B) {
	benchmarkEngineRunSingleModeScale(b, "policy")
}

// BenchmarkEngineRunScaleCompiledPolicy isolates rule-count scaling with
// precompiled policy.
func BenchmarkEngineRunScaleCompiledPolicy(b *testing.B) {
	benchmarkEngineRunSingleModeScale(b, "compiled_policy")
}

// BenchmarkRegistryExport measures deterministic metadata export flows.
func BenchmarkRegistryExport(b *testing.B) {
	specs := benchmarkRuleSpecs(2048)
	registry := NewRegistry()
	if err := registry.RegisterMany(specs...); err != nil {
		b.Fatalf("RegisterMany() error: %v", err)
	}

	b.Run("json_pretty", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := registry.ExportJSON(true); err != nil {
				b.Fatalf("ExportJSON(true) error: %v", err)
			}
		}
	})
}

// benchmarkEngine builds benchmark-ready engine with synthetic rules.
func benchmarkEngine(
	b *testing.B,
	ruleCount int,
	diagnosticsPerRule int,
) *Engine {
	b.Helper()

	specs := benchmarkRuleSpecs(ruleCount)
	runners := benchmarkRunners(specs, diagnosticsPerRule)

	engine := NewEngine()
	if err := engine.Register(runners...); err != nil {
		b.Fatalf("Register() error: %v", err)
	}

	return engine
}

// benchmarkRuleSpecs generates deterministic synthetic rule specs.
func benchmarkRuleSpecs(ruleCount int) []lint.RuleSpec {
	specs := make([]lint.RuleSpec, 0, ruleCount)
	const moduleCount = 16

	for index := 0; index < ruleCount; index++ {
		module := fmt.Sprintf("module_%02d", index%moduleCount)
		ruleID := fmt.Sprintf("%s.R%04d", module, index+1)
		specs = append(specs, lint.RuleSpec{
			ID:              ruleID,
			Module:          module,
			Message:         "synthetic rule",
			Description:     "synthetic benchmark rule metadata",
			DefaultSeverity: lint.SeverityWarning,
			FileKinds:       []lint.FileKind{"source.cfg"},
		})
	}

	return specs
}

// benchmarkRunners converts specs to runtime rule runner instances.
func benchmarkRunners(
	specs []lint.RuleSpec,
	diagnosticsPerRule int,
) []lint.RuleRunner {
	runners := make([]lint.RuleRunner, 0, len(specs))

	for index := range specs {
		runners = append(runners, benchmarkRunner{
			spec:              specs[index],
			diagnosticsPerRun: diagnosticsPerRule,
		})
	}

	return runners
}

// benchmarkRunPolicy builds deterministic synthetic policy.
func benchmarkRunPolicy(specs []lint.RuleSpec) RunPolicy {
	exactSelector := specs[len(specs)-1].ID
	moduleSelector := specs[0].Module + ".*"

	return RunPolicy{
		Rules: map[string]RuleSettings{
			RuleSelectorAll: {
				Severity: lint.SeverityInfo,
			},
			moduleSelector: {
				Severity: lint.SeverityError,
			},
			exactSelector: {
				Enabled: BoolPtr(false),
			},
		},
		Overrides: []PolicyOverride{
			{
				Name: "vendor_disable",
				Matcher: PathMatcherFunc(func(path string, _ bool) bool {
					return strings.Contains(path, "/vendor/")
				}),
				Rules: map[string]RuleSettings{
					RuleSelectorAll: {
						Enabled: BoolPtr(false),
					},
				},
			},
			{
				Name: "strict_upgrade",
				Matcher: PathMatcherFunc(func(path string, _ bool) bool {
					return strings.Contains(path, "/strict/")
				}),
				Rules: map[string]RuleSettings{
					RuleSelectorAll: {
						Enabled:  BoolPtr(true),
						Severity: lint.SeverityError,
					},
				},
			},
		},
	}
}

// benchmarkEngineRunModes runs no_policy/policy/compiled_policy modes.
func benchmarkEngineRunModes(
	b *testing.B,
	ruleCount int,
	diagnosticsPerRule int,
) {
	b.Helper()

	engine := benchmarkEngine(b, ruleCount, diagnosticsPerRule)
	specs := engine.Rules()
	policy := benchmarkRunPolicy(specs)
	compiled, err := policy.Compile(specs)
	if err != nil {
		b.Fatalf("Compile() error: %v", err)
	}

	runContext := lint.RunContext{
		TargetPath: "workspace/main/source.cfg",
		TargetKind: "source.cfg",
		Values:     make(map[string]any),
	}

	policyOptions := &RunOptions{
		Policy: &policy,
	}
	compiledOptions := &RunOptions{
		CompiledPolicy: compiled,
	}
	noPolicyProfile := RunProfile{
		Enabled:                  true,
		FailOn:                   lint.SeverityError,
		EnableServiceDiagnostics: true,
	}
	compiledPolicyProfile := RunProfile{
		Enabled:                  true,
		FailOn:                   lint.SeverityError,
		CompiledPolicy:           compiled,
		EnableServiceDiagnostics: true,
	}

	b.Run("no_policy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := engine.RunDefault(runContext, nil); err != nil {
				b.Fatalf("RunDefault(no policy) error: %v", err)
			}
		}
	})

	b.Run("policy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := engine.RunDefault(runContext, policyOptions); err != nil {
				b.Fatalf("RunDefault(policy) error: %v", err)
			}
		}
	})

	b.Run("compiled_policy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := engine.RunDefault(runContext, compiledOptions); err != nil {
				b.Fatalf("RunDefault(compiled policy) error: %v", err)
			}
		}
	})

	b.Run("profile_no_policy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := engine.RunWithProfile(
				context.Background(),
				runContext,
				noPolicyProfile,
			); err != nil {
				b.Fatalf("RunWithProfile(no policy) error: %v", err)
			}
		}
	})

	b.Run("profile_compiled_policy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			if _, err := engine.RunWithProfile(
				context.Background(),
				runContext,
				compiledPolicyProfile,
			); err != nil {
				b.Fatalf("RunWithProfile(compiled policy) error: %v", err)
			}
		}
	})
}

// benchmarkEngineRunSingleModeScale runs scale benchmark for one run mode.
func benchmarkEngineRunSingleModeScale(b *testing.B, mode string) {
	b.Helper()

	ruleCounts := []int{16, 64, 256, 1024, 4096}
	for index := range ruleCounts {
		ruleCount := ruleCounts[index]
		b.Run(fmt.Sprintf("rules_%d_emit_1", ruleCount), func(b *testing.B) {
			benchmarkEngineRunSingleMode(b, ruleCount, 1, mode)
		})
	}
}

// benchmarkEngineRunSingleMode runs one selected policy mode benchmark.
func benchmarkEngineRunSingleMode(
	b *testing.B,
	ruleCount int,
	diagnosticsPerRule int,
	mode string,
) {
	b.Helper()

	engine := benchmarkEngine(b, ruleCount, diagnosticsPerRule)
	specs := engine.Rules()
	policy := benchmarkRunPolicy(specs)
	compiled, err := policy.Compile(specs)
	if err != nil {
		b.Fatalf("Compile() error: %v", err)
	}

	runContext := lint.RunContext{
		TargetPath: "workspace/main/source.cfg",
		TargetKind: "source.cfg",
		Values:     make(map[string]any),
	}

	options := (*RunOptions)(nil)
	switch mode {
	case "policy":
		options = &RunOptions{Policy: &policy}
	case "compiled_policy":
		options = &RunOptions{CompiledPolicy: compiled}
	case "no_policy":
		options = nil
	default:
		b.Fatalf("unknown benchmark mode: %q", mode)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := engine.RunDefault(runContext, options); err != nil {
			b.Fatalf("RunDefault(%s) error: %v", mode, err)
		}
	}
}

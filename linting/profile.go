// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

var (
	// runProfileServiceDiagnosticsDisabled stores shared explicit false pointer.
	runProfileServiceDiagnosticsDisabled = false
)

// RunProfile is canonical runtime lint execution profile.
type RunProfile struct {

	// CompiledPolicy stores precompiled policy settings for repeated runs.
	CompiledPolicy *CompiledRunPolicy `json:"-" yaml:"-" jsonschema:"-"`

	// FailOn defines minimal severity level that fails run result.
	FailOn lint.Severity `json:"fail_on,omitempty" yaml:"fail_on,omitempty" jsonschema:"default=error,enum=error,enum=warning,enum=info,enum=notice,example=error"`

	// Enabled toggles lint execution for current run.
	Enabled bool `json:"enabled" yaml:"enabled" jsonschema:"required,default=true"`

	// EnableServiceDiagnostics enables built-in lintkit policy diagnostics.
	EnableServiceDiagnostics bool `json:"enable_service_diagnostics" yaml:"enable_service_diagnostics" jsonschema:"required,default=true"`
}

// BuildRunProfileOptions controls BuildRunProfile behavior.
type BuildRunProfileOptions struct {
	// Enabled overrides final profile enabled switch.
	Enabled *bool

	// Compiler compiles path pattern matchers.
	Compiler PatternMatcherCompiler

	// EnableServiceDiagnostics overrides service diagnostics switch.
	EnableServiceDiagnostics *bool

	// FailOn overrides final profile fail threshold.
	FailOn lint.Severity

	// Overlays stores ordered policy overlays.
	Overlays []RunPolicyConfig

	// Registered stores current registered rules for strict selector checks.
	Registered []lint.RuleSpec

	// Base stores base policy config.
	Base RunPolicyConfig
}

// OverlayInput stores generic policy overlay inputs.
type OverlayInput struct {

	// FailOn overrides fail threshold when set.
	FailOn *lint.Severity

	// SoftUnknownSelectors enables soft unknown selector handling when set.
	SoftUnknownSelectors *bool
	// DisableRules stores selectors that should be disabled.
	DisableRules []string

	// Exclude stores global exclude path patterns.
	Exclude []string
}

// Normalize validates and normalizes run profile values.
func (profile RunProfile) Normalize() (RunProfile, error) {
	normalizedThreshold, err := normalizeFailOnSeverity(profile.FailOn)
	if err != nil {
		return RunProfile{}, err
	}

	out := profile
	out.FailOn = normalizedThreshold
	return out, nil
}

// ShouldFail reports whether result fails by current profile threshold.
//
// Runtime rule errors are always treated as critical failures.
func (profile RunProfile) ShouldFail(result RunResult) (bool, error) {
	normalized, err := profile.Normalize()
	if err != nil {
		return false, err
	}

	if !normalized.Enabled {
		return false, nil
	}

	return result.ShouldFail(normalized.FailOn, true)
}

// Options builds normalized low-level run options from profile.
func (profile RunProfile) Options() (RunOptions, error) {
	normalized, err := profile.Normalize()
	if err != nil {
		return RunOptions{}, err
	}

	return normalized.optionsFromNormalized(), nil
}

// optionsFromNormalized maps already normalized profile to run options.
func (profile RunProfile) optionsFromNormalized() RunOptions {
	options := RunOptions{
		CompiledPolicy: profile.CompiledPolicy,
	}
	if !profile.EnableServiceDiagnostics {
		options.EnableServiceDiagnostics = &runProfileServiceDiagnosticsDisabled
	}

	return options
}

// BuildRunProfile builds canonical runtime profile from base and overlays.
func BuildRunProfile(options BuildRunProfileOptions) (RunProfile, error) {
	merged := options.Base
	var err error

	for index := range options.Overlays {
		merged, err = MergeRunPolicyConfig(merged, options.Overlays[index])
		if err != nil {
			return RunProfile{}, fmt.Errorf("merge overlay[%d]: %w", index, err)
		}
	}

	var compiledPolicy *CompiledRunPolicy
	if requiresCompiledPolicy(merged) {
		runtimePolicy, err := merged.Build(options.Compiler)
		if err != nil {
			return RunProfile{}, fmt.Errorf("build merged policy: %w", err)
		}

		compiledPolicy, err = CompileRunPolicy(&runtimePolicy, options.Registered)
		if err != nil {
			return RunProfile{}, fmt.Errorf("compile merged policy: %w", err)
		}
	}

	finalFailOn := merged.FailOn
	if options.FailOn != "" {
		finalFailOn = options.FailOn
	}

	profile := RunProfile{
		Enabled:                  true,
		FailOn:                   finalFailOn,
		CompiledPolicy:           compiledPolicy,
		EnableServiceDiagnostics: true,
	}
	if options.Enabled != nil {
		profile.Enabled = *options.Enabled
	}

	if options.EnableServiceDiagnostics != nil {
		profile.EnableServiceDiagnostics = *options.EnableServiceDiagnostics
	}

	return profile.Normalize()
}

// requiresCompiledPolicy reports whether config needs runtime policy resolver.
func requiresCompiledPolicy(config RunPolicyConfig) bool {
	if len(config.Rules) > 0 {
		return true
	}

	if len(config.Exclude) > 0 {
		return true
	}

	return false
}

// BuildPolicyOverlay builds one generic runtime policy overlay.
func BuildPolicyOverlay(input OverlayInput) RunPolicyConfig {
	overlay := RunPolicyConfig{
		Exclude: normalizePatternList(input.Exclude),
		Rules:   make([]RunPolicyRuleConfig, 0, len(input.DisableRules)),
	}

	if input.SoftUnknownSelectors != nil {
		overlay.SoftUnknownSelectors = *input.SoftUnknownSelectors
	}

	if input.FailOn != nil {
		overlay.FailOn = *input.FailOn
	}

	seenRules := make(map[string]struct{}, len(input.DisableRules))
	for index := range input.DisableRules {
		selector := strings.TrimSpace(input.DisableRules[index])
		if selector == "" {
			continue
		}

		if _, exists := seenRules[selector]; exists {
			continue
		}

		seenRules[selector] = struct{}{}
		overlay.Rules = append(overlay.Rules, RunPolicyRuleConfig{
			Rule:    selector,
			Enabled: BoolPtr(false),
		})
	}

	if len(overlay.Rules) == 0 {
		overlay.Rules = nil
	}

	return overlay
}

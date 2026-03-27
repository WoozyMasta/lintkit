// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

// PatternMatcherCompiler builds one path matcher from raw pattern list.
type PatternMatcherCompiler func(patterns []string) (PathMatcher, error)

// RunPolicyRuleConfig is one ordered rule policy entry.
type RunPolicyRuleConfig struct {
	// Options overrides current rule options payload when set.
	Options any `json:"options,omitempty" yaml:"options,omitempty" jsonschema:"description=Arbitrary rule options payload for runner-specific behavior."`

	// Enabled overrides rule enable state when set.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" jsonschema:"example=true"`

	// Rule is rule selector token.
	// Supported forms: `*`, `<module>.*`, `<module>.<scope>.*`, `<rule-id>`, `<CODE>`.
	Rule string `json:"rule,omitempty" yaml:"rule,omitempty" jsonschema:"required,example=*,example=module_alpha.*,example=module_alpha.parse.*,example=MODULE2001,example=module_alpha.parse.rule-name"`

	// Severity overrides effective diagnostic severity when set.
	Severity lint.Severity `json:"severity,omitempty" yaml:"severity,omitempty" jsonschema:"enum=error,enum=warning,enum=info,enum=notice,example=error"`

	// Exclude stores optional path patterns where this entry is not applied.
	Exclude []string `json:"exclude,omitempty" yaml:"exclude,omitempty" jsonschema:"example=**/generated/**,example=**/vendor/**"`
}

// RunPolicyConfig is schema-friendly policy model for config files.
type RunPolicyConfig struct {

	// FailOn defines minimum diagnostic severity that fails run result.
	// Empty value defaults to "error".
	FailOn lint.Severity `json:"fail_on,omitempty" yaml:"fail_on,omitempty" jsonschema:"default=error,enum=error,enum=warning,enum=info,enum=notice,example=error"`
	// Exclude stores optional global exclude path patterns.
	Exclude []string `json:"exclude,omitempty" yaml:"exclude,omitempty" jsonschema:"example=**/vendor/**,example=**/*.generated.*"`

	// Rules stores ordered selector-based rule settings.
	// Later entries override earlier entries.
	Rules []RunPolicyRuleConfig `json:"rules,omitempty" yaml:"rules,omitempty"`

	// SoftUnknownSelectors enables soft mode for unknown selectors.
	// Default false means unknown selectors fail build/compile.
	SoftUnknownSelectors bool `json:"soft_unknown_selectors,omitempty" yaml:"soft_unknown_selectors,omitempty" jsonschema:"default=false"`
}

// SchemaModel is schema root for lintkit public contract documentation.
//
// Use schemadoc `mod2schema` with this type name to generate JSON Schema.
type SchemaModel struct {
	// RuleSettings is one selector settings object.
	RuleSettings RuleSettings `json:"rule_settings" jsonschema:"required"`

	// RuleSpec is one rule metadata contract sample.
	RuleSpec lint.RuleSpec `json:"rule_spec" jsonschema:"required"`

	// RunResult is runtime output payload model.
	RunResult RunResult `json:"run_result" jsonschema:"required"`

	// RegistrySnapshot is exported registry payload model.
	RegistrySnapshot lint.RegistrySnapshot `json:"registry_snapshot" jsonschema:"required"`

	// RunPolicyConfig is schema-friendly lint policy config model.
	RunPolicyConfig RunPolicyConfig `json:"run_policy_config" jsonschema:"required"`

	// RunContext is runtime target context model.
	RunContext lint.RunContext `json:"run_context" jsonschema:"required"`

	// RunOptions is runtime execution options model.
	RunOptions RunOptions `json:"run_options" jsonschema:"required"`

	// Diagnostic is one normalized lint finding contract sample.
	Diagnostic lint.Diagnostic `json:"diagnostic" jsonschema:"required"`
}

// Build converts schema-friendly config model to runtime RunPolicy.
func (config RunPolicyConfig) Build(compiler PatternMatcherCompiler) (RunPolicy, error) {
	if err := validateRunPolicyConfigFailOn(config.FailOn); err != nil {
		return RunPolicy{}, err
	}

	policy := RunPolicy{
		Strict:      !config.SoftUnknownSelectors,
		RuleEntries: make([]PolicyRuleEntry, 0, len(config.Rules)),
	}

	excludePatterns := normalizePatternList(config.Exclude)
	if len(excludePatterns) > 0 {
		if compiler == nil {
			return RunPolicy{}, ErrNilPatternMatcherCompiler
		}

		matcher, err := compiler(excludePatterns)
		if err != nil {
			return RunPolicy{}, fmt.Errorf("compile exclude matcher: %w", err)
		}

		policy.Exclude = matcher
	}

	for index := range config.Rules {
		entry, err := buildRuleConfigEntry(
			compiler,
			index,
			config.Rules[index],
		)
		if err != nil {
			return RunPolicy{}, err
		}

		policy.RuleEntries = append(policy.RuleEntries, entry)
	}

	if len(policy.RuleEntries) == 0 {
		policy.RuleEntries = nil
	}

	return policy, nil
}

// ShouldFail reports whether run result should fail by this policy config.
//
// Runtime rule errors are always treated as critical failures.
func (config RunPolicyConfig) ShouldFail(result RunResult) (bool, error) {
	if err := validateRunPolicyConfigFailOn(config.FailOn); err != nil {
		return false, err
	}

	return result.ShouldFail(config.FailOn, true)
}

// buildRuleConfigEntry builds one ordered runtime rule entry from config.
func buildRuleConfigEntry(
	compiler PatternMatcherCompiler,
	index int,
	config RunPolicyRuleConfig,
) (PolicyRuleEntry, error) {
	rule := strings.TrimSpace(config.Rule)
	if rule == "" {
		return PolicyRuleEntry{}, fmt.Errorf(
			"%w: rules[%d].rule is empty",
			ErrInvalidRunPolicy,
			index,
		)
	}

	settings := RuleSettings{
		Options:  config.Options,
		Enabled:  config.Enabled,
		Severity: config.Severity,
	}

	normalizedRules, err := NormalizeRuleSettingsMap(
		map[string]RuleSettings{rule: settings},
	)
	if err != nil {
		return PolicyRuleEntry{}, fmt.Errorf("rules[%d]: %w", index, err)
	}

	selector := ""
	setting := RuleSettings{}
	for key, value := range normalizedRules {
		selector = key
		setting = value
		break
	}

	entry := PolicyRuleEntry{
		Selector: selector,
		Settings: setting,
	}

	excludePatterns := normalizePatternList(config.Exclude)
	if len(excludePatterns) == 0 {
		return entry, nil
	}

	if compiler == nil {
		return PolicyRuleEntry{}, fmt.Errorf("rules[%d]: %w", index, ErrNilPatternMatcherCompiler)
	}

	excludeMatcher, err := compiler(excludePatterns)
	if err != nil {
		return PolicyRuleEntry{}, fmt.Errorf("compile rules[%d] exclude matcher: %w", index, err)
	}

	entry.Matcher = PathMatcherFunc(func(path string, isDir bool) bool {
		return !excludeMatcher.Match(path, isDir)
	})

	return entry, nil
}

// normalizePatternList trims empty entries and deduplicates patterns in order.
func normalizePatternList(patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(patterns))
	out := make([]string, 0, len(patterns))
	for index := range patterns {
		pattern := strings.TrimSpace(patterns[index])
		if pattern == "" {
			continue
		}

		if _, exists := seen[pattern]; exists {
			continue
		}

		seen[pattern] = struct{}{}
		out = append(out, pattern)
	}

	return out
}

// validateRunPolicyConfigFailOn validates config fail threshold severity.
func validateRunPolicyConfigFailOn(value lint.Severity) error {
	if value == "" {
		return nil
	}

	if !isSupportedSeverity(value) {
		return fmt.Errorf("%w: fail_on %q", ErrInvalidRunPolicy, value)
	}

	return nil
}

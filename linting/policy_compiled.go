// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

// CompiledRunPolicy stores prevalidated and preindexed policy settings.
//
// Use compiled policy for repeated runs over many files with the same
// effective rule set.
type CompiledRunPolicy struct {
	// ServiceDiagnostics stores one-time policy diagnostics for callers.
	ServiceDiagnostics []lint.Diagnostic

	// Rules stores preindexed global selector settings.
	Rules compiledRuleSettings

	// RuleEntries stores preindexed ordered selector entries.
	RuleEntries []compiledRuleEntry

	// Include is optional path allow-list matcher.
	Include PathMatcher `json:"-" yaml:"-"`

	// Exclude is optional path block-list matcher.
	Exclude PathMatcher `json:"-" yaml:"-"`

	// Enabled controls global lint execution enablement.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" jsonschema:"default=true,example=true"`

	// Overrides stores preindexed path-scoped selector settings.
	Overrides []CompiledPolicyOverride
}

// CompiledPolicyOverride stores preindexed path-scoped settings.
type CompiledPolicyOverride struct {
	// Rules stores preindexed selector settings for this override.
	Rules compiledRuleSettings

	// Matcher decides whether this override applies to current path.
	Matcher PathMatcher `json:"-" yaml:"-"`

	// Name is optional human-readable override label.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// compiledRuleEntry stores one preparsed ordered selector entry.
type compiledRuleEntry struct {
	// Matcher decides whether this entry applies to current path.
	Matcher PathMatcher

	// Settings stores one selector settings payload.
	Settings RuleSettings

	// Selector stores parsed selector.
	Selector parsedRuleSelector
}

// compiledRuleSettings stores preindexed selector settings maps.
type compiledRuleSettings struct {
	// All stores wildcard selector settings ("*"), when set.
	All *RuleSettings

	// ByModule stores module wildcard selector settings ("<module>.*").
	ByModule map[string]RuleSettings

	// ByScope stores module scope selector settings ("<module>.<scope>.*").
	ByScope map[string]RuleSettings

	// ByCode stores public code selector settings ("<CODE>").
	ByCode map[string]RuleSettings

	// ByRule stores exact selector settings ("<module>.<rule_name>").
	ByRule map[string]RuleSettings
}

// Compile builds prevalidated and preindexed policy object for one rule set.
func (policy *RunPolicy) Compile(registered []lint.RuleSpec) (*CompiledRunPolicy, error) {
	return CompileRunPolicy(policy, registered)
}

// CompileRunPolicy builds prevalidated and preindexed policy object.
func CompileRunPolicy(
	policy *RunPolicy,
	registered []lint.RuleSpec,
) (*CompiledRunPolicy, error) {
	if policy == nil {
		return nil, nil
	}

	var knownRuleIDs map[string]struct{}
	var knownModules map[string]struct{}
	var knownScopes map[string]struct{}
	var knownCodeCounts map[string]int
	if policy.Strict {
		knownRuleIDs, knownModules, knownScopes = knownSelectors(registered)
		knownCodeCounts = knownCodeSelectors(registered)
	}

	if err := validateCompiledPolicyInput(
		policy,
		knownRuleIDs,
		knownModules,
		knownScopes,
		knownCodeCounts,
	); err != nil {
		return nil, err
	}

	compiled := &CompiledRunPolicy{
		ServiceDiagnostics: collectPolicyServiceDiagnostics(policy, registered),
		Enabled:            cloneBoolPtr(policy.Enabled),
		Include:            policy.Include,
		Exclude:            policy.Exclude,
		Rules:              compileRuleSettingsMap(policy.Rules),
		RuleEntries:        compileRuleEntries(policy.RuleEntries),
		Overrides:          make([]CompiledPolicyOverride, 0, len(policy.Overrides)),
	}

	for index := range policy.Overrides {
		compiled.Overrides = append(compiled.Overrides, CompiledPolicyOverride{
			Matcher: policy.Overrides[index].Matcher,
			Name:    strings.TrimSpace(policy.Overrides[index].Name),
			Rules:   compileRuleSettingsMap(policy.Overrides[index].Rules),
		})
	}

	return compiled, nil
}

// PathEnabled reports whether policy allows lint run for one path.
func (policy *CompiledRunPolicy) PathEnabled(path string, isDir bool) bool {
	if policy == nil {
		return true
	}

	enabled := true
	if policy.Enabled != nil {
		enabled = *policy.Enabled
	}

	if !enabled {
		return false
	}

	if policy.Include != nil && !policy.Include.Match(path, isDir) {
		return false
	}

	if policy.Exclude != nil && policy.Exclude.Match(path, isDir) {
		return false
	}

	return true
}

// Resolve returns effective rule settings for one rule and path.
func (policy *CompiledRunPolicy) Resolve(
	rule lint.RuleSpec,
	path string,
	isDir bool,
) RuleDecision {
	enabled := true
	if rule.DefaultEnabled != nil {
		enabled = *rule.DefaultEnabled
	}

	decision := RuleDecision{
		Enabled:  enabled,
		Severity: rule.DefaultSeverity,
		Options:  rule.DefaultOptions,
	}
	if decision.Severity == "" {
		decision.Severity = lint.SeverityWarning
	}

	if policy == nil {
		return decision
	}

	if policy.Enabled != nil {
		decision.Enabled = *policy.Enabled
	}

	decision = applyCompiledRuleSettings(decision, rule, policy.Rules)
	decision = applyCompiledRuleEntries(decision, rule, path, isDir, policy.RuleEntries)

	for index := range policy.Overrides {
		if !matchesCompiledPolicyOverride(policy.Overrides[index], path, isDir) {
			continue
		}

		decision = applyCompiledRuleSettings(
			decision,
			rule,
			policy.Overrides[index].Rules,
		)
	}

	return decision
}

// applyCompiledRuleEntries applies precompiled ordered selector entries.
func applyCompiledRuleEntries(
	decision RuleDecision,
	rule lint.RuleSpec,
	path string,
	isDir bool,
	entries []compiledRuleEntry,
) RuleDecision {
	if len(entries) == 0 {
		return decision
	}

	for index := range entries {
		if entries[index].Matcher != nil && !entries[index].Matcher.Match(path, isDir) {
			continue
		}

		if !selectorMatchesRule(entries[index].Selector, rule) {
			continue
		}

		decision = applyRuleSettings(decision, entries[index].Settings)
	}

	return decision
}

// matchesCompiledPolicyOverride reports whether override applies to path.
func matchesCompiledPolicyOverride(
	override CompiledPolicyOverride,
	path string,
	isDir bool,
) bool {
	if override.Matcher == nil {
		return true
	}

	return override.Matcher.Match(path, isDir)
}

// applyCompiledRuleSettings applies preindexed selector settings.
func applyCompiledRuleSettings(
	decision RuleDecision,
	rule lint.RuleSpec,
	settings compiledRuleSettings,
) RuleDecision {
	chain := resolveRuleSettingsChainFromCompiled(rule, settings)
	return applyRuleSettingsChain(decision, chain)
}

// compileRuleSettingsMap preindexes selector settings by selector kind.
func compileRuleSettingsMap(settings map[string]RuleSettings) compiledRuleSettings {
	compiled := compiledRuleSettings{}

	for rawSelector, setting := range settings {
		selector, err := parseRuleSelector(rawSelector)
		if err != nil {
			// compile path receives already validated settings maps.
			continue
		}

		switch selector.kind {
		case selectorKindAll:
			item := cloneRuleSettings(setting)
			compiled.All = &item
		case selectorKindModule:
			if compiled.ByModule == nil {
				compiled.ByModule = make(map[string]RuleSettings)
			}

			compiled.ByModule[selector.module] = cloneRuleSettings(setting)
		case selectorKindScope:
			if compiled.ByScope == nil {
				compiled.ByScope = make(map[string]RuleSettings)
			}

			compiled.ByScope[selector.module+"."+selector.scope] = cloneRuleSettings(setting)
		case selectorKindCode:
			if compiled.ByCode == nil {
				compiled.ByCode = make(map[string]RuleSettings)
			}

			compiled.ByCode[selector.raw] = cloneRuleSettings(setting)
		default:
			if compiled.ByRule == nil {
				compiled.ByRule = make(map[string]RuleSettings)
			}

			compiled.ByRule[selector.ruleID] = cloneRuleSettings(setting)
		}
	}

	return compiled
}

// compileRuleEntries precompiles ordered selector entry list.
func compileRuleEntries(entries []PolicyRuleEntry) []compiledRuleEntry {
	if len(entries) == 0 {
		return nil
	}

	out := make([]compiledRuleEntry, 0, len(entries))
	for index := range entries {
		selector, err := parseRuleSelector(entries[index].Selector)
		if err != nil {
			// compile path receives already validated settings maps.
			continue
		}

		out = append(out, compiledRuleEntry{
			Selector: selector,
			Settings: cloneRuleSettings(entries[index].Settings),
			Matcher:  entries[index].Matcher,
		})
	}

	return out
}

// cloneRuleSettings returns settings copy with detached bool pointer value.
func cloneRuleSettings(setting RuleSettings) RuleSettings {
	out := setting
	out.Enabled = cloneBoolPtr(setting.Enabled)
	out.Options = cloneDynamicValue(setting.Options)
	return out
}

// cloneBoolPtr returns copied bool pointer value.
func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}

	out := *value
	return &out
}

// knownSelectors returns known strict selector lookup maps.
func knownSelectors(registered []lint.RuleSpec) (
	map[string]struct{},
	map[string]struct{},
	map[string]struct{},
) {
	knownRuleIDs := make(map[string]struct{}, len(registered))
	knownModules := make(map[string]struct{}, len(registered))
	knownScopes := make(map[string]struct{}, len(registered))

	for index := range registered {
		knownRuleIDs[registered[index].ID] = struct{}{}
		knownModules[registered[index].Module] = struct{}{}
		scope := registered[index].Scope
		if scope != "" {
			knownScopes[registered[index].Module+"."+scope] = struct{}{}
		}
	}

	return knownRuleIDs, knownModules, knownScopes
}

// validateCompiledPolicyInput validates global and override settings maps.
func validateCompiledPolicyInput(
	policy *RunPolicy,
	knownRuleIDs map[string]struct{},
	knownModules map[string]struct{},
	knownScopes map[string]struct{},
	knownCodeCounts map[string]int,
) error {
	if policy == nil {
		return nil
	}

	var err error

	if policy.Strict {
		err = validateRuleSettingsMap(
			policy.Rules,
			knownRuleIDs,
			knownModules,
			knownScopes,
			knownCodeCounts,
		)
	} else {
		err = validateRuleSettingsMap(policy.Rules, nil, nil, nil, nil)
	}

	if err != nil {
		return err
	}

	if policy.Strict {
		err = validateRuleEntries(
			policy.RuleEntries,
			knownRuleIDs,
			knownModules,
			knownScopes,
			knownCodeCounts,
		)
	} else {
		err = validateRuleEntries(policy.RuleEntries, nil, nil, nil, nil)
	}

	if err != nil {
		return err
	}

	for index := range policy.Overrides {
		if policy.Strict {
			err = validateRuleSettingsMap(
				policy.Overrides[index].Rules,
				knownRuleIDs,
				knownModules,
				knownScopes,
				knownCodeCounts,
			)
		} else {
			err = validateRuleSettingsMap(policy.Overrides[index].Rules, nil, nil, nil, nil)
		}

		if err != nil {
			return fmt.Errorf("override[%d]: %w", index, err)
		}
	}

	return nil
}

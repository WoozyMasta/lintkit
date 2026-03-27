// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"maps"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// RuleSelectorAll is wildcard selector for all registered rules.
	RuleSelectorAll = "*"
)

// PathMatcher reports whether one path matches policy condition.
type PathMatcher interface {
	// Match reports whether provided path and dir flag satisfy matcher.
	Match(path string, isDir bool) bool
}

// PathMatcherFunc adapts function values to PathMatcher interface.
type PathMatcherFunc func(path string, isDir bool) bool

// Match reports function matcher result.
func (fn PathMatcherFunc) Match(path string, isDir bool) bool {
	if fn == nil {
		return false
	}

	return fn(path, isDir)
}

// RuleSettings stores optional rule enablement and severity overrides.
type RuleSettings struct {

	// Options overrides current rule options payload when set.
	Options any `json:"options,omitempty" yaml:"options,omitempty" jsonschema:"description=Arbitrary rule options payload for runner-specific behavior."`
	// Enabled overrides rule enable state when set.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" jsonschema:"example=true"`

	// Severity overrides effective diagnostic severity when set.
	Severity lint.Severity `json:"severity,omitempty" yaml:"severity,omitempty" jsonschema:"enum=error,enum=warning,enum=info,enum=notice,example=error"`
}

// PolicyOverride stores path-scoped rule settings overrides.
type PolicyOverride struct {
	// Matcher decides whether this override applies to current path.
	Matcher PathMatcher `json:"-" yaml:"-"`

	// Rules stores rule settings by selector.
	// Supported selectors:
	//   - "*": all rules
	//   - "<module>.*": all rules in one module namespace
	//   - "<module>.<scope>.*": all rules in one module scope
	//   - "<module>.<rule_name>": one exact rule ID
	//   - "<CODE>": one exact public code token (for example: "RVMAT2028")
	Rules map[string]RuleSettings `json:"rules,omitempty" yaml:"rules,omitempty"`

	// Name is optional human-readable override label.
	Name string `json:"name,omitempty" yaml:"name,omitempty" jsonschema:"example=Disable noisy warning in vendor subtree"`
}

// PolicyRuleEntry stores one ordered selector rule settings entry.
type PolicyRuleEntry struct {
	// Matcher decides whether this entry applies to current path.
	Matcher PathMatcher `json:"-" yaml:"-"`

	// Selector stores one normalized selector token.
	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`

	// Settings stores resolved selector settings payload.
	Settings RuleSettings `json:"settings" yaml:"settings"`
}

// RunPolicy stores global and path-scoped rule execution settings.
type RunPolicy struct {
	// Enabled controls global lint execution enablement.
	// Nil means "enabled".
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" jsonschema:"default=true,example=true"`

	// Include is optional path allow-list matcher.
	// When set and path does not match, no rules are executed.
	Include PathMatcher `json:"-" yaml:"-"`

	// Exclude is optional path block-list matcher.
	// When set and path matches, no rules are executed.
	Exclude PathMatcher `json:"-" yaml:"-"`

	// Rules stores global rule settings by selector.
	// Supported selectors:
	//   - "*": all rules
	//   - "<module>.*": all rules in one module namespace
	//   - "<module>.<scope>.*": all rules in one module scope
	//   - "<module>.<rule_name>": one exact rule ID
	//   - "<CODE>": one exact public code token (for example: "RVMAT2028")
	Rules map[string]RuleSettings `json:"rules,omitempty" yaml:"rules,omitempty"`

	// RuleEntries stores ordered selector settings entries.
	// Later entries override earlier entries.
	RuleEntries []PolicyRuleEntry `json:"rule_entries,omitempty" yaml:"rule_entries,omitempty"`

	// Overrides stores path-scoped rule settings.
	// Order matters: later overrides win.
	Overrides []PolicyOverride `json:"overrides,omitempty" yaml:"overrides,omitempty"`

	// Strict enables selector validation against registered rules.
	// When true, unknown selectors are returned as errors before rule run.
	Strict bool `json:"strict,omitempty" yaml:"strict,omitempty" jsonschema:"default=false"`
}

// RuleDecision stores resolved effective settings for one rule and path.
type RuleDecision struct {

	// Options stores resolved current rule options payload.
	Options any `json:"options,omitempty" yaml:"options,omitempty" jsonschema:"description=Resolved arbitrary rule options payload."`
	// Severity is effective diagnostic severity.
	Severity lint.Severity `json:"severity" yaml:"severity" jsonschema:"required,enum=error,enum=warning,enum=info,enum=notice"`

	// Enabled is effective rule enable state.
	Enabled bool `json:"enabled" yaml:"enabled" jsonschema:"required"`
}

// BoolPtr allocates one bool pointer for optional settings fields.
func BoolPtr(value bool) *bool {
	return &value
}

// SetAll configures global wildcard selector ("*").
func (policy *RunPolicy) SetAll(settings RuleSettings) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	return policy.setSelector(RuleSelectorAll, settings)
}

// SetModule configures module wildcard selector ("<module>.*").
func (policy *RunPolicy) SetModule(module string, settings RuleSettings) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	selector, err := ModuleSelector(module)
	if err != nil {
		return err
	}

	return policy.setSelector(selector, settings)
}

// SetScope configures module scope selector ("<module>.<scope>.*").
func (policy *RunPolicy) SetScope(
	module string,
	scope string,
	settings RuleSettings,
) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	selector, err := ScopeSelector(module, scope)
	if err != nil {
		return err
	}

	return policy.setSelector(selector, settings)
}

// SetRule configures exact rule selector ("<module>.<rule_name>").
func (policy *RunPolicy) SetRule(ruleID string, settings RuleSettings) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	selector, err := RuleSelector(ruleID)
	if err != nil {
		return err
	}

	return policy.setSelector(selector, settings)
}

// SetCode configures public code selector ("<CODE>").
func (policy *RunPolicy) SetCode(code string, settings RuleSettings) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	selector, err := CodeSelector(code)
	if err != nil {
		return err
	}

	return policy.setSelector(selector, settings)
}

// AddOverride appends one path-scoped override in declaration order.
func (policy *RunPolicy) AddOverride(override PolicyOverride) error {
	if policy == nil {
		return ErrNilRunPolicy
	}

	if err := validatePolicyOverrideSelectors(override); err != nil {
		return err
	}

	policy.Overrides = append(policy.Overrides, override)
	return nil
}

// SetAll configures override wildcard selector ("*").
func (override *PolicyOverride) SetAll(settings RuleSettings) error {
	if override == nil {
		return ErrNilPolicyOverride
	}

	return override.setSelector(RuleSelectorAll, settings)
}

// SetModule configures override module wildcard selector ("<module>.*").
func (override *PolicyOverride) SetModule(module string, settings RuleSettings) error {
	if override == nil {
		return ErrNilPolicyOverride
	}

	selector, err := ModuleSelector(module)
	if err != nil {
		return err
	}

	return override.setSelector(selector, settings)
}

// SetScope configures override scope selector ("<module>.<scope>.*").
func (override *PolicyOverride) SetScope(
	module string,
	scope string,
	settings RuleSettings,
) error {
	if override == nil {
		return ErrNilPolicyOverride
	}

	selector, err := ScopeSelector(module, scope)
	if err != nil {
		return err
	}

	return override.setSelector(selector, settings)
}

// SetRule configures override exact rule selector ("<module>.<rule_name>").
func (override *PolicyOverride) SetRule(ruleID string, settings RuleSettings) error {
	if override == nil {
		return ErrNilPolicyOverride
	}

	selector, err := RuleSelector(ruleID)
	if err != nil {
		return err
	}

	return override.setSelector(selector, settings)
}

// SetCode configures override public code selector ("<CODE>").
func (override *PolicyOverride) SetCode(code string, settings RuleSettings) error {
	if override == nil {
		return ErrNilPolicyOverride
	}

	selector, err := CodeSelector(code)
	if err != nil {
		return err
	}

	return override.setSelector(selector, settings)
}

// PathEnabled reports whether policy allows lint run for one path.
func (policy *RunPolicy) PathEnabled(path string, isDir bool) bool {
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
func (policy *RunPolicy) Resolve(rule lint.RuleSpec, path string, isDir bool) RuleDecision {
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

	decision = applyRuleSettingsMap(decision, rule, policy.Rules)
	decision = applyRuleEntryList(decision, rule, path, isDir, policy.RuleEntries)

	for index := range policy.Overrides {
		if !matchesPolicyOverride(policy.Overrides[index], path, isDir) {
			continue
		}

		decision = applyRuleSettingsMap(decision, rule, policy.Overrides[index].Rules)
	}

	return decision
}

// Validate checks policy selector and severity validity for known rules.
func (policy *RunPolicy) Validate(registered []lint.RuleSpec) error {
	if policy == nil {
		return nil
	}

	knownRuleIDs := make(map[string]struct{}, len(registered))
	knownModules := make(map[string]struct{}, len(registered))
	knownScopes := make(map[string]struct{}, len(registered))
	for index := range registered {
		knownRuleIDs[registered[index].ID] = struct{}{}
		knownModules[registered[index].Module] = struct{}{}
		scope := strings.TrimSpace(registered[index].Scope)
		if scope != "" {
			knownScopes[registered[index].Module+"."+scope] = struct{}{}
		}
	}

	knownCodeCounts := knownCodeSelectors(registered)

	if err := validateRuleSettingsMap(
		policy.Rules,
		knownRuleIDs,
		knownModules,
		knownScopes,
		knownCodeCounts,
	); err != nil {
		return err
	}

	if err := validateRuleEntries(
		policy.RuleEntries,
		knownRuleIDs,
		knownModules,
		knownScopes,
		knownCodeCounts,
	); err != nil {
		return err
	}

	for index := range policy.Overrides {
		if err := validateRuleSettingsMap(
			policy.Overrides[index].Rules,
			knownRuleIDs,
			knownModules,
			knownScopes,
			knownCodeCounts,
		); err != nil {
			return fmt.Errorf("override[%d]: %w", index, err)
		}
	}

	return nil
}

// applyRuleEntryList applies ordered selector entries for one path and rule.
func applyRuleEntryList(
	decision RuleDecision,
	rule lint.RuleSpec,
	path string,
	isDir bool,
	entries []PolicyRuleEntry,
) RuleDecision {
	if len(entries) == 0 {
		return decision
	}

	for index := range entries {
		if entries[index].Matcher != nil && !entries[index].Matcher.Match(path, isDir) {
			continue
		}

		selector, err := parseRuleSelector(entries[index].Selector)
		if err != nil {
			continue
		}

		if !selectorMatchesRule(selector, rule) {
			continue
		}

		decision = applyRuleSettings(decision, entries[index].Settings)
	}

	return decision
}

// validateRuleEntries validates ordered selector entries.
func validateRuleEntries(
	entries []PolicyRuleEntry,
	knownRuleIDs map[string]struct{},
	knownModules map[string]struct{},
	knownScopes map[string]struct{},
	knownCodeCounts map[string]int,
) error {
	if len(entries) == 0 {
		return nil
	}

	settings := make(map[string]RuleSettings, 1)
	for index := range entries {
		selector := strings.TrimSpace(entries[index].Selector)
		if selector == "" {
			return fmt.Errorf("%w: rule_entries[%d].selector is empty", ErrInvalidRunPolicy, index)
		}

		settings[selector] = entries[index].Settings
		err := validateRuleSettingsMap(
			settings,
			knownRuleIDs,
			knownModules,
			knownScopes,
			knownCodeCounts,
		)
		delete(settings, selector)
		if err != nil {
			return fmt.Errorf("rule_entries[%d]: %w", index, err)
		}
	}

	return nil
}

// selectorMatchesRule reports whether parsed selector matches one rule.
func selectorMatchesRule(selector parsedRuleSelector, rule lint.RuleSpec) bool {
	switch selector.kind {
	case selectorKindAll:
		return true
	case selectorKindModule:
		return rule.Module == selector.module
	case selectorKindScope:
		return rule.Module == selector.module && rule.Scope == selector.scope
	case selectorKindRule:
		return rule.ID == selector.ruleID
	case selectorKindCode:
		return rule.Code == selector.raw
	default:
		return false
	}
}

// RuleSelector validates and returns exact rule selector.
func RuleSelector(ruleID string) (string, error) {
	selector, err := parseRuleSelector(ruleID)
	if err != nil {
		return "", err
	}

	if selector.kind != selectorKindRule {
		return "", fmt.Errorf("%w: invalid rule selector %q", ErrInvalidRunPolicy, strings.TrimSpace(ruleID))
	}

	return selector.ruleID, nil
}

// ModuleSelector validates and returns module wildcard selector.
func ModuleSelector(module string) (string, error) {
	selector, err := parseRuleSelector(strings.TrimSpace(module) + ".*")
	if err != nil {
		return "", err
	}

	if selector.kind != selectorKindModule {
		return "", fmt.Errorf("%w: invalid module selector %q", ErrInvalidRunPolicy, strings.TrimSpace(module))
	}

	return selector.module + ".*", nil
}

// ScopeSelector validates and returns scope wildcard selector.
func ScopeSelector(module string, scope string) (string, error) {
	selector, err := parseRuleSelector(
		strings.TrimSpace(module) + "." + strings.TrimSpace(scope) + ".*",
	)
	if err != nil {
		return "", err
	}

	if selector.kind != selectorKindScope {
		return "", fmt.Errorf(
			"%w: invalid scope selector %q",
			ErrInvalidRunPolicy,
			strings.TrimSpace(module)+"."+strings.TrimSpace(scope),
		)
	}

	return selector.module + "." + selector.scope + ".*", nil
}

// CodeSelector validates and returns public code selector.
func CodeSelector(code string) (string, error) {
	normalizedCode := strings.TrimSpace(code)
	selector, err := parseRuleSelector(normalizedCode)
	if err != nil {
		return "", err
	}

	if selector.kind != selectorKindCode {
		return "", fmt.Errorf(
			"%w: invalid code selector %q",
			ErrInvalidRunPolicy,
			strings.TrimSpace(code),
		)
	}

	return selector.raw, nil
}

// applyRuleSettingsMap applies merged selector settings to one decision.
func applyRuleSettingsMap(
	decision RuleDecision,
	rule lint.RuleSpec,
	settings map[string]RuleSettings,
) RuleDecision {
	chain := resolveRuleSettingsChainFromMap(rule, settings)
	return applyRuleSettingsChain(decision, chain)
}

// applyRuleSettings merges one rule settings item into current decision.
func applyRuleSettings(decision RuleDecision, settings RuleSettings) RuleDecision {
	if settings.Enabled != nil {
		decision.Enabled = *settings.Enabled
	}

	if settings.Severity != "" {
		decision.Severity = settings.Severity
	}

	if settings.Options != nil {
		decision.Options = cloneDynamicValue(settings.Options)
	}

	return decision
}

// matchesPolicyOverride reports whether one override applies to current path.
func matchesPolicyOverride(override PolicyOverride, path string, isDir bool) bool {
	if override.Matcher == nil {
		return true
	}

	return override.Matcher.Match(path, isDir)
}

// setSelector validates and assigns one policy selector settings entry.
func (policy *RunPolicy) setSelector(selector string, settings RuleSettings) error {
	if policy.Rules == nil {
		policy.Rules = make(map[string]RuleSettings)
	}

	normalized, err := NormalizeRuleSettingsMap(
		map[string]RuleSettings{selector: settings},
	)
	if err != nil {
		return err
	}

	maps.Copy(policy.Rules, normalized)

	return nil
}

// setSelector validates and assigns one override selector settings entry.
func (override *PolicyOverride) setSelector(selector string, settings RuleSettings) error {
	if override.Rules == nil {
		override.Rules = make(map[string]RuleSettings)
	}

	normalized, err := NormalizeRuleSettingsMap(
		map[string]RuleSettings{selector: settings},
	)
	if err != nil {
		return err
	}

	maps.Copy(override.Rules, normalized)

	return nil
}

// validatePolicyOverrideSelectors validates selector syntax for one override.
func validatePolicyOverrideSelectors(override PolicyOverride) error {
	return validateRuleSettingsMap(override.Rules, nil, nil, nil, nil)
}

// validateRuleSettingsMap validates selector keys and settings values.
func validateRuleSettingsMap(
	settings map[string]RuleSettings,
	knownRuleIDs map[string]struct{},
	knownModules map[string]struct{},
	knownScopes map[string]struct{},
	knownCodeCounts map[string]int,
) error {
	strictSelectors := len(knownRuleIDs) > 0 ||
		len(knownModules) > 0 ||
		len(knownScopes) > 0 ||
		len(knownCodeCounts) > 0

	for rawSelector, setting := range settings {
		selector, err := parseRuleSelector(rawSelector)
		if err != nil {
			return err
		}

		if setting.Severity != "" && !isSupportedSeverity(setting.Severity) {
			return fmt.Errorf(
				"%w: selector %q has invalid severity %q",
				ErrInvalidRunPolicy,
				selector.raw,
				setting.Severity,
			)
		}

		if selector.kind == selectorKindAll {
			continue
		}

		if strictSelectors {
			switch selector.kind {
			case selectorKindModule:
				if _, exists := knownModules[selector.module]; !exists {
					return fmt.Errorf("%w: unknown selector %q", ErrUnknownRuleSelector, selector.raw)
				}
			case selectorKindScope:
				key := selector.module + "." + selector.scope
				if _, exists := knownScopes[key]; !exists {
					return fmt.Errorf("%w: unknown selector %q", ErrUnknownRuleSelector, selector.raw)
				}
			case selectorKindRule:
				if _, exists := knownRuleIDs[selector.ruleID]; !exists {
					return fmt.Errorf("%w: unknown selector %q", ErrUnknownRuleSelector, selector.raw)
				}
			case selectorKindCode:
				count := knownCodeCounts[selector.raw]
				if count == 0 {
					return fmt.Errorf("%w: unknown selector %q", ErrUnknownRuleSelector, selector.raw)
				}

				if count > 1 {
					return fmt.Errorf("%w: ambiguous selector %q", ErrAmbiguousRuleSelector, selector.raw)
				}
			}
		}
	}

	return nil
}

// knownCodeSelectors builds strict-lookup maps for code-based selectors.
func knownCodeSelectors(registered []lint.RuleSpec) map[string]int {
	var knownCodeCounts map[string]int

	for index := range registered {
		normalizedCode, ok := normalizePolicyCode(registered[index].Code)
		if !ok {
			continue
		}

		if knownCodeCounts == nil {
			knownCodeCounts = make(map[string]int, 8)
		}

		knownCodeCounts[normalizedCode]++
	}

	return knownCodeCounts
}

// normalizePolicyCode returns normalized public code selector token.
func normalizePolicyCode(value string) (string, bool) {
	normalized := strings.TrimSpace(value)
	if !isCodeSelectorToken(normalized) {
		return "", false
	}

	code, ok := lint.ParsePublicCode(normalized)
	if !ok || code == 0 {
		return "", false
	}

	return normalized, true
}

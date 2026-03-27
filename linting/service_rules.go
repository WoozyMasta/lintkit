// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// serviceRulesModuleID is module id for lintkit built-in policy diagnostics.
	serviceRulesModuleID = "lintkit"

	// serviceRulesScope is scope token for lintkit built-in policy diagnostics.
	serviceRulesScope = "policy"

	// serviceRulesScopeDescription describes built-in policy diagnostics scope.
	serviceRulesScopeDescription = "Lint policy configuration diagnostics."
)

var (
	// serviceRuleSpecs stores immutable metadata for built-in service rules.
	serviceRuleSpecs = []lint.RuleSpec{
		{
			ID:               ruleIDPolicyDeprecatedSelector,
			Module:           serviceRulesModuleID,
			Scope:            serviceRulesScope,
			ScopeDescription: serviceRulesScopeDescription,
			Code:             codePolicyDeprecatedSelector,
			Message:          "policy selector references deprecated rule",
			Description: "Reports a selector that resolves to at least one " +
				"deprecated rule.",
			DefaultSeverity: lint.SeverityWarning,
		},
		{
			ID:               ruleIDPolicyUnknownSelector,
			Module:           serviceRulesModuleID,
			Scope:            serviceRulesScope,
			ScopeDescription: serviceRulesScopeDescription,
			Code:             codePolicyUnknownSelector,
			Message:          "unknown policy selector",
			Description: "Reports a selector that matches no registered rules " +
				"in soft unknown selector mode.",
			DefaultSeverity: lint.SeverityWarning,
		},
		{
			ID:               ruleIDPolicyAmbiguousSelector,
			Module:           serviceRulesModuleID,
			Scope:            serviceRulesScope,
			ScopeDescription: serviceRulesScopeDescription,
			Code:             codePolicyAmbiguousSelector,
			Message:          "ambiguous code selector",
			Description: "Reports a code selector that resolves to multiple " +
				"registered rules in soft unknown selector mode.",
			DefaultSeverity: lint.SeverityWarning,
		},
		{
			ID:               ruleIDPolicyShadowedEntry,
			Module:           serviceRulesModuleID,
			Scope:            serviceRulesScope,
			ScopeDescription: serviceRulesScopeDescription,
			Code:             codePolicyShadowedEntry,
			Message:          "policy entry is shadowed by later entry",
			Description: "Reports a rule entry that is fully shadowed by a " +
				"later entry with the same selector and matcher.",
			DefaultSeverity: lint.SeverityNotice,
		},
		{
			ID:               ruleIDPolicyNeverMatchesOverride,
			Module:           serviceRulesModuleID,
			Scope:            serviceRulesScope,
			ScopeDescription: serviceRulesScopeDescription,
			Code:             codePolicyNeverMatchesOverride,
			Message:          "override matcher never matches sampled paths",
			Description: "Reports an override matcher that does not match a " +
				"small built-in sample set of common project paths.",
			DefaultSeverity: lint.SeverityNotice,
		},
	}
)

// ServiceModuleSpec returns built-in lintkit service diagnostics module metadata.
func ServiceModuleSpec() lint.ModuleSpec {
	return lint.ModuleSpec{
		ID:          serviceRulesModuleID,
		Name:        "lintkit policy diagnostics",
		Description: "Built-in diagnostics for lint policy configuration quality.",
	}
}

// ServiceRuleSpecs returns built-in lintkit service diagnostics rule metadata.
func ServiceRuleSpecs() []lint.RuleSpec {
	out := make([]lint.RuleSpec, len(serviceRuleSpecs))
	copy(out, serviceRuleSpecs)
	return out
}

// AppendServiceRules appends built-in service module and rules to snapshot.
//
// Existing service module/rules are not duplicated.
func AppendServiceRules(snapshot lint.RegistrySnapshot) lint.RegistrySnapshot {
	module := ServiceModuleSpec()
	rules := ServiceRuleSpecs()

	hasModule := false
	for index := range snapshot.Modules {
		if strings.TrimSpace(snapshot.Modules[index].ID) ==
			strings.TrimSpace(module.ID) {
			hasModule = true
			break
		}
	}

	if !hasModule {
		snapshot.Modules = append(snapshot.Modules, module)
	}

	knownRules := make(map[string]struct{}, len(snapshot.Rules))
	for index := range snapshot.Rules {
		id := strings.TrimSpace(snapshot.Rules[index].ID)
		if id != "" {
			knownRules[id] = struct{}{}
		}
	}

	for index := range rules {
		id := strings.TrimSpace(rules[index].ID)
		if _, exists := knownRules[id]; exists {
			continue
		}

		knownRules[id] = struct{}{}
		snapshot.Rules = append(snapshot.Rules, rules[index])
	}

	return snapshot
}

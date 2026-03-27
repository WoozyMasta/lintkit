// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import "github.com/woozymasta/lintkit/lint"

// ruleSettingsChain stores selector settings in deterministic apply order.
type ruleSettingsChain struct {
	// All stores "*" selector settings.
	All RuleSettings

	// Module stores "<module>.*" selector settings.
	Module RuleSettings

	// Scope stores "<module>.<scope>.*" selector settings.
	Scope RuleSettings

	// Code stores "<CODE>" selector settings.
	Code RuleSettings

	// Rule stores exact "<rule-id>" selector settings.
	Rule RuleSettings

	// HasAll reports whether All is set.
	HasAll bool

	// HasModule reports whether Module is set.
	HasModule bool

	// HasScope reports whether Scope is set.
	HasScope bool

	// HasCode reports whether Code is set.
	HasCode bool

	// HasRule reports whether Rule is set.
	HasRule bool
}

// applyRuleSettingsChain applies one selector chain in fixed precedence order.
func applyRuleSettingsChain(
	decision RuleDecision,
	chain ruleSettingsChain,
) RuleDecision {
	if chain.HasAll {
		decision = applyRuleSettings(decision, chain.All)
	}

	if chain.HasModule {
		decision = applyRuleSettings(decision, chain.Module)
	}

	if chain.HasScope {
		decision = applyRuleSettings(decision, chain.Scope)
	}

	if chain.HasCode {
		decision = applyRuleSettings(decision, chain.Code)
	}

	if chain.HasRule {
		decision = applyRuleSettings(decision, chain.Rule)
	}

	return decision
}

// resolveRuleSettingsChainFromMap resolves selector settings from raw selector map.
func resolveRuleSettingsChainFromMap(
	rule lint.RuleSpec,
	settings map[string]RuleSettings,
) ruleSettingsChain {
	if len(settings) == 0 {
		return ruleSettingsChain{}
	}

	chain := ruleSettingsChain{}

	if all, ok := settings[RuleSelectorAll]; ok {
		chain.All = all
		chain.HasAll = true
	}

	if rule.Module != "" {
		if moduleRules, ok := settings[rule.Module+".*"]; ok {
			chain.Module = moduleRules
			chain.HasModule = true
		}
	}

	if rule.Module != "" && rule.Scope != "" {
		scopeSelector := rule.Module + "." + rule.Scope + ".*"
		if scopeRules, ok := settings[scopeSelector]; ok {
			chain.Scope = scopeRules
			chain.HasScope = true
		}
	}

	if code := rule.Code; code != "" {
		if codeRules, ok := settings[code]; ok {
			chain.Code = codeRules
			chain.HasCode = true
		}
	}

	if exact, ok := settings[rule.ID]; ok {
		chain.Rule = exact
		chain.HasRule = true
	}

	return chain
}

// resolveRuleSettingsChainFromCompiled resolves selector settings from preindexed maps.
func resolveRuleSettingsChainFromCompiled(
	rule lint.RuleSpec,
	settings compiledRuleSettings,
) ruleSettingsChain {
	chain := ruleSettingsChain{}

	if settings.All != nil {
		chain.All = *settings.All
		chain.HasAll = true
	}

	if rule.Module != "" && settings.ByModule != nil {
		if item, ok := settings.ByModule[rule.Module]; ok {
			chain.Module = item
			chain.HasModule = true
		}
	}

	if rule.Module != "" && rule.Scope != "" && settings.ByScope != nil {
		scopeKey := rule.Module + "." + rule.Scope
		if item, ok := settings.ByScope[scopeKey]; ok {
			chain.Scope = item
			chain.HasScope = true
		}
	}

	if code := rule.Code; code != "" && settings.ByCode != nil {
		if item, ok := settings.ByCode[code]; ok {
			chain.Code = item
			chain.HasCode = true
		}
	}

	if settings.ByRule != nil {
		if item, ok := settings.ByRule[rule.ID]; ok {
			chain.Rule = item
			chain.HasRule = true
		}
	}

	return chain
}

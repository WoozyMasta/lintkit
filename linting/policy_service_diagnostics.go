// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// policyDiagnosticKey stores emitted policy service diagnostic fingerprints.
	policyDiagnosticKey = "lintkit.policy.emitted"

	// ruleIDPolicyDeprecatedSelector reports deprecated selector usage.
	ruleIDPolicyDeprecatedSelector = "lintkit.policy.deprecated-selector"

	// ruleIDPolicyUnknownSelector reports unknown selector in soft mode.
	ruleIDPolicyUnknownSelector = "lintkit.policy.unknown-selector"

	// ruleIDPolicyAmbiguousSelector reports ambiguous code selector usage.
	ruleIDPolicyAmbiguousSelector = "lintkit.policy.ambiguous-code-selector"

	// ruleIDPolicyShadowedEntry reports shadowed ordered entry.
	ruleIDPolicyShadowedEntry = "lintkit.policy.shadowed-entry"

	// ruleIDPolicyNeverMatchesOverride reports never-matching override matcher.
	ruleIDPolicyNeverMatchesOverride = "lintkit.policy.override-never-matches"
)

const (
	// codePolicyDeprecatedSelector is deprecated selector service code.
	codePolicyDeprecatedSelector = "LINTKIT1001"

	// codePolicyUnknownSelector is unknown selector service code.
	codePolicyUnknownSelector = "LINTKIT1002"

	// codePolicyAmbiguousSelector is ambiguous selector service code.
	codePolicyAmbiguousSelector = "LINTKIT1003"

	// codePolicyShadowedEntry is shadowed entry service code.
	codePolicyShadowedEntry = "LINTKIT1004"

	// codePolicyNeverMatchesOverride is never-matching override service code.
	codePolicyNeverMatchesOverride = "LINTKIT1005"
)

// policySelectorSource stores selector text with stable source position.
type policySelectorSource struct {
	// Selector is one normalized selector token.
	Selector string

	// Source is human-readable selector source label.
	Source string
}

// policySelectorStats stores selector match summary.
type policySelectorStats struct {
	// Deprecated stores matched deprecated rules.
	Deprecated []lint.RuleSpec

	// Count stores matched rule count.
	Count int
}

// serviceDiagnosticsEnabled reports whether service diagnostics are enabled.
func serviceDiagnosticsEnabled(options *RunOptions) bool {
	if options == nil || options.EnableServiceDiagnostics == nil {
		return true
	}

	return *options.EnableServiceDiagnostics
}

// collectPolicyServiceDiagnostics builds one-time policy service diagnostics.
func collectPolicyServiceDiagnostics(
	policy *RunPolicy,
	registered []lint.RuleSpec,
) []lint.Diagnostic {
	if policy == nil {
		return nil
	}

	selectors := collectPolicySelectors(policy)
	if len(selectors) == 0 && len(policy.RuleEntries) == 0 &&
		len(policy.Overrides) == 0 {
		return nil
	}

	out := make([]lint.Diagnostic, 0, 8)
	deprecatedSeen := make(map[string]struct{}, 4)
	shadowedSeen := make(map[int]struct{}, 2)
	neverMatchSeen := make(map[int]struct{}, 2)

	for index := range selectors {
		selector := selectors[index]
		parsed, err := parseRuleSelector(selector.Selector)
		if err != nil {
			continue
		}

		stats := evaluatePolicySelector(parsed, registered)

		if !policy.Strict {
			if parsed.kind == selectorKindAll {
				continue
			}

			if stats.Count == 0 {
				out = append(out, lint.Diagnostic{
					RuleID:   ruleIDPolicyUnknownSelector,
					Code:     codePolicyUnknownSelector,
					Severity: lint.SeverityWarning,
					Message: fmt.Sprintf(
						"unknown policy selector %q at %s",
						selector.Selector,
						selector.Source,
					),
				})
				continue
			}

			if parsed.kind == selectorKindCode && stats.Count > 1 {
				out = append(out, lint.Diagnostic{
					RuleID:   ruleIDPolicyAmbiguousSelector,
					Code:     codePolicyAmbiguousSelector,
					Severity: lint.SeverityWarning,
					Message: fmt.Sprintf(
						"ambiguous code selector %q at %s",
						selector.Selector,
						selector.Source,
					),
				})
			}
		}

		for deprecatedIndex := range stats.Deprecated {
			rule := stats.Deprecated[deprecatedIndex]
			if _, exists := deprecatedSeen[rule.ID]; exists {
				continue
			}

			deprecatedSeen[rule.ID] = struct{}{}
			out = append(out, lint.Diagnostic{
				RuleID:   ruleIDPolicyDeprecatedSelector,
				Code:     codePolicyDeprecatedSelector,
				Severity: lint.SeverityWarning,
				Message: fmt.Sprintf(
					"policy selector %q references deprecated rule %q",
					selector.Selector,
					rule.ID,
				),
			})
		}
	}

	for index := range policy.RuleEntries {
		if policy.RuleEntries[index].Matcher != nil {
			continue
		}

		selector := strings.TrimSpace(policy.RuleEntries[index].Selector)
		if selector == "" {
			continue
		}

		for next := index + 1; next < len(policy.RuleEntries); next++ {
			if policy.RuleEntries[next].Matcher != nil {
				continue
			}

			if selector != strings.TrimSpace(policy.RuleEntries[next].Selector) {
				continue
			}

			if _, exists := shadowedSeen[index]; exists {
				break
			}

			shadowedSeen[index] = struct{}{}
			out = append(out, lint.Diagnostic{
				RuleID:   ruleIDPolicyShadowedEntry,
				Code:     codePolicyShadowedEntry,
				Severity: lint.SeverityNotice,
				Message: fmt.Sprintf(
					"rule_entries[%d] selector %q is shadowed by later entry",
					index,
					selector,
				),
			})
			break
		}
	}

	for index := range policy.Overrides {
		matcher := policy.Overrides[index].Matcher
		if matcher == nil || !matcherSeemsNeverMatch(matcher) {
			continue
		}

		if _, exists := neverMatchSeen[index]; exists {
			continue
		}

		neverMatchSeen[index] = struct{}{}
		name := strings.TrimSpace(policy.Overrides[index].Name)
		label := "override[" + strconv.Itoa(index) + "]"
		if name != "" {
			label = label + " (" + name + ")"
		}

		out = append(out, lint.Diagnostic{
			RuleID:   ruleIDPolicyNeverMatchesOverride,
			Code:     codePolicyNeverMatchesOverride,
			Severity: lint.SeverityNotice,
			Message:  label + " matcher does not match any sampled paths",
		})
	}

	slices.SortStableFunc(out, func(left lint.Diagnostic, right lint.Diagnostic) int {
		if left.RuleID != right.RuleID {
			if left.RuleID < right.RuleID {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Message, right.Message)
	})

	return out
}

// collectPolicySelectors extracts all policy selectors with stable source labels.
func collectPolicySelectors(policy *RunPolicy) []policySelectorSource {
	if policy == nil {
		return nil
	}

	out := make([]policySelectorSource, 0, 8)
	if len(policy.Rules) > 0 {
		keys := make([]string, 0, len(policy.Rules))
		for key := range policy.Rules {
			keys = append(keys, key)
		}

		slices.Sort(keys)
		for index := range keys {
			out = append(out, policySelectorSource{
				Selector: strings.TrimSpace(keys[index]),
				Source:   "rules[" + keys[index] + "]",
			})
		}
	}

	for index := range policy.RuleEntries {
		selector := strings.TrimSpace(policy.RuleEntries[index].Selector)
		if selector == "" {
			continue
		}

		out = append(out, policySelectorSource{
			Selector: selector,
			Source:   "rule_entries[" + strconv.Itoa(index) + "]",
		})
	}

	for index := range policy.Overrides {
		if len(policy.Overrides[index].Rules) == 0 {
			continue
		}

		keys := make([]string, 0, len(policy.Overrides[index].Rules))
		for key := range policy.Overrides[index].Rules {
			keys = append(keys, key)
		}

		slices.Sort(keys)
		for keyIndex := range keys {
			out = append(out, policySelectorSource{
				Selector: strings.TrimSpace(keys[keyIndex]),
				Source: "overrides[" + strconv.Itoa(index) + "].rules[" +
					keys[keyIndex] + "]",
			})
		}
	}

	return out
}

// evaluatePolicySelector returns matched and deprecated rule stats for selector.
func evaluatePolicySelector(
	selector parsedRuleSelector,
	registered []lint.RuleSpec,
) policySelectorStats {
	stats := policySelectorStats{}
	if len(registered) == 0 {
		return stats
	}

	for index := range registered {
		if !selectorMatchesRule(selector, registered[index]) {
			continue
		}

		stats.Count++
		if registered[index].Deprecated {
			stats.Deprecated = append(stats.Deprecated, registered[index])
		}
	}

	return stats
}

// matcherSeemsNeverMatch reports whether matcher misses common path samples.
func matcherSeemsNeverMatch(matcher PathMatcher) bool {
	if matcher == nil {
		return false
	}

	samples := []struct {
		Path  string
		IsDir bool
	}{
		{Path: "", IsDir: false},
		{Path: ".", IsDir: true},
		{Path: "src/main.cfg", IsDir: false},
		{Path: "src", IsDir: true},
	}

	for index := range samples {
		if matcher.Match(samples[index].Path, samples[index].IsDir) {
			return false
		}
	}

	return true
}

// appendUniquePolicyDiagnostics appends policy diagnostics with value-key dedupe.
func appendUniquePolicyDiagnostics(
	run *lint.RunContext,
	target *[]lint.Diagnostic,
	diagnostics []lint.Diagnostic,
) {
	if run == nil || target == nil || len(diagnostics) == 0 {
		return
	}

	if run.Values == nil {
		run.Values = make(map[string]any)
	}

	seen, _ := run.Values[policyDiagnosticKey].(map[string]struct{})
	if seen == nil {
		seen = make(map[string]struct{}, len(diagnostics))
		run.Values[policyDiagnosticKey] = seen
	}

	for index := range diagnostics {
		fingerprint := diagnostics[index].RuleID + "\n" + diagnostics[index].Message
		if _, exists := seen[fingerprint]; exists {
			continue
		}

		seen[fingerprint] = struct{}{}
		*target = append(*target, diagnostics[index])
	}
}

// cloneDiagnostics returns shallow copy of diagnostics slice.
func cloneDiagnostics(items []lint.Diagnostic) []lint.Diagnostic {
	if len(items) == 0 {
		return nil
	}

	out := make([]lint.Diagnostic, len(items))
	copy(out, items)
	return out
}

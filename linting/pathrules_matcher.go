// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import "github.com/woozymasta/pathrules"

// PathRulesCompilerOptions configures pathrules-backed matcher compiler.
type PathRulesCompilerOptions struct {
	// CaseInsensitive enables ASCII case-insensitive matching.
	CaseInsensitive bool `json:"case_insensitive,omitempty" yaml:"case_insensitive,omitempty" jsonschema:"default=false"`
}

// PathRulesCompiler returns PatternMatcherCompiler backed by pathrules.Matcher.
func PathRulesCompiler(options PathRulesCompilerOptions) PatternMatcherCompiler {
	return func(patterns []string) (PathMatcher, error) {
		return CompilePathRulesMatcher(patterns, options)
	}
}

// CompilePathRulesMatcher compiles one path-matching policy from pattern list.
//
// Patterns are interpreted as "match when included":
// each pattern is compiled as include rule and matcher default decision is exclude.
func CompilePathRulesMatcher(
	patterns []string,
	options PathRulesCompilerOptions,
) (PathMatcher, error) {
	normalizedPatterns := normalizePatternList(patterns)
	if len(normalizedPatterns) == 0 {
		return PathMatcherFunc(func(string, bool) bool {
			return false
		}), nil
	}

	rules := make([]pathrules.Rule, 0, len(normalizedPatterns))
	for index := range normalizedPatterns {
		rules = append(rules, pathrules.Rule{
			Pattern: normalizedPatterns[index],
			Action:  pathrules.ActionInclude,
		})
	}

	matcher, err := pathrules.NewMatcher(rules, pathrules.MatcherOptions{
		CaseInsensitive: options.CaseInsensitive,
		DefaultAction:   pathrules.ActionExclude,
	})
	if err != nil {
		return nil, err
	}

	return PathMatcherFunc(func(path string, isDir bool) bool {
		return matcher.Included(path, isDir)
	}), nil
}

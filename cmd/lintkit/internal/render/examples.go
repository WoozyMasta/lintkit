// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

// Package render renders lint registry documentation outputs.
package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/yamlutil"
	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

// sampleSelectors stores one deterministic set of real rule selector examples.
type sampleSelectors struct {
	// Module stores "<module>.*" selector.
	Module string

	// Scope stores "<module>.<scope>.*" selector.
	Scope string

	// Code stores one public code selector.
	Code string

	// RuleID stores one exact rule-id selector.
	RuleID string
}

// renderPolicyExampleContent renders policy example content in selected format.
func renderPolicyExampleContent(
	snapshot lint.RegistrySnapshot,
	format string,
) (string, error) {
	switch format {
	case "json":
		content, err := renderPolicyExampleJSON(snapshot)
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(content), nil
	case "yaml":
		content, err := renderPolicyExampleYAML(snapshot)
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(content), nil
	default:
		return "", fmt.Errorf("unsupported example format %q", format)
	}
}

// renderPolicyExampleJSON renders compact JSON policy example for markdown block.
func renderPolicyExampleJSON(snapshot lint.RegistrySnapshot) (string, error) {
	config := sampleRunPolicyConfig(snapshot)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal policy example json: %w", err)
	}

	return string(data), nil
}

// renderPolicyExampleYAML renders compact YAML policy example for markdown block.
func renderPolicyExampleYAML(snapshot lint.RegistrySnapshot) (string, error) {
	config := sampleRunPolicyConfig(snapshot)
	data, err := yamlutil.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal policy example yaml: %w", err)
	}

	return string(data), nil
}

// sampleRunPolicyConfig builds deterministic policy config example.
func sampleRunPolicyConfig(snapshot lint.RegistrySnapshot) linting.RunPolicyConfig {
	selectors := sampleRuleSelectors(snapshot)
	if selectors.Module == "" && selectors.Code == "" && selectors.RuleID == "" {
		return linting.RunPolicyConfig{}
	}

	rules := make([]linting.RunPolicyRuleConfig, 0, 5)
	rules = append(rules, linting.RunPolicyRuleConfig{
		Rule:     "*",
		Severity: lint.SeverityWarning,
	})

	if selectors.Module != "" {
		rules = append(rules, linting.RunPolicyRuleConfig{
			Rule:     selectors.Module,
			Severity: lint.SeverityError,
		})
	}

	if selectors.Scope != "" {
		rules = append(rules, linting.RunPolicyRuleConfig{
			Rule:     selectors.Scope,
			Severity: lint.SeverityNotice,
		})
	}

	if selectors.Code != "" {
		overlay := linting.BuildPolicyOverlay(linting.OverlayInput{
			DisableRules: []string{selectors.Code},
		})
		rules = append(rules, overlay.Rules...)
	}

	if selectors.RuleID != "" {
		rules = append(rules, linting.RunPolicyRuleConfig{
			Rule:     selectors.RuleID,
			Severity: lint.SeverityInfo,
		})
	}

	config := linting.RunPolicyConfig{
		Exclude: []string{"**/vendor/**", "**/*.generated.*"},
		FailOn:  lint.SeverityError,
		Rules:   rules,
	}

	if selectors.Module != "" {
		config.Rules = append(config.Rules, linting.RunPolicyRuleConfig{
			Rule:     selectors.Module,
			Exclude:  []string{"**/generated/**"},
			Severity: lint.SeverityNotice,
		})
	}

	return config
}

// sampleRuleSelectors picks deterministic real module/code/rule-id selectors.
func sampleRuleSelectors(snapshot lint.RegistrySnapshot) sampleSelectors {
	specs := make([]lint.RuleSpec, len(snapshot.Rules))
	copy(specs, snapshot.Rules)
	sortRuleSpecs(specs)

	out := sampleSelectors{}

	for index := range specs {
		if out.RuleID == "" {
			ruleID := strings.TrimSpace(specs[index].ID)
			if ruleID != "" {
				out.RuleID = ruleID
			}
		}

		if out.Module == "" {
			module := strings.TrimSpace(specs[index].Module)
			if module != "" {
				out.Module = module + ".*"
			}
		}

		if out.Scope == "" {
			module := strings.TrimSpace(specs[index].Module)
			scope := strings.TrimSpace(specs[index].Scope)
			if module != "" && scope != "" {
				out.Scope = module + "." + scope + ".*"
			}
		}

		codeText := strings.TrimSpace(specs[index].Code)
		if codeText == "" {
			continue
		}

		if _, ok := lint.ParsePublicCode(codeText); !ok {
			continue
		}

		if out.Code == "" {
			out.Code = codeText
		}
	}

	return out
}

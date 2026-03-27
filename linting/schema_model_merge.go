// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

// Merge returns merged config where argument values override receiver values.
//
// Merge strategy:
//   - Exclude: append base+overlay with normalization and dedupe.
//   - Rules: append base rules then overlay rules with normalization.
//   - SoftUnknownSelectors: logical OR to avoid unexpected strict failures.
//   - FailOn: overlay non-empty value overrides base value.
func (config RunPolicyConfig) Merge(
	overlay RunPolicyConfig,
) (RunPolicyConfig, error) {
	return MergeRunPolicyConfig(config, overlay)
}

// MergeRunPolicyConfig merges two policy configs (base then overlay).
func MergeRunPolicyConfig(
	base RunPolicyConfig,
	overlay RunPolicyConfig,
) (RunPolicyConfig, error) {
	if err := validateRunPolicyConfigFailOn(base.FailOn); err != nil {
		return RunPolicyConfig{}, err
	}

	if err := validateRunPolicyConfigFailOn(overlay.FailOn); err != nil {
		return RunPolicyConfig{}, err
	}

	failOn := base.FailOn
	if overlay.FailOn != "" {
		failOn = overlay.FailOn
	}

	merged := RunPolicyConfig{
		Exclude:              mergePatternLists(base.Exclude, overlay.Exclude),
		SoftUnknownSelectors: base.SoftUnknownSelectors || overlay.SoftUnknownSelectors,
		FailOn:               failOn,
		Rules:                make([]RunPolicyRuleConfig, 0, len(base.Rules)+len(overlay.Rules)),
	}

	if err := appendMergedRuleConfigs(&merged.Rules, base.Rules); err != nil {
		return RunPolicyConfig{}, err
	}

	if err := appendMergedRuleConfigs(&merged.Rules, overlay.Rules); err != nil {
		return RunPolicyConfig{}, err
	}

	if len(merged.Rules) == 0 {
		merged.Rules = nil
	}

	return merged, nil
}

// appendMergedRuleConfigs appends normalized rule configs to target.
func appendMergedRuleConfigs(
	target *[]RunPolicyRuleConfig,
	source []RunPolicyRuleConfig,
) error {
	if len(source) == 0 {
		return nil
	}

	for index := range source {
		rule := source[index].Rule
		settings := RuleSettings{
			Options:  source[index].Options,
			Enabled:  source[index].Enabled,
			Severity: source[index].Severity,
		}

		normalizedRules, err := NormalizeRuleSettingsMap(
			map[string]RuleSettings{rule: settings},
		)
		if err != nil {
			return err
		}

		normalizedSelector := ""
		normalizedSettings := RuleSettings{}
		for key, value := range normalizedRules {
			normalizedSelector = key
			normalizedSettings = value
			break
		}

		*target = append(*target, RunPolicyRuleConfig{
			Rule:     normalizedSelector,
			Exclude:  normalizePatternList(source[index].Exclude),
			Options:  normalizedSettings.Options,
			Enabled:  normalizedSettings.Enabled,
			Severity: normalizedSettings.Severity,
		})
	}

	return nil
}

// mergePatternLists appends pattern lists and normalizes result.
func mergePatternLists(base []string, overlay []string) []string {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}

	combined := make([]string, 0, len(base)+len(overlay))
	combined = append(combined, base...)
	combined = append(combined, overlay...)

	return normalizePatternList(combined)
}

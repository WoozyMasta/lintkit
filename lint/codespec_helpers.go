// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "strings"

// NewCodeSpec builds normalized lint code metadata row.
func NewCodeSpec(spec CodeSpec) CodeSpec {
	out := spec
	out.Stage = Stage(strings.TrimSpace(string(out.Stage)))
	out.Message = strings.TrimSpace(out.Message)

	if out.Severity == "" {
		out.Severity = SeverityWarning
	}

	if out.Rule != nil {
		copyRule := *out.Rule
		out.Rule = &copyRule
	}

	return out
}

// ErrorCodeSpec builds code metadata with default error severity.
func ErrorCodeSpec(
	code Code,
	stage Stage,
	message string,
) CodeSpec {
	return NewCodeSpec(CodeSpec{
		Code:     code,
		Stage:    stage,
		Severity: SeverityError,
		Message:  message,
	})
}

// WarningCodeSpec builds code metadata with default warning severity.
func WarningCodeSpec(
	code Code,
	stage Stage,
	message string,
) CodeSpec {
	return NewCodeSpec(CodeSpec{
		Code:     code,
		Stage:    stage,
		Severity: SeverityWarning,
		Message:  message,
	})
}

// InfoCodeSpec builds code metadata with default info severity.
func InfoCodeSpec(
	code Code,
	stage Stage,
	message string,
) CodeSpec {
	return NewCodeSpec(CodeSpec{
		Code:     code,
		Stage:    stage,
		Severity: SeverityInfo,
		Message:  message,
	})
}

// NoticeCodeSpec builds code metadata with default notice severity.
func NoticeCodeSpec(
	code Code,
	stage Stage,
	message string,
) CodeSpec {
	return NewCodeSpec(CodeSpec{
		Code:     code,
		Stage:    stage,
		Severity: SeverityNotice,
		Message:  message,
	})
}

// WithCodeEnabled returns spec copy with default enabled state override.
func WithCodeEnabled(spec CodeSpec, enabled bool) CodeSpec {
	spec.Enabled = boolPtr(enabled)
	return spec
}

// WithCodeRule returns spec copy with explicit rule metadata override payload.
func WithCodeRule(spec CodeSpec, rule CodeRuleOverride) CodeSpec {
	if spec.Rule == nil {
		copyRule := rule
		spec.Rule = &copyRule
		return spec
	}

	spec.Rule = mergeCodeRuleFieldOverride(spec.Rule, &rule)
	return spec
}

// WithCodeOptions returns spec copy with optional options payload.
func WithCodeOptions(spec CodeSpec, options any) CodeSpec {
	spec.Options = options
	return spec
}

// WithCodeSeverity returns spec copy with explicit default severity.
func WithCodeSeverity(spec CodeSpec, severity Severity) CodeSpec {
	spec.Severity = severity
	return spec
}

// boolPtr allocates one bool pointer value.
func boolPtr(value bool) *bool {
	return &value
}

// mergeCodeRuleFieldOverride merges non-zero override fields into code-rule payload.
func mergeCodeRuleFieldOverride(
	base *CodeRuleOverride,
	override *CodeRuleOverride,
) *CodeRuleOverride {
	if base == nil && override == nil {
		return nil
	}

	if base == nil {
		copyOverride := *override
		return &copyOverride
	}

	if override == nil {
		copyBase := *base
		return &copyBase
	}

	out := *base
	if strings.TrimSpace(override.ID) != "" {
		out.ID = strings.TrimSpace(override.ID)
	}

	if strings.TrimSpace(override.Scope) != "" {
		out.Scope = strings.TrimSpace(override.Scope)
	}

	if strings.TrimSpace(override.ScopeDescription) != "" {
		out.ScopeDescription = strings.TrimSpace(override.ScopeDescription)
	}

	if strings.TrimSpace(override.Code) != "" {
		out.Code = strings.TrimSpace(override.Code)
	}

	if len(override.FileKinds) > 0 {
		out.FileKinds = NormalizeFileKinds(override.FileKinds)
	}

	if override.Deprecated {
		out.Deprecated = true
	}

	return &out
}

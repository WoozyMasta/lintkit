// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"

	"github.com/woozymasta/lintkit/lint"
)

// DiagnosticSummary stores aggregated diagnostic counters.
type DiagnosticSummary struct {
	// ErrorCount stores number of "error" diagnostics.
	ErrorCount int `json:"error_count" yaml:"error_count"`

	// WarningCount stores number of "warning" diagnostics.
	WarningCount int `json:"warning_count" yaml:"warning_count"`

	// InfoCount stores number of "info" diagnostics.
	InfoCount int `json:"info_count" yaml:"info_count"`

	// NoticeCount stores number of "notice" diagnostics.
	NoticeCount int `json:"notice_count" yaml:"notice_count"`

	// UnknownCount stores number of diagnostics with unsupported severity.
	UnknownCount int `json:"unknown_count" yaml:"unknown_count"`

	// TotalCount stores total number of diagnostics.
	TotalCount int `json:"total_count" yaml:"total_count"`

	// RuleErrorCount stores number of runtime rule execution errors.
	RuleErrorCount int `json:"rule_error_count" yaml:"rule_error_count"`
}

// Summary returns aggregated diagnostic counters.
func (result RunResult) Summary() DiagnosticSummary {
	summary := DiagnosticSummary{
		RuleErrorCount: len(result.RuleErrors),
	}

	for index := range result.Diagnostics {
		switch result.Diagnostics[index].Severity {
		case lint.SeverityError:
			summary.ErrorCount++
		case lint.SeverityWarning:
			summary.WarningCount++
		case lint.SeverityInfo:
			summary.InfoCount++
		case lint.SeverityNotice:
			summary.NoticeCount++
		default:
			summary.UnknownCount++
		}
	}

	summary.TotalCount = len(result.Diagnostics)
	return summary
}

// HighestSeverity returns highest severity currently present in summary.
func (summary DiagnosticSummary) HighestSeverity() lint.Severity {
	if summary.ErrorCount > 0 {
		return lint.SeverityError
	}

	if summary.WarningCount > 0 {
		return lint.SeverityWarning
	}

	if summary.InfoCount > 0 {
		return lint.SeverityInfo
	}

	if summary.NoticeCount > 0 {
		return lint.SeverityNotice
	}

	return ""
}

// ShouldFail reports whether run result exceeds fail threshold.
//
// Empty failOn is normalized to "error".
// When failOnRuleError is true, any runtime rule error fails the result.
func (result RunResult) ShouldFail(
	failOn lint.Severity,
	failOnRuleError bool,
) (bool, error) {
	threshold, err := normalizeFailOnSeverity(failOn)
	if err != nil {
		return false, err
	}

	summary := result.Summary()
	if failOnRuleError && summary.RuleErrorCount > 0 {
		return true, nil
	}

	if summary.TotalCount == 0 {
		return false, nil
	}

	highest := summary.HighestSeverity()
	if highest == "" {
		return false, nil
	}

	return lint.SeverityRank(highest) >= lint.SeverityRank(threshold), nil
}

// normalizeFailOnSeverity validates threshold severity and applies defaults.
func normalizeFailOnSeverity(value lint.Severity) (lint.Severity, error) {
	if value == "" {
		return lint.SeverityError, nil
	}

	if !isSupportedSeverity(value) {
		return "", fmt.Errorf("%w: %q", ErrInvalidFailSeverity, value)
	}

	return value, nil
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "fmt"

// DiagnosticsError stores diagnostics that matched fail threshold.
type DiagnosticsError struct {
	// Diagnostics stores diagnostics at or above threshold severity.
	Diagnostics []Diagnostic

	// HighestSeverity stores highest matched severity level.
	HighestSeverity Severity

	// Threshold stores normalized fail threshold severity.
	Threshold Severity
}

// Error returns compact threshold violation text.
func (item DiagnosticsError) Error() string {
	if len(item.Diagnostics) == 0 {
		return ""
	}

	highest := item.HighestSeverity
	if highest == "" {
		highest = highestDiagnosticsSeverity(item.Diagnostics)
	}

	return fmt.Sprintf(
		"diagnostics threshold %q matched %d item(s), highest=%q",
		item.Threshold,
		len(item.Diagnostics),
		highest,
	)
}

// ErrorFromDiagnostics returns typed error when diagnostics reach threshold.
//
// Empty minSeverity defaults to SeverityError.
func ErrorFromDiagnostics(
	diagnostics []Diagnostic,
	minSeverity Severity,
) error {
	threshold, err := normalizeDiagnosticThreshold(minSeverity)
	if err != nil {
		return err
	}

	matched := diagnosticsAtOrAbove(diagnostics, threshold)
	if len(matched) == 0 {
		return nil
	}

	return &DiagnosticsError{
		Diagnostics:     matched,
		Threshold:       threshold,
		HighestSeverity: highestDiagnosticsSeverity(matched),
	}
}

// normalizeDiagnosticThreshold normalizes fail threshold and validates severity.
func normalizeDiagnosticThreshold(value Severity) (Severity, error) {
	if value == "" {
		return SeverityError, nil
	}

	if !IsSupportedSeverity(value) {
		return "", fmt.Errorf("%w: %q", ErrInvalidDiagnosticSeverity, value)
	}

	return value, nil
}

// diagnosticsAtOrAbove returns diagnostics at or above threshold severity.
func diagnosticsAtOrAbove(
	diagnostics []Diagnostic,
	threshold Severity,
) []Diagnostic {
	if len(diagnostics) == 0 {
		return nil
	}

	out := make([]Diagnostic, 0, len(diagnostics))
	thresholdRank := SeverityRank(threshold)
	for index := range diagnostics {
		if SeverityRank(diagnostics[index].Severity) < thresholdRank {
			continue
		}

		out = append(out, diagnostics[index])
	}

	return out
}

// highestDiagnosticsSeverity returns highest supported severity in diagnostics.
func highestDiagnosticsSeverity(diagnostics []Diagnostic) Severity {
	if len(diagnostics) == 0 {
		return ""
	}

	highest := Severity("")
	highestRank := 0
	for index := range diagnostics {
		rank := SeverityRank(diagnostics[index].Severity)
		if rank <= highestRank {
			continue
		}

		highestRank = rank
		highest = diagnostics[index].Severity
	}

	return highest
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linttest

import (
	"fmt"
	"testing"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

// SortDiagnostics applies deterministic order for diagnostics slice in place.
func SortDiagnostics(items []lint.Diagnostic) {
	if len(items) < 2 {
		return
	}

	result := linting.RunResult{
		Diagnostics: items,
	}
	result.SortDiagnostics()
}

// SortedDiagnostics returns sorted copy of diagnostics slice.
func SortedDiagnostics(items []lint.Diagnostic) []lint.Diagnostic {
	out := make([]lint.Diagnostic, len(items))
	copy(out, items)
	SortDiagnostics(out)

	return out
}

// AssertDiagnosticsEqual sorts both diagnostic slices and fails test on mismatch.
func AssertDiagnosticsEqual(
	tb testing.TB,
	got []lint.Diagnostic,
	want []lint.Diagnostic,
) {
	tb.Helper()

	if err := ValidateDiagnosticsEqual(got, want); err != nil {
		tb.Fatal(err)
	}
}

// ValidateDiagnosticsEqual sorts and compares diagnostic slices.
func ValidateDiagnosticsEqual(
	got []lint.Diagnostic,
	want []lint.Diagnostic,
) error {
	gotSorted := SortedDiagnostics(got)
	wantSorted := SortedDiagnostics(want)
	if len(gotSorted) != len(wantSorted) {
		return fmt.Errorf(
			"diagnostics mismatch: len(got)=%d, len(want)=%d",
			len(gotSorted),
			len(wantSorted),
		)
	}

	for index := range wantSorted {
		if gotSorted[index] == wantSorted[index] {
			continue
		}

		return fmt.Errorf(
			"diagnostics mismatch at index %d:\n  got:  %s\n  want: %s",
			index,
			formatDiagnostic(gotSorted[index]),
			formatDiagnostic(wantSorted[index]),
		)
	}

	return nil
}

// formatDiagnostic returns compact one-line diagnostic representation.
func formatDiagnostic(item lint.Diagnostic) string {
	return fmt.Sprintf(
		"{rule_id=%q code=%q severity=%q path=%q start=%s end=%s message=%q}",
		item.RuleID,
		item.Code,
		item.Severity,
		item.Path,
		formatPosition(item.Start),
		formatPosition(item.End),
		item.Message,
	)
}

// formatPosition returns compact one-line source position representation.
func formatPosition(item lint.Position) string {
	return fmt.Sprintf(
		"{file=%q line=%d column=%d offset=%d}",
		item.File,
		item.Line,
		item.Column,
		item.Offset,
	)
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"errors"
	"testing"
)

func TestErrorFromDiagnosticsDefaultThreshold(t *testing.T) {
	t.Parallel()

	diagnostics := []Diagnostic{
		{
			RuleID:   "module_alpha.parse.a001",
			Severity: SeverityWarning,
			Message:  "warning",
		},
	}

	if err := ErrorFromDiagnostics(diagnostics, ""); err != nil {
		t.Fatalf("ErrorFromDiagnostics(default threshold) error=%v, want nil", err)
	}

	diagnostics = append(diagnostics, Diagnostic{
		RuleID:   "module_alpha.parse.a002",
		Severity: SeverityError,
		Message:  "error",
	})
	err := ErrorFromDiagnostics(diagnostics, "")
	if err == nil {
		t.Fatal("ErrorFromDiagnostics(default threshold) error=nil, want non-nil")
	}

	var diagnosticsErr *DiagnosticsError
	if !errors.As(err, &diagnosticsErr) {
		t.Fatalf("error=%T, want *DiagnosticsError", err)
	}

	if diagnosticsErr.Threshold != SeverityError {
		t.Fatalf(
			"DiagnosticsError.Threshold=%q, want %q",
			diagnosticsErr.Threshold,
			SeverityError,
		)
	}

	if len(diagnosticsErr.Diagnostics) != 1 {
		t.Fatalf(
			"len(DiagnosticsError.Diagnostics)=%d, want 1",
			len(diagnosticsErr.Diagnostics),
		)
	}
}

func TestErrorFromDiagnosticsWarningThreshold(t *testing.T) {
	t.Parallel()

	diagnostics := []Diagnostic{
		{
			RuleID:   "module_alpha.parse.a001",
			Severity: SeverityWarning,
			Message:  "warning",
		},
		{
			RuleID:   "module_alpha.parse.a002",
			Severity: SeverityError,
			Message:  "error",
		},
		{
			RuleID:   "module_alpha.parse.a003",
			Severity: SeverityNotice,
			Message:  "notice",
		},
	}

	err := ErrorFromDiagnostics(diagnostics, SeverityWarning)
	if err == nil {
		t.Fatal("ErrorFromDiagnostics(warning threshold) error=nil, want non-nil")
	}

	var diagnosticsErr *DiagnosticsError
	if !errors.As(err, &diagnosticsErr) {
		t.Fatalf("error=%T, want *DiagnosticsError", err)
	}

	if diagnosticsErr.HighestSeverity != SeverityError {
		t.Fatalf(
			"DiagnosticsError.HighestSeverity=%q, want %q",
			diagnosticsErr.HighestSeverity,
			SeverityError,
		)
	}

	if len(diagnosticsErr.Diagnostics) != 2 {
		t.Fatalf(
			"len(DiagnosticsError.Diagnostics)=%d, want 2",
			len(diagnosticsErr.Diagnostics),
		)
	}
}

func TestErrorFromDiagnosticsInvalidThreshold(t *testing.T) {
	t.Parallel()

	err := ErrorFromDiagnostics(nil, Severity("fatal"))
	if !errors.Is(err, ErrInvalidDiagnosticSeverity) {
		t.Fatalf(
			"ErrorFromDiagnostics(invalid threshold) error=%v, want ErrInvalidDiagnosticSeverity",
			err,
		)
	}
}

func TestErrorFromDiagnosticsReturnsDetachedCopy(t *testing.T) {
	t.Parallel()

	diagnostics := []Diagnostic{
		{
			RuleID:   "module_alpha.parse.a001",
			Severity: SeverityError,
			Message:  "before",
		},
	}

	err := ErrorFromDiagnostics(diagnostics, SeverityError)
	if err == nil {
		t.Fatal("ErrorFromDiagnostics(error threshold) error=nil, want non-nil")
	}

	var diagnosticsErr *DiagnosticsError
	if !errors.As(err, &diagnosticsErr) {
		t.Fatalf("error=%T, want *DiagnosticsError", err)
	}

	diagnostics[0].Message = "after"
	if diagnosticsErr.Diagnostics[0].Message != "before" {
		t.Fatalf(
			"DiagnosticsError.Diagnostics[0].Message=%q, want before",
			diagnosticsErr.Diagnostics[0].Message,
		)
	}
}

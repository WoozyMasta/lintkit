// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"github.com/woozymasta/lintkit/lint"
	"testing"
)

func TestRunResultSummary(t *testing.T) {
	t.Parallel()

	result := RunResult{
		Diagnostics: []lint.Diagnostic{
			{Severity: lint.SeverityError},
			{Severity: lint.SeverityWarning},
			{Severity: lint.SeverityInfo},
			{Severity: lint.SeverityNotice},
			{Severity: lint.Severity("unknown")},
		},
		RuleErrors: []RuleError{
			{RuleID: "module_alpha.R015", Cause: errors.New("runtime")},
		},
	}

	summary := result.Summary()
	if summary.ErrorCount != 1 {
		t.Fatalf("ErrorCount=%d, want 1", summary.ErrorCount)
	}

	if summary.WarningCount != 1 {
		t.Fatalf("WarningCount=%d, want 1", summary.WarningCount)
	}

	if summary.InfoCount != 1 {
		t.Fatalf("InfoCount=%d, want 1", summary.InfoCount)
	}

	if summary.NoticeCount != 1 {
		t.Fatalf("NoticeCount=%d, want 1", summary.NoticeCount)
	}

	if summary.UnknownCount != 1 {
		t.Fatalf("UnknownCount=%d, want 1", summary.UnknownCount)
	}

	if summary.TotalCount != 5 {
		t.Fatalf("TotalCount=%d, want 5", summary.TotalCount)
	}

	if summary.RuleErrorCount != 1 {
		t.Fatalf("RuleErrorCount=%d, want 1", summary.RuleErrorCount)
	}

	if summary.HighestSeverity() != lint.SeverityError {
		t.Fatalf("HighestSeverity=%q, want %q", summary.HighestSeverity(), lint.SeverityError)
	}
}

func TestRunResultShouldFail(t *testing.T) {
	t.Parallel()

	result := RunResult{
		Diagnostics: []lint.Diagnostic{
			{Severity: lint.SeverityWarning},
		},
	}

	fail, err := result.ShouldFail(lint.SeverityError, true)
	if err != nil {
		t.Fatalf("ShouldFail(error) error: %v", err)
	}

	if fail {
		t.Fatal("ShouldFail(error)=true, want false")
	}

	fail, err = result.ShouldFail(lint.SeverityWarning, true)
	if err != nil {
		t.Fatalf("ShouldFail(warning) error: %v", err)
	}

	if !fail {
		t.Fatal("ShouldFail(warning)=false, want true")
	}
}

func TestRunResultShouldFailRuleErrors(t *testing.T) {
	t.Parallel()

	result := RunResult{
		RuleErrors: []RuleError{
			{RuleID: "module_alpha.R015", Cause: errors.New("runtime")},
		},
	}

	fail, err := result.ShouldFail(lint.SeverityError, true)
	if err != nil {
		t.Fatalf("ShouldFail(rule-errors) error: %v", err)
	}

	if !fail {
		t.Fatal("ShouldFail(rule-errors)=false, want true")
	}

	fail, err = result.ShouldFail(lint.SeverityError, false)
	if err != nil {
		t.Fatalf("ShouldFail(rule-errors disabled) error: %v", err)
	}

	if fail {
		t.Fatal("ShouldFail(rule-errors disabled)=true, want false")
	}
}

func TestRunResultShouldFailValidation(t *testing.T) {
	t.Parallel()

	result := RunResult{}
	_, err := result.ShouldFail(lint.Severity("fatal"), true)
	if !errors.Is(err, ErrInvalidFailSeverity) {
		t.Fatalf("ShouldFail() error=%v, want ErrInvalidFailSeverity", err)
	}
}

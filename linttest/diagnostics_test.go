// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linttest

import (
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestSortDiagnostics(t *testing.T) {
	t.Parallel()

	items := []lint.Diagnostic{
		{
			RuleID:  "module.parse.b",
			Path:    "b.cpp",
			Start:   lint.Position{Line: 4, Column: 1},
			Message: "b",
		},
		{
			RuleID:  "module.parse.a",
			Path:    "a.cpp",
			Start:   lint.Position{Line: 10, Column: 1},
			Message: "a",
		},
		{
			RuleID:  "module.parse.a",
			Path:    "a.cpp",
			Start:   lint.Position{Line: 2, Column: 1},
			Message: "a",
		},
		{
			RuleID:  "module.parse.a",
			Path:    "a.cpp",
			Start:   lint.Position{Line: 2, Column: 1},
			Message: "z",
		},
	}

	SortDiagnostics(items)

	if items[0].Path != "a.cpp" || items[0].Start.Line != 2 || items[0].Message != "a" {
		t.Fatalf("items[0]=%v", items[0])
	}

	if items[1].Path != "a.cpp" || items[1].Start.Line != 2 || items[1].Message != "z" {
		t.Fatalf("items[1]=%v", items[1])
	}

	if items[2].Path != "a.cpp" || items[2].Start.Line != 10 {
		t.Fatalf("items[2]=%v", items[2])
	}

	if items[3].Path != "b.cpp" || items[3].Start.Line != 4 {
		t.Fatalf("items[3]=%v", items[3])
	}
}

func TestSortedDiagnosticsReturnsCopy(t *testing.T) {
	t.Parallel()

	items := []lint.Diagnostic{
		{RuleID: "module.rule.b", Path: "b.cpp"},
		{RuleID: "module.rule.a", Path: "a.cpp"},
	}

	sorted := SortedDiagnostics(items)

	if len(sorted) != 2 {
		t.Fatalf("len(sorted)=%d, want 2", len(sorted))
	}

	if sorted[0].Path != "a.cpp" {
		t.Fatalf("sorted[0].Path=%q, want a.cpp", sorted[0].Path)
	}

	if items[0].Path != "b.cpp" {
		t.Fatalf("items[0].Path=%q, want b.cpp", items[0].Path)
	}
}

func TestValidateDiagnosticsEqualIgnoresOrder(t *testing.T) {
	t.Parallel()

	got := []lint.Diagnostic{
		{
			RuleID:  "module.rule.b",
			Path:    "b.cpp",
			Message: "b",
		},
		{
			RuleID:  "module.rule.a",
			Path:    "a.cpp",
			Message: "a",
		},
	}
	want := []lint.Diagnostic{
		{
			RuleID:  "module.rule.a",
			Path:    "a.cpp",
			Message: "a",
		},
		{
			RuleID:  "module.rule.b",
			Path:    "b.cpp",
			Message: "b",
		},
	}

	if err := ValidateDiagnosticsEqual(got, want); err != nil {
		t.Fatalf("ValidateDiagnosticsEqual() error: %v", err)
	}
}

func TestValidateDiagnosticsEqualReportsMismatch(t *testing.T) {
	t.Parallel()

	got := []lint.Diagnostic{
		{
			RuleID:  "module.rule.a",
			Path:    "a.cpp",
			Message: "a",
		},
	}
	want := []lint.Diagnostic{
		{
			RuleID:  "module.rule.a",
			Path:    "a.cpp",
			Message: "b",
		},
	}

	err := ValidateDiagnosticsEqual(got, want)
	if err == nil {
		t.Fatal("ValidateDiagnosticsEqual() error=nil, want mismatch")
	}

	if !strings.Contains(err.Error(), "index 0") {
		t.Fatalf("error=%q, want mismatch index", err)
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linttest

import (
	"errors"
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestValidateCatalogContract(t *testing.T) {
	t.Parallel()

	catalog := []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
		lint.WarningCodeSpec(1002, "parse", "autofix inserted semicolon"),
	}
	ruleSpecs := []lint.RuleSpec{
		{
			ID:              "alpha.parse.unexpected-token",
			Module:          "alpha",
			Message:         "unexpected token",
			Description:     "alpha parse diagnostic: unexpected token",
			DefaultSeverity: lint.SeverityError,
			FileKinds:       []lint.FileKind{"source.cfg"},
		},
		{
			ID:              "alpha.parse.autofix-inserted-semicolon",
			Module:          "alpha",
			Message:         "autofix inserted semicolon",
			Description:     "alpha parse diagnostic: autofix inserted semicolon",
			DefaultSeverity: lint.SeverityWarning,
			FileKinds:       []lint.FileKind{"source.cfg"},
		},
	}
	ruleID := func(code lint.Code) string {
		switch code {
		case 1001:
			return "alpha.parse.unexpected-token"
		case 1002:
			return "alpha.parse.autofix-inserted-semicolon"
		default:
			return "alpha.unknown"
		}
	}

	if err := ValidateCatalogContract("alpha", catalog, ruleSpecs, ruleID); err != nil {
		t.Fatalf("ValidateCatalogContract() error: %v", err)
	}
}

func TestValidateCatalogContractDuplicateCode(t *testing.T) {
	t.Parallel()

	catalog := []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
		lint.WarningCodeSpec(1001, "parse", "duplicate code"),
	}
	ruleSpecs := []lint.RuleSpec{
		{
			ID:              "alpha.parse.unexpected-token",
			Module:          "alpha",
			Message:         "unexpected token",
			DefaultSeverity: lint.SeverityError,
		},
		{
			ID:              "alpha.parse.duplicate-code",
			Module:          "alpha",
			Message:         "duplicate code",
			DefaultSeverity: lint.SeverityWarning,
		},
	}
	ruleID := func(code lint.Code) string {
		if code == 1001 {
			return "alpha.parse.unexpected-token"
		}

		return "alpha.unknown"
	}

	err := ValidateCatalogContract("alpha", catalog, ruleSpecs, ruleID)
	if err == nil || !strings.Contains(err.Error(), "duplicate code") {
		t.Fatalf("ValidateCatalogContract(duplicate code) error=%v", err)
	}
}

func TestValidateCatalogContractUnknownRuleID(t *testing.T) {
	t.Parallel()

	catalog := []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
	}
	ruleSpecs := []lint.RuleSpec{
		{
			ID:              "alpha.parse.unexpected-token",
			Module:          "alpha",
			Message:         "unexpected token",
			DefaultSeverity: lint.SeverityError,
		},
	}
	ruleID := func(lint.Code) string {
		return "alpha.unknown"
	}

	err := ValidateCatalogContract("alpha", catalog, ruleSpecs, ruleID)
	if err == nil || !strings.Contains(err.Error(), "unknown rule id") {
		t.Fatalf("ValidateCatalogContract(unknown) error=%v", err)
	}
}

func TestValidateCatalogContractNonNormalizedFileKinds(t *testing.T) {
	t.Parallel()

	catalog := []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
	}
	ruleSpecs := []lint.RuleSpec{
		{
			ID:              "alpha.parse.unexpected-token",
			Module:          "alpha",
			Message:         "unexpected token",
			DefaultSeverity: lint.SeverityError,
			FileKinds:       []lint.FileKind{" Source.CFG "},
		},
	}
	ruleID := func(lint.Code) string {
		return "alpha.parse.unexpected-token"
	}

	err := ValidateCatalogContract("alpha", catalog, ruleSpecs, ruleID)
	if err == nil || !strings.Contains(err.Error(), "non-normalized file_kinds") {
		t.Fatalf("ValidateCatalogContract(non-normalized file kinds) error=%v", err)
	}
}

func TestValidateCodeCatalogContract(t *testing.T) {
	t.Parallel()

	catalog, err := lint.NewCodeCatalog(
		lint.CodeCatalogConfig{
			Module:     "alpha",
			CodePrefix: "ALPHA",
			ScopeDescriptions: map[lint.Stage]string{
				"parse": "Parser diagnostics.",
			},
		},
		[]lint.CodeSpec{
			lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	if err := ValidateCodeCatalogContract(catalog); err != nil {
		t.Fatalf("ValidateCodeCatalogContract() error: %v", err)
	}
}

func TestValidateCodeCatalogContractEmptyModule(t *testing.T) {
	t.Parallel()

	_, err := lint.NewCodeCatalog(
		lint.CodeCatalogConfig{
			Module:     " ",
			CodePrefix: "ALPHA",
		},
		nil,
	)
	if err == nil {
		t.Fatalf("NewCodeCatalog(empty module) error=nil, want failure")
	}

	// Build zero-value catalog directly to verify helper behavior.
	err = ValidateCodeCatalogContract(lint.CodeCatalog{})
	if !errors.Is(err, ErrEmptyCatalogModule) {
		t.Fatalf(
			"ValidateCodeCatalogContract(zero) error=%v, want ErrEmptyCatalogModule",
			err,
		)
	}
}

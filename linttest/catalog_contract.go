// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linttest

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

// AssertCatalogContract validates module lint catalog contract in tests.
func AssertCatalogContract(
	tb testing.TB,
	module string,
	catalog []lint.CodeSpec,
	ruleSpecs []lint.RuleSpec,
	ruleID func(code lint.Code) string,
) {
	tb.Helper()

	if err := ValidateCatalogContract(module, catalog, ruleSpecs, ruleID); err != nil {
		tb.Fatal(err)
	}
}

// AssertCodeCatalogContract validates lint.CodeCatalog contract in tests.
func AssertCodeCatalogContract(
	tb testing.TB,
	catalog lint.CodeCatalog,
) {
	tb.Helper()

	if err := ValidateCodeCatalogContract(catalog); err != nil {
		tb.Fatal(err)
	}
}

// ValidateCatalogContract validates module lint catalog contract.
func ValidateCatalogContract(
	module string,
	catalog []lint.CodeSpec,
	ruleSpecs []lint.RuleSpec,
	ruleID func(code lint.Code) string,
) error {
	module = strings.TrimSpace(module)
	if module == "" {
		return ErrEmptyCatalogModule
	}

	if ruleID == nil {
		return ErrNilRuleIDMapper
	}

	if len(catalog) != len(ruleSpecs) {
		return fmt.Errorf(
			"catalog contract: len(catalog)=%d, len(rule_specs)=%d",
			len(catalog),
			len(ruleSpecs),
		)
	}

	codeByID := make(map[string]lint.CodeSpec, len(catalog))
	idByCode := make(map[lint.Code]string, len(catalog))
	for index := range catalog {
		spec := catalog[index]
		code := spec.Code
		if code == 0 {
			return fmt.Errorf("catalog contract: catalog[%d] has empty code", index)
		}

		spec.Code = code
		if _, exists := idByCode[code]; exists {
			return fmt.Errorf("catalog contract: duplicate code %q", code)
		}

		if !lint.IsSupportedSeverity(spec.Severity) {
			return fmt.Errorf(
				"catalog contract: code %q has unsupported severity %q",
				code,
				spec.Severity,
			)
		}

		currentRuleID := strings.TrimSpace(ruleID(code))
		if currentRuleID == "" {
			return fmt.Errorf(
				"catalog contract: code %q maps to empty rule id",
				code,
			)
		}

		if currentRuleID == module+".unknown" {
			return fmt.Errorf(
				"catalog contract: code %q maps to unknown rule id %q",
				code,
				currentRuleID,
			)
		}

		if _, exists := codeByID[currentRuleID]; exists {
			return fmt.Errorf(
				"catalog contract: duplicate rule id %q for different codes",
				currentRuleID,
			)
		}

		codeByID[currentRuleID] = spec
		idByCode[code] = currentRuleID
	}

	registry := linting.NewRegistry()
	if err := registry.RegisterMany(ruleSpecs...); err != nil {
		return fmt.Errorf("catalog contract: invalid rule specs: %w", err)
	}

	byID := make(map[string]lint.RuleSpec, len(ruleSpecs))
	for index := range ruleSpecs {
		spec := ruleSpecs[index]
		if spec.Module != module {
			return fmt.Errorf(
				"catalog contract: rule_specs[%d].module=%q, want %q",
				index,
				spec.Module,
				module,
			)
		}

		if _, exists := byID[spec.ID]; exists {
			return fmt.Errorf("catalog contract: duplicate rule spec id %q", spec.ID)
		}

		if normalized := lint.NormalizeFileKinds(spec.FileKinds); !slices.Equal(
			spec.FileKinds,
			normalized,
		) {
			return fmt.Errorf(
				"catalog contract: rule %q has non-normalized file_kinds %v",
				spec.ID,
				spec.FileKinds,
			)
		}

		byID[spec.ID] = spec
	}

	for code, expectedID := range idByCode {
		spec, exists := byID[expectedID]
		if !exists {
			return fmt.Errorf(
				"catalog contract: missing rule spec for code %q mapped to %q",
				code,
				expectedID,
			)
		}

		catalogSpec := codeByID[expectedID]
		if spec.DefaultSeverity != catalogSpec.Severity {
			return fmt.Errorf(
				"catalog contract: rule %q default_severity=%q, want %q",
				expectedID,
				spec.DefaultSeverity,
				catalogSpec.Severity,
			)
		}
	}

	return nil
}

// ValidateCodeCatalogContract validates contract for one lint.CodeCatalog helper.
func ValidateCodeCatalogContract(catalog lint.CodeCatalog) error {
	module := strings.TrimSpace(catalog.ModuleSpec().ID)
	if module == "" {
		return ErrEmptyCatalogModule
	}

	return ValidateCatalogContract(
		module,
		catalog.CodeSpecs(),
		catalog.RuleSpecs(),
		func(code lint.Code) string {
			ruleID, err := catalog.RuleID(code)
			if err != nil {
				return module + ".unknown"
			}

			return ruleID
		},
	)
}

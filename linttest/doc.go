// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package linttest provides downstream catalog contract checks for tests.

Use it to ensure local catalog rows, exported rule specs,
and code->rule-id mapping remain consistent.

Quick start

	func TestCatalogContract(t *testing.T) {
		linttest.AssertCatalogContract(
			t,
			LintModule,
			DiagnosticCatalog(),
			LintRuleSpecs(),
			LintRuleID,
		)
	}

	func TestCodeCatalogContract(t *testing.T) {
		catalog, err := lint.NewCodeCatalog(...)
		if err != nil {
			t.Fatal(err)
		}

		linttest.AssertCodeCatalogContract(t, catalog)
	}

	func TestDiagnostics(t *testing.T) {
		linttest.AssertDiagnosticsEqual(t, gotDiagnostics, wantDiagnostics)
	}
*/
package linttest

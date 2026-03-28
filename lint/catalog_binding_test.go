// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"testing"
)

func TestNewCodeCatalogBindingAndRun(t *testing.T) {
	t.Parallel()

	catalog, err := NewCodeCatalog(
		CodeCatalogConfig{
			Module:     "alpha",
			CodePrefix: "ALPHA",
			ScopeDescriptions: map[Stage]string{
				"parse": "Parser diagnostics.",
			},
		},
		[]CodeSpec{
			ErrorCodeSpec(2001, "parse", "unexpected token"),
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	binding, err := NewCodeCatalogBinding(
		CodeCatalogBindingConfig[catalogProviderTestDiagnostic]{
			RunValueKey: "alpha.catalog.by_code",
			Catalog:     catalog,
			CodeFromDiagnostic: func(
				item catalogProviderTestDiagnostic,
			) Code {
				code, _ := ParseCode(item.Code)
				return code
			},
			DiagnosticToLint: func(
				item catalogProviderTestDiagnostic,
			) Diagnostic {
				return Diagnostic{
					Severity: SeverityError,
					Message:  item.Message,
				}
			},
			UnknownCodePolicy: UnknownCodeDrop,
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalogBinding() error: %v", err)
	}

	var registrar catalogProviderTestRegistrar
	if err := binding.RegisterRules(&registrar); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	run := RunContext{
		TargetPath: "workspace/main/source.cfg",
		TargetKind: "source.cfg",
	}
	ok := binding.Attach(&run, []catalogProviderTestDiagnostic{
		{Code: "2001", Message: "known"},
		{Code: "9999", Message: "unknown"},
	})
	if !ok {
		t.Fatal("Attach() returned false")
	}

	grouped, ok := GetIndexedByCode[catalogProviderTestDiagnostic, Code](
		&run,
		"alpha.catalog.by_code",
	)
	if !ok {
		t.Fatal("GetIndexedByCode() returned false")
	}

	if len(grouped[2001]) != 1 {
		t.Fatalf("grouped[2001] len=%d, want 1", len(grouped[2001]))
	}

	if len(grouped[9999]) != 0 {
		t.Fatalf("grouped[9999] len=%d, want 0", len(grouped[9999]))
	}

	diagnostics := make([]Diagnostic, 0, 1)
	err = registrar.runners[0].Check(
		context.Background(),
		&run,
		func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	if err != nil {
		t.Fatalf("runner.Check() error: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("len(diagnostics)=%d, want 1", len(diagnostics))
	}

	if diagnostics[0].RuleID != "alpha.parse.unexpected-token" {
		t.Fatalf(
			"diagnostics[0].RuleID=%q, want alpha.parse.unexpected-token",
			diagnostics[0].RuleID,
		)
	}
}

func TestNewCodeCatalogBindingValidateOptions(t *testing.T) {
	t.Parallel()

	catalog := CodeCatalog{}

	_, err := NewCodeCatalogBinding(
		CodeCatalogBindingConfig[catalogProviderTestDiagnostic]{
			RunValueKey: "bad",
			Catalog:     catalog,
		},
	)
	if !errors.Is(err, ErrInvalidCatalogProvider) {
		t.Fatalf(
			"NewCodeCatalogBinding(invalid key) error=%v, want ErrInvalidCatalogProvider",
			err,
		)
	}

	_, err = NewCodeCatalogBinding(
		CodeCatalogBindingConfig[catalogProviderTestDiagnostic]{
			RunValueKey: "alpha.catalog.by_code",
			Catalog:     catalog,
			DiagnosticToLint: func(item catalogProviderTestDiagnostic) Diagnostic {
				return Diagnostic{}
			},
		},
	)
	if !errors.Is(err, ErrInvalidCatalogProvider) {
		t.Fatalf(
			"NewCodeCatalogBinding(nil code mapper) error=%v, want ErrInvalidCatalogProvider",
			err,
		)
	}
}

func TestCodeCatalogBindingRegisterRulesByStage(t *testing.T) {
	t.Parallel()

	catalog, err := NewCodeCatalog(
		CodeCatalogConfig{
			Module:     "alpha",
			CodePrefix: "ALPHA",
			ScopeDescriptions: map[Stage]string{
				"parse":      "Parser diagnostics.",
				"preprocess": "Preprocessor diagnostics.",
			},
		},
		[]CodeSpec{
			ErrorCodeSpec(2001, "parse", "unexpected token"),
			ErrorCodeSpec(3001, "preprocess", "include not found"),
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	binding, err := NewCodeCatalogBinding(
		CodeCatalogBindingConfig[catalogProviderTestDiagnostic]{
			RunValueKey: "alpha.catalog.by_code",
			Catalog:     catalog,
			CodeFromDiagnostic: func(
				item catalogProviderTestDiagnostic,
			) Code {
				code, _ := ParseCode(item.Code)
				return code
			},
			DiagnosticToLint: func(
				item catalogProviderTestDiagnostic,
			) Diagnostic {
				return Diagnostic{
					Severity: SeverityError,
					Message:  item.Message,
				}
			},
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalogBinding() error: %v", err)
	}

	var registrar catalogProviderTestRegistrar
	if err := binding.RegisterRulesByStage(&registrar, Stage("parse")); err != nil {
		t.Fatalf("RegisterRulesByStage() error: %v", err)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "parse" {
		t.Fatalf(
			"registered scope=%q, want parse",
			registrar.runners[0].RuleSpec().Scope,
		)
	}
}

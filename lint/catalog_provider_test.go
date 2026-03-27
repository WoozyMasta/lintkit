// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"testing"
)

// catalogProviderTestItem stores one synthetic catalog metadata row.
type catalogProviderTestItem struct {
	// Code is stable synthetic code token.
	Code string
}

// catalogProviderTestDiagnostic stores one synthetic runtime diagnostic.
type catalogProviderTestDiagnostic struct {
	// Code is stable synthetic code token.
	Code string

	// Message is synthetic diagnostic text.
	Message string
}

// catalogProviderTestRegistrar stores registered runners for tests.
type catalogProviderTestRegistrar struct {
	// runners stores registered runners.
	runners []RuleRunner

	// modules stores registered module descriptors.
	modules []ModuleSpec
}

// Register appends provided runners into in-memory test storage.
func (registrar *catalogProviderTestRegistrar) Register(
	runners ...RuleRunner,
) error {
	registrar.runners = append(registrar.runners, runners...)
	return nil
}

// RegisterModule appends module descriptor into in-memory test storage.
func (registrar *catalogProviderTestRegistrar) RegisterModule(
	spec ModuleSpec,
) error {
	registrar.modules = append(registrar.modules, spec)
	return nil
}

func TestNewCatalogProviderAndRun(t *testing.T) {
	t.Parallel()

	provider, err := NewCatalogProvider(
		"test.catalog.by_code",
		[]catalogProviderTestItem{
			{Code: "A001"},
			{Code: "B001"},
		},
		func(item catalogProviderTestItem) RuleSpec {
			return RuleSpec{
				ID:              "module_alpha." + item.Code,
				Module:          "module_alpha",
				Message:         "rule " + item.Code,
				DefaultSeverity: SeverityWarning,
				FileKinds:       []FileKind{"source.cfg"},
			}
		},
		func(item catalogProviderTestItem) string {
			return item.Code
		},
		func(diagnostic catalogProviderTestDiagnostic) Diagnostic {
			return Diagnostic{
				RuleID:   "module_alpha." + diagnostic.Code,
				Severity: SeverityError,
				Message:  diagnostic.Message,
			}
		},
	)
	if err != nil {
		t.Fatalf("NewCatalogProvider() error: %v", err)
	}

	var registrar catalogProviderTestRegistrar
	if err := provider.RegisterRules(&registrar); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if len(registrar.runners) != 2 {
		t.Fatalf("registered runners=%d, want 2", len(registrar.runners))
	}

	if len(registrar.modules) != 1 {
		t.Fatalf("registered modules=%d, want 1", len(registrar.modules))
	}

	if registrar.modules[0].ID != "module_alpha" {
		t.Fatalf("modules[0].ID=%q, want module_alpha", registrar.modules[0].ID)
	}

	runner := registrar.runners[0]
	runContext := RunContext{
		TargetPath: "workspace/main/source.cfg",
		TargetKind: "source.cfg",
	}
	ok := AttachCatalogDiagnostics(
		&runContext,
		"test.catalog.by_code",
		[]catalogProviderTestDiagnostic{
			{Code: "A001", Message: "first"},
			{Code: "B001", Message: "second"},
			{Code: "A001", Message: "third"},
		},
		func(diagnostic catalogProviderTestDiagnostic) string {
			return diagnostic.Code
		},
	)
	if !ok {
		t.Fatal("AttachCatalogDiagnostics() returned false")
	}

	diagnostics := make([]Diagnostic, 0, 2)
	err = runner.Check(
		context.Background(),
		&runContext,
		func(diagnostic Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	if err != nil {
		t.Fatalf("runner.Check() error: %v", err)
	}

	if len(diagnostics) != 2 {
		t.Fatalf("len(Diagnostics)=%d, want 2", len(diagnostics))
	}
}

func TestNewCatalogProviderValidateOptions(t *testing.T) {
	t.Parallel()

	_, err := NewCatalogProvider[catalogProviderTestItem, catalogProviderTestDiagnostic, string](
		"",
		nil,
		nil,
		nil,
		nil,
	)
	if !errors.Is(err, ErrInvalidCatalogProvider) {
		t.Fatalf("NewCatalogProvider(empty) error=%v, want ErrInvalidCatalogProvider", err)
	}
}

func TestCatalogProviderRegisterRulesNilRegistrar(t *testing.T) {
	t.Parallel()

	provider, err := NewCatalogProvider(
		"test.catalog.by_code",
		[]catalogProviderTestItem{
			{Code: "A001"},
		},
		func(item catalogProviderTestItem) RuleSpec {
			return RuleSpec{
				ID:              "module_alpha." + item.Code,
				Module:          "module_alpha",
				Message:         "rule",
				DefaultSeverity: SeverityWarning,
			}
		},
		func(item catalogProviderTestItem) string {
			return item.Code
		},
		func(diagnostic catalogProviderTestDiagnostic) Diagnostic {
			return Diagnostic{
				RuleID:   "module_alpha." + diagnostic.Code,
				Severity: SeverityWarning,
			}
		},
	)
	if err != nil {
		t.Fatalf("NewCatalogProvider() error: %v", err)
	}

	if err := provider.RegisterRules(nil); !errors.Is(err, ErrNilRuleRegistrar) {
		t.Fatalf("RegisterRules(nil) error=%v, want ErrNilRuleRegistrar", err)
	}
}

func TestAttachCatalogDiagnosticsNilCodeMapper(t *testing.T) {
	t.Parallel()

	run := RunContext{}
	ok := AttachCatalogDiagnostics[catalogProviderTestDiagnostic, string](
		&run,
		"test.catalog.by_code",
		nil,
		nil,
	)
	if ok {
		t.Fatal("AttachCatalogDiagnostics(nil mapper) returned true, want false")
	}
}

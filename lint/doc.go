// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package lint provides downstream lint contracts and helper primitives.

Use this package in modules that own diagnostics and rule metadata.
It does not execute lint checks.

RunContext value keys use namespaced form "module.key" to avoid
cross-module collisions in shared runtime state.

Quick start: define lazy catalog handle without panic

	import "github.com/woozymasta/lintkit/lint"

	var catalogHandle = lint.NewCodeCatalogHandle(lint.CodeCatalogConfig{
		Module:            "module_alpha",
		CodePrefix:        "ALPHA",
		ScopeDescriptions: map[lint.Stage]string{"parse": "Parser diagnostics."},
	}, []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
	})

Quick start: build unified register+attach binding

	type itemDiagnostic struct {
		Code    lint.Code
		Message string
	}

	catalog, err := catalogHandle.Catalog()
	if err != nil {
		return err
	}

	binding, err := lint.NewCodeCatalogBinding(
		lint.CodeCatalogBindingConfig[itemDiagnostic]{
			RunValueKey: "module_alpha.by_code",
			Catalog:     catalog,
			CodeFromDiagnostic: func(item itemDiagnostic) lint.Code {
				return item.Code
			},
			DiagnosticToLint: func(item itemDiagnostic) lint.Diagnostic {
				return lint.Diagnostic{
					Message: item.Message,
				}
			},
			UnknownCodePolicy: lint.UnknownCodeDrop,
		},
	)
	if err != nil {
		return err
	}

Quick start: attach precomputed module data to run context

	_ = binding.Attach(&runContext, diagnostics)

Quick start: fail utility flow by diagnostics threshold

	if err := lint.ErrorFromDiagnostics(diagnostics, lint.SeverityError); err != nil {
		return err
	}
*/
package lint

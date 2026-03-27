// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package lint provides downstream lint contracts and helper primitives.

Use this package in modules that own diagnostics and rule metadata.
It does not execute lint checks.

RunContext value keys use namespaced form "module.key" to avoid
cross-module collisions in shared runtime state.

Quick start: define one catalog

	import "github.com/woozymasta/lintkit/lint"

	var catalog, _ = lint.NewCodeCatalog(lint.CodeCatalogConfig{
		Module:            "module_alpha",
		CodePrefix:        "ALPHA",
		ScopeDescriptions: map[lint.Stage]string{"parse": "Parser diagnostics."},
	}, []lint.CodeSpec{
		lint.ErrorCodeSpec(1001, "parse", "unexpected token"),
	})

Quick start: build provider from catalog

	type itemDiagnostic struct {
		Code    lint.Code
		Message string
	}

	provider, _ := lint.NewCodeCatalogProvider(
		"module_alpha.by_code",
		catalog,
		func(item itemDiagnostic) lint.Diagnostic {
			ruleID, err := catalog.RuleID(item.Code)
			if err != nil {
				ruleID = ""
			}

			return lint.Diagnostic{
				RuleID:  ruleID,
				Message: item.Message,
			}
		},
	)

Quick start: attach precomputed module data to run context

	lint.SetRunValue(&runContext, "module_alpha.ast", astValue)
	ast, ok := lint.GetRunValue[*AST](&runContext, "module_alpha.ast")
	_ = ast
	_ = ok
*/
package lint

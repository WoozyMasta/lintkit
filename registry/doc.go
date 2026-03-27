// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package registry provides in-process registry snapshot assembly helpers.

Use this package when provider packages are already linked in your binary
and you need deterministic snapshot data without invoking CLI collection.

Quick start

	snapshot, err := registry.SnapshotFromProviders(
		modulea.LintRulesProvider{},
		moduleb.LintRulesProvider{},
	)
	if err != nil {
		return err
	}

	_ = snapshot
*/
package registry

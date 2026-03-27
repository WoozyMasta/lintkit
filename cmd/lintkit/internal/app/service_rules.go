// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

// serviceDiagnosticsEnabled reports whether service diagnostics are enabled.
func serviceDiagnosticsEnabled(options SnapshotCommandOptions) bool {
	if options.EnableServiceDiagnostics == nil {
		return true
	}

	return *options.EnableServiceDiagnostics
}

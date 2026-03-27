// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"testing"
)

// TestServiceDiagnosticsEnabledDefaultsTrue checks default enable state.
func TestServiceDiagnosticsEnabledDefaultsTrue(t *testing.T) {
	if !serviceDiagnosticsEnabled(SnapshotCommandOptions{}) {
		t.Fatalf("serviceDiagnosticsEnabled(default)=false, want true")
	}
}

// TestServiceDiagnosticsEnabledRespectsExplicitFalse checks explicit disable.
func TestServiceDiagnosticsEnabledRespectsExplicitFalse(t *testing.T) {
	disabled := false
	if serviceDiagnosticsEnabled(SnapshotCommandOptions{
		EnableServiceDiagnostics: &disabled,
	}) {
		t.Fatalf("serviceDiagnosticsEnabled(false)=true, want false")
	}
}

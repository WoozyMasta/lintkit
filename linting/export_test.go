// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"encoding/json"
	"github.com/woozymasta/lintkit/lint"
	"testing"
)

func TestRegistryExportJSON(t *testing.T) {
	t.Parallel()

	registry := buildTestRegistry(t)

	payload, err := registry.ExportJSON(true)
	if err != nil {
		t.Fatalf("ExportJSON() error: %v", err)
	}

	var snapshot lint.RegistrySnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("Unmarshal(export) error: %v", err)
	}

	if len(snapshot.Rules) != 2 {
		t.Fatalf("len(snapshot.Rules)=%d, want 2", len(snapshot.Rules))
	}

	if len(snapshot.Modules) != 2 {
		t.Fatalf("len(snapshot.Modules)=%d, want 2", len(snapshot.Modules))
	}

	if snapshot.Modules[0].ID != "module_alpha" {
		t.Fatalf("snapshot.Modules[0].ID=%q", snapshot.Modules[0].ID)
	}

	if snapshot.Rules[0].ID != "module_alpha.R015" {
		t.Fatalf("snapshot.Rules[0].ID=%q", snapshot.Rules[0].ID)
	}
}

// buildTestRegistry builds deterministic registry test fixture.
func buildTestRegistry(t *testing.T) *Registry {
	t.Helper()

	registry := NewRegistry()
	if err := registry.RegisterMany(
		lint.RuleSpec{
			ID:              "module_alpha.R015",
			Module:          "module_alpha",
			Code:            "PAR015",
			Message:         "macro redefinition",
			Description:     "Warn when macro is redefined.",
			DefaultSeverity: lint.SeverityWarning,
			DefaultEnabled:  BoolPtr(false),
			DefaultOptions: map[string]any{
				"mode": "strict",
			},
			FileKinds: []lint.FileKind{"source.cfg"},
		},
		lint.RuleSpec{
			ID:              "module_gamma.asset_exists",
			Module:          "module_gamma",
			Message:         "asset exists",
			Description:     "Warn when asset file cannot be resolved.",
			DefaultSeverity: lint.SeverityWarning,
			FileKinds:       []lint.FileKind{"module_gamma"},
		},
	); err != nil {
		t.Fatalf("RegisterMany() error: %v", err)
	}

	if err := registry.RegisterModule(lint.ModuleSpec{
		ID:          "module_alpha",
		Name:        "Module Alpha",
		Description: "Rules for module_alpha parse pipeline.",
	}); err != nil {
		t.Fatalf("RegisterModule(module_alpha) error: %v", err)
	}

	if err := registry.RegisterModule(lint.ModuleSpec{
		ID:          "module_gamma",
		Name:        "Module Gamma",
		Description: "Rules for module_gamma validation pipeline.",
	}); err != nil {
		t.Fatalf("RegisterModule(module_gamma) error: %v", err)
	}

	return registry
}

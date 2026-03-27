// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestRegistryRegisterAndRuleLookup(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	spec := lint.RuleSpec{
		ID:              "module_alpha.R015",
		Module:          "module_alpha",
		Message:         "macro redefinition",
		Description:     "Warn when macro is redefined.",
		DefaultSeverity: lint.SeverityWarning,
		FileKinds:       []lint.FileKind{"source.cfg"},
	}

	if err := registry.Register(spec); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	got, ok := registry.Rule("module_alpha.R015")
	if !ok {
		t.Fatal("Rule() returned not found")
	}

	if got.ID != spec.ID {
		t.Fatalf("Rule().ID=%q, want %q", got.ID, spec.ID)
	}

	if got.Module != spec.Module {
		t.Fatalf("Rule().Module=%q, want %q", got.Module, spec.Module)
	}
}

func TestRegistryRegisterDuplicateRuleID(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	spec := lint.RuleSpec{
		ID:              "module_beta.missing_resource",
		Module:          "module_beta",
		Message:         "missing resource",
		DefaultSeverity: lint.SeverityWarning,
	}

	if err := registry.Register(spec); err != nil {
		t.Fatalf("Register(first) error: %v", err)
	}

	err := registry.Register(spec)
	if !errors.Is(err, ErrDuplicateRuleID) {
		t.Fatalf("Register(duplicate) error=%v, want ErrDuplicateRuleID", err)
	}
}

func TestRegistryRegisterDuplicateRuleCode(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(lint.RuleSpec{
		ID:              "module_alpha.parse.first",
		Module:          "module_alpha",
		Code:            "ALPHA001",
		Message:         "first rule",
		DefaultSeverity: lint.SeverityWarning,
	}); err != nil {
		t.Fatalf("Register(first) error: %v", err)
	}

	err := registry.Register(lint.RuleSpec{
		ID:              "module_beta.parse.second",
		Module:          "module_beta",
		Code:            "ALPHA001",
		Message:         "second rule",
		DefaultSeverity: lint.SeverityWarning,
	})
	if !errors.Is(err, ErrDuplicateRuleCode) {
		t.Fatalf("Register(duplicate code) error=%v, want ErrDuplicateRuleCode", err)
	}
}

func TestRegistryRegisterManyIsAtomicAgainstExistingDuplicates(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()

	if err := registry.Register(lint.RuleSpec{
		ID:              "pofile.PO2001",
		Module:          "pofile",
		Message:         "duplicate entry",
		DefaultSeverity: lint.SeverityError,
	}); err != nil {
		t.Fatalf("Register(seed) error: %v", err)
	}

	err := registry.RegisterMany(
		lint.RuleSpec{
			ID:              "imageset.duplicate_name",
			Module:          "imageset",
			Message:         "duplicate name",
			DefaultSeverity: lint.SeverityError,
		},
		lint.RuleSpec{
			ID:              "pofile.PO2001",
			Module:          "pofile",
			Message:         "duplicate entry",
			DefaultSeverity: lint.SeverityError,
		},
	)
	if !errors.Is(err, ErrDuplicateRuleID) {
		t.Fatalf("RegisterMany() error=%v, want ErrDuplicateRuleID", err)
	}

	if _, ok := registry.Rule("imageset.duplicate_name"); ok {
		t.Fatal("RegisterMany() inserted partial batch on error")
	}
}

func TestRegistryRegisterManyDuplicateCodeInBatch(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	err := registry.RegisterMany(
		lint.RuleSpec{
			ID:              "module_alpha.parse.a",
			Module:          "module_alpha",
			Code:            "DUP001",
			Message:         "a",
			DefaultSeverity: lint.SeverityWarning,
		},
		lint.RuleSpec{
			ID:              "module_beta.parse.b",
			Module:          "module_beta",
			Code:            "DUP001",
			Message:         "b",
			DefaultSeverity: lint.SeverityWarning,
		},
	)
	if !errors.Is(err, ErrDuplicateRuleCode) {
		t.Fatalf("RegisterMany(duplicate code) error=%v, want ErrDuplicateRuleCode", err)
	}

	if got := len(registry.Rules()); got != 0 {
		t.Fatalf("RegisterMany inserted partial batch, rules=%d", got)
	}
}

func TestRegistryRulesDeterministicOrder(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	err := registry.RegisterMany(
		lint.RuleSpec{
			ID:              "module_beta.z_rule",
			Module:          "module_beta",
			Message:         "z",
			DefaultSeverity: lint.SeverityWarning,
		},
		lint.RuleSpec{
			ID:              "module_gamma.a_rule",
			Module:          "module_gamma",
			Message:         "a",
			DefaultSeverity: lint.SeverityWarning,
		},
		lint.RuleSpec{
			ID:              "module_beta.a_rule",
			Module:          "module_beta",
			Message:         "a",
			DefaultSeverity: lint.SeverityWarning,
		},
	)
	if err != nil {
		t.Fatalf("RegisterMany() error: %v", err)
	}

	got := registry.Rules()
	if len(got) != 3 {
		t.Fatalf("len(Rules())=%d, want 3", len(got))
	}

	want := []string{"module_beta.a_rule", "module_beta.z_rule", "module_gamma.a_rule"}
	for index := range want {
		if got[index].ID != want[index] {
			t.Fatalf("Rules()[%d].ID=%q, want %q", index, got[index].ID, want[index])
		}
	}
}

func TestRegistryRegisterValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    lint.RuleSpec
		wantErr error
	}{
		{
			name: "empty_id",
			spec: lint.RuleSpec{
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_id",
			spec: lint.RuleSpec{
				ID:      "bad id",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_id_starts_with_digit",
			spec: lint.RuleSpec{
				ID:      "1module_alpha.PAR001",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_id_without_dot",
			spec: lint.RuleSpec{
				ID:      "module_alpha",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_id_trailing_dot",
			spec: lint.RuleSpec{
				ID:      "module_alpha.PAR001.",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "empty_module",
			spec: lint.RuleSpec{
				ID:      "module_alpha.PAR001",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "module_prefix_mismatch",
			spec: lint.RuleSpec{
				ID:      "module_beta.R001",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "empty_message",
			spec: lint.RuleSpec{
				ID:     "module_alpha.PAR001",
				Module: "module_alpha",
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_severity",
			spec: lint.RuleSpec{
				ID:              "module_alpha.PAR001",
				Module:          "module_alpha",
				Message:         "title",
				DefaultSeverity: lint.Severity("fatal"),
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "invalid_scope",
			spec: lint.RuleSpec{
				ID:              "module_alpha.PAR002",
				Module:          "module_alpha",
				Scope:           "bad scope",
				Message:         "title",
				DefaultSeverity: lint.SeverityWarning,
			},
			wantErr: ErrInvalidRuleSpec,
		},
		{
			name: "valid_defaults",
			spec: lint.RuleSpec{
				ID:      "module_alpha.PAR001",
				Module:  "module_alpha",
				Message: "title",
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRegistry()
			err := registry.Register(test.spec)
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Register() error=%v, want nil", err)
				}

				got, ok := registry.Rule(test.spec.ID)
				if !ok {
					t.Fatal("Rule() returned not found for valid spec")
				}

				if got.DefaultSeverity != lint.SeverityWarning {
					t.Fatalf(
						"Rule().DefaultSeverity=%q, want %q",
						got.DefaultSeverity,
						lint.SeverityWarning,
					)
				}
				return
			}

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Register() error=%v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRegistryRegisterModule(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.RegisterModule(lint.ModuleSpec{
		ID:          "module_alpha",
		Name:        "Module Alpha",
		Description: "Rules for module_alpha.",
	}); err != nil {
		t.Fatalf("RegisterModule() error: %v", err)
	}

	module, ok := registry.Module("module_alpha")
	if !ok {
		t.Fatal("Module() returned not found")
	}

	if module.Name != "Module Alpha" {
		t.Fatalf("Module().Name=%q, want Module Alpha", module.Name)
	}
}

func TestRegistryRegisterModuleConflict(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.RegisterModule(lint.ModuleSpec{
		ID:   "module_alpha",
		Name: "Module Alpha",
	}); err != nil {
		t.Fatalf("RegisterModule(first) error: %v", err)
	}

	err := registry.RegisterModule(lint.ModuleSpec{
		ID:   "module_alpha",
		Name: "Other Alpha",
	})
	if !errors.Is(err, ErrConflictingModuleSpec) {
		t.Fatalf("RegisterModule(conflict) error=%v, want ErrConflictingModuleSpec", err)
	}
}

func TestRegistrySnapshotIncludesModules(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.RegisterModule(lint.ModuleSpec{
		ID:          "module_alpha",
		Name:        "Module Alpha",
		Description: "Rules for module_alpha.",
	}); err != nil {
		t.Fatalf("RegisterModule() error: %v", err)
	}

	if err := registry.Register(lint.RuleSpec{
		ID:              "module_alpha.parse.rule",
		Module:          "module_alpha",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
	}); err != nil {
		t.Fatalf("Register(rule) error: %v", err)
	}

	snapshot := registry.Snapshot()
	if len(snapshot.Modules) != 1 {
		t.Fatalf("len(Snapshot().Modules)=%d, want 1", len(snapshot.Modules))
	}

	if snapshot.Modules[0].ID != "module_alpha" {
		t.Fatalf("Snapshot().Modules[0].ID=%q, want module_alpha", snapshot.Modules[0].ID)
	}
}

func TestRegistrySnapshotCacheInvalidationAndClone(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(lint.RuleSpec{
		ID:              "module_alpha.parse.rule_one",
		Module:          "module_alpha",
		Message:         "rule one",
		DefaultSeverity: lint.SeverityWarning,
	}); err != nil {
		t.Fatalf("Register(rule_one) error: %v", err)
	}

	first := registry.Snapshot()
	if len(first.Rules) != 1 {
		t.Fatalf("len(first.Rules)=%d, want 1", len(first.Rules))
	}

	first.Rules[0].ID = "mutated"
	second := registry.Snapshot()
	if second.Rules[0].ID == "mutated" {
		t.Fatal("Snapshot() returned shared cached slice, want cloned value")
	}

	if err := registry.Register(lint.RuleSpec{
		ID:              "module_alpha.parse.rule_two",
		Module:          "module_alpha",
		Message:         "rule two",
		DefaultSeverity: lint.SeverityWarning,
	}); err != nil {
		t.Fatalf("Register(rule_two) error: %v", err)
	}

	third := registry.Snapshot()
	if len(third.Rules) != 2 {
		t.Fatalf("len(third.Rules)=%d, want 2", len(third.Rules))
	}
}

func TestRegistryRuleDetachedFromInputAndOutputMutations(t *testing.T) {
	t.Parallel()

	enabled := true
	spec := lint.RuleSpec{
		ID:              "module_alpha.parse.rule",
		Module:          "module_alpha",
		Message:         "rule",
		DefaultSeverity: lint.SeverityWarning,
		DefaultEnabled:  &enabled,
		DefaultOptions: map[string]any{
			"mode": "strict",
		},
		FileKinds: []lint.FileKind{"source.cfg"},
	}

	registry := NewRegistry()
	if err := registry.Register(spec); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	enabled = false
	spec.FileKinds[0] = "mutated.kind"
	specOptions := spec.DefaultOptions.(map[string]any)
	specOptions["mode"] = "mutated"

	stored, ok := registry.Rule("module_alpha.parse.rule")
	if !ok {
		t.Fatal("Rule() returned not found")
	}

	if stored.DefaultEnabled == nil || !*stored.DefaultEnabled {
		t.Fatal("stored.DefaultEnabled mutated via source pointer")
	}

	if stored.FileKinds[0] != "source.cfg" {
		t.Fatalf("stored.FileKinds[0]=%q, want source.cfg", stored.FileKinds[0])
	}

	storedOptions, ok := stored.DefaultOptions.(map[string]any)
	if !ok {
		t.Fatalf("stored.DefaultOptions type=%T, want map[string]any", stored.DefaultOptions)
	}

	if storedOptions["mode"] != "strict" {
		t.Fatalf("stored.DefaultOptions[mode]=%v, want strict", storedOptions["mode"])
	}

	stored.FileKinds[0] = "changed.after.read"
	storedOptions["mode"] = "changed.after.read"
	*stored.DefaultEnabled = false

	readAgain, ok := registry.Rule("module_alpha.parse.rule")
	if !ok {
		t.Fatal("Rule() returned not found after mutation")
	}

	if readAgain.FileKinds[0] != "source.cfg" {
		t.Fatalf("readAgain.FileKinds[0]=%q, want source.cfg", readAgain.FileKinds[0])
	}

	readAgainOptions, ok := readAgain.DefaultOptions.(map[string]any)
	if !ok {
		t.Fatalf(
			"readAgain.DefaultOptions type=%T, want map[string]any",
			readAgain.DefaultOptions,
		)
	}

	if readAgainOptions["mode"] != "strict" {
		t.Fatalf("readAgain.DefaultOptions[mode]=%v, want strict", readAgainOptions["mode"])
	}

	if readAgain.DefaultEnabled == nil || !*readAgain.DefaultEnabled {
		t.Fatal("readAgain.DefaultEnabled mutated via returned pointer")
	}
}

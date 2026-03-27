// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"testing"
)

// providerTestRunner stores one synthetic runner.
type providerTestRunner struct {
	// spec stores synthetic metadata.
	spec RuleSpec
}

// RuleSpec returns synthetic metadata.
func (runner providerTestRunner) RuleSpec() RuleSpec {
	return runner.spec
}

// Check does nothing in tests.
func (providerTestRunner) Check(
	_ context.Context,
	_ *RunContext,
	_ DiagnosticEmit,
) error {
	return nil
}

// providerTestRegistrar stores registered runners for tests.
type providerTestRegistrar struct {
	// runners stores registered runner list.
	runners []RuleRunner

	// modules stores registered module metadata list.
	modules []ModuleSpec
}

// Register appends runners to in-memory test storage.
func (registrar *providerTestRegistrar) Register(
	runners ...RuleRunner,
) error {
	registrar.runners = append(registrar.runners, runners...)
	return nil
}

// RegisterModule stores module metadata in in-memory test storage.
func (registrar *providerTestRegistrar) RegisterModule(spec ModuleSpec) error {
	registrar.modules = append(registrar.modules, spec)
	return nil
}

func TestRegisterRuleProviders(t *testing.T) {
	t.Parallel()

	var registrar providerTestRegistrar
	err := RegisterRuleProviders(
		&registrar,
		RuleProviderFunc(func(registrar RuleRegistrar) error {
			return registrar.Register(providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			})
		}),
		RuleProviderFunc(func(registrar RuleRegistrar) error {
			return registrar.Register(providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a002",
					Module:          "module_alpha",
					Message:         "second",
					DefaultSeverity: SeverityWarning,
				},
			})
		}),
	)
	if err != nil {
		t.Fatalf("RegisterRuleProviders() error: %v", err)
	}

	if len(registrar.runners) != 2 {
		t.Fatalf("registered runners=%d, want 2", len(registrar.runners))
	}
}

func TestRegisterRuleProvidersNilRegistrar(t *testing.T) {
	t.Parallel()

	err := RegisterRuleProviders(nil, RuleProviderFunc(func(RuleRegistrar) error {
		return nil
	}))
	if !errors.Is(err, ErrNilRuleRegistrar) {
		t.Fatalf("RegisterRuleProviders(nil) error=%v, want ErrNilRuleRegistrar", err)
	}
}

func TestRegisterRuleProvidersNilProvider(t *testing.T) {
	t.Parallel()

	var registrar providerTestRegistrar

	var nilProvider RuleProvider
	err := RegisterRuleProviders(&registrar, nilProvider)
	if !errors.Is(err, ErrNilRuleProvider) {
		t.Fatalf("RegisterRuleProviders(nil provider) error=%v, want ErrNilRuleProvider", err)
	}

	var nilFuncProvider RuleProviderFunc
	err = RegisterRuleProviders(&registrar, nilFuncProvider)
	if !errors.Is(err, ErrNilRuleProvider) {
		t.Fatalf("RegisterRuleProviders(nil func provider) error=%v, want ErrNilRuleProvider", err)
	}
}

func TestRegisterRuleProvidersWithModuleSpec(t *testing.T) {
	t.Parallel()

	var registrar providerTestRegistrar
	err := RegisterRuleProviders(
		&registrar,
		WithModuleSpec(
			RuleProviderFunc(func(registrar RuleRegistrar) error {
				return registrar.Register(providerTestRunner{
					spec: RuleSpec{
						ID:              "module_alpha.parse.a001",
						Module:          "module_alpha",
						Message:         "first",
						DefaultSeverity: SeverityWarning,
					},
				})
			}),
			ModuleSpec{
				ID:          "module_alpha",
				Name:        "Module Alpha",
				Description: "Rules for module_alpha.",
			},
		),
	)
	if err != nil {
		t.Fatalf("RegisterRuleProviders() error: %v", err)
	}

	if len(registrar.modules) != 1 {
		t.Fatalf("registered modules=%d, want 1", len(registrar.modules))
	}

	if registrar.modules[0].ID != "module_alpha" {
		t.Fatalf("modules[0].ID=%q, want module_alpha", registrar.modules[0].ID)
	}
}

func TestComposeProvidersNilOnEmpty(t *testing.T) {
	t.Parallel()

	if provider := ComposeProviders(); provider != nil {
		t.Fatalf("ComposeProviders()=%T, want nil", provider)
	}
}

func TestComposeProvidersReturnsSingle(t *testing.T) {
	t.Parallel()

	single := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(providerTestRunner{
			spec: RuleSpec{
				ID:              "module_alpha.parse.a001",
				Module:          "module_alpha",
				Message:         "single",
				DefaultSeverity: SeverityWarning,
			},
		})
	})

	provider := ComposeProviders(single)
	if provider == nil {
		t.Fatal("ComposeProviders(single)=nil")
	}

	var registrar providerTestRegistrar
	if err := provider.RegisterRules(&registrar); err != nil {
		t.Fatalf("ComposeProviders(single).RegisterRules() error: %v", err)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("len(composed single runners)=%d, want 1", len(registrar.runners))
	}
}

func TestComposeProvidersAppliesInOrder(t *testing.T) {
	t.Parallel()

	first := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(providerTestRunner{
			spec: RuleSpec{
				ID:              "module_alpha.parse.a001",
				Module:          "module_alpha",
				Message:         "first",
				DefaultSeverity: SeverityWarning,
			},
		})
	})
	second := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(providerTestRunner{
			spec: RuleSpec{
				ID:              "module_alpha.parse.a002",
				Module:          "module_alpha",
				Message:         "second",
				DefaultSeverity: SeverityWarning,
			},
		})
	})

	composed := ComposeProviders(first, second)
	if composed == nil {
		t.Fatal("ComposeProviders(first, second)=nil")
	}

	var registrar providerTestRegistrar
	if err := composed.RegisterRules(&registrar); err != nil {
		t.Fatalf("composed.RegisterRules() error: %v", err)
	}

	if len(registrar.runners) != 2 {
		t.Fatalf("len(composed runners)=%d, want 2", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().ID != "module_alpha.parse.a001" {
		t.Fatalf("first composed rule id=%q", registrar.runners[0].RuleSpec().ID)
	}

	if registrar.runners[1].RuleSpec().ID != "module_alpha.parse.a002" {
		t.Fatalf("second composed rule id=%q", registrar.runners[1].RuleSpec().ID)
	}
}

func TestComposeProvidersForwardsModuleMetadata(t *testing.T) {
	t.Parallel()

	moduleProvider := WithModuleSpec(
		RuleProviderFunc(func(registrar RuleRegistrar) error {
			return registrar.Register(providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			})
		}),
		ModuleSpec{
			ID:          "module_alpha",
			Name:        "Module Alpha",
			Description: "Rules for module_alpha.",
		},
	)

	composed := ComposeProviders(moduleProvider)
	var registrar providerTestRegistrar
	if err := RegisterRuleProviders(&registrar, composed); err != nil {
		t.Fatalf("RegisterRuleProviders(composed) error: %v", err)
	}

	if len(registrar.modules) != 1 {
		t.Fatalf("len(modules)=%d, want 1", len(registrar.modules))
	}

	if registrar.modules[0].ID != "module_alpha" {
		t.Fatalf("modules[0].ID=%q, want module_alpha", registrar.modules[0].ID)
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestEngineRegisterProviders(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.RegisterProviders(
		lint.RuleProviderFunc(func(registrar lint.RuleRegistrar) error {
			return registrar.Register(testRunner{
				spec: lint.RuleSpec{
					ID:              "module_alpha.R001",
					Module:          "module_alpha",
					Message:         "provider one",
					DefaultSeverity: lint.SeverityWarning,
				},
			})
		}),
		lint.RuleProviderFunc(func(registrar lint.RuleRegistrar) error {
			return registrar.Register(testRunner{
				spec: lint.RuleSpec{
					ID:              "module_beta.R001",
					Module:          "module_beta",
					Message:         "provider two",
					DefaultSeverity: lint.SeverityWarning,
				},
			})
		}),
	)
	if err != nil {
		t.Fatalf("RegisterProviders() error: %v", err)
	}

	rules := engine.Rules()
	if len(rules) != 2 {
		t.Fatalf("len(Rules())=%d, want 2", len(rules))
	}
}

func TestRegisterProvidersNilEngine(t *testing.T) {
	t.Parallel()

	err := RegisterProviders(nil, lint.RuleProviderFunc(func(lint.RuleRegistrar) error {
		return nil
	}))
	if !errors.Is(err, ErrNilEngine) {
		t.Fatalf("RegisterProviders(nil) error=%v, want ErrNilEngine", err)
	}
}

func TestRegisterProvidersNilProvider(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	var nilProvider lint.RuleProvider
	err := engine.RegisterProviders(nilProvider)
	if !errors.Is(err, ErrNilRuleProvider) {
		t.Fatalf("RegisterProviders(nil provider) error=%v, want ErrNilRuleProvider", err)
	}

	var nilFuncProvider lint.RuleProviderFunc
	err = engine.RegisterProviders(nilFuncProvider)
	if !errors.Is(err, ErrNilRuleProvider) {
		t.Fatalf("RegisterProviders(nil func provider) error=%v, want ErrNilRuleProvider", err)
	}
}

func TestRegisterProvidersDuplicateRuleID(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	err := engine.RegisterProviders(
		lint.RuleProviderFunc(func(registrar lint.RuleRegistrar) error {
			return registrar.Register(testRunner{
				spec: lint.RuleSpec{
					ID:              "module_alpha.R015",
					Module:          "module_alpha",
					Message:         "first",
					DefaultSeverity: lint.SeverityWarning,
				},
			})
		}),
		lint.RuleProviderFunc(func(registrar lint.RuleRegistrar) error {
			return registrar.Register(testRunner{
				spec: lint.RuleSpec{
					ID:              "module_alpha.R015",
					Module:          "module_alpha",
					Message:         "second",
					DefaultSeverity: lint.SeverityWarning,
				},
			})
		}),
	)
	if !errors.Is(err, ErrDuplicateRuleID) {
		t.Fatalf("RegisterProviders(duplicate) error=%v, want ErrDuplicateRuleID", err)
	}
}

func TestNewEngineWithProviders(t *testing.T) {
	t.Parallel()

	engine, err := NewEngineWithProviders(
		lint.RuleProviderFunc(func(registrar lint.RuleRegistrar) error {
			return registrar.Register(testRunner{
				spec: lint.RuleSpec{
					ID:              "module_gamma.R001",
					Module:          "module_gamma",
					Message:         "provider",
					DefaultSeverity: lint.SeverityWarning,
				},
			})
		}),
	)
	if err != nil {
		t.Fatalf("NewEngineWithProviders() error: %v", err)
	}

	if engine == nil {
		t.Fatal("NewEngineWithProviders() returned nil engine")
	}

	if len(engine.Rules()) != 1 {
		t.Fatalf("len(NewEngineWithProviders().Rules())=%d, want 1", len(engine.Rules()))
	}
}

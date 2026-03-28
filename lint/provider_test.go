// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"slices"
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

// providerTestScopeAware stores scope-filter aware provider state.
type providerTestScopeAware struct {
	// scopes stores latest requested scope filters.
	scopes []string

	// runners stores runners returned by provider.
	runners []RuleRunner
}

// RegisterRules registers all runners without scope filtering.
func (provider *providerTestScopeAware) RegisterRules(
	registrar RuleRegistrar,
) error {
	return registrar.Register(provider.runners...)
}

// RegisterRulesByScope registers only runners matched by scope filter.
func (provider *providerTestScopeAware) RegisterRulesByScope(
	registrar RuleRegistrar,
	scopes ...string,
) error {
	provider.scopes = append(provider.scopes[:0], scopes...)

	allowed := make(map[string]struct{}, len(scopes))
	for scopeIndex := range scopes {
		allowed[scopes[scopeIndex]] = struct{}{}
	}

	filtered := make([]RuleRunner, 0, len(provider.runners))
	for runnerIndex := range provider.runners {
		spec := provider.runners[runnerIndex].RuleSpec()
		if _, ok := allowed[spec.Scope]; !ok {
			continue
		}

		filtered = append(filtered, provider.runners[runnerIndex])
	}

	return registrar.Register(filtered...)
}

// RegisterRulesByStage forwards stage filters to scope registration.
func (provider *providerTestScopeAware) RegisterRulesByStage(
	registrar RuleRegistrar,
	stages ...Stage,
) error {
	scopeTokens := make([]string, 0, len(stages))
	for stageIndex := range stages {
		scopeTokens = append(scopeTokens, string(stages[stageIndex]))
	}

	return provider.RegisterRulesByScope(registrar, scopeTokens...)
}

// providerTestScopeOnly stores provider state with scope-only filtering support.
type providerTestScopeOnly struct {
	// scopes stores latest requested scope filters.
	scopes []string

	// runners stores runners returned by provider.
	runners []RuleRunner
}

// RegisterRules registers all runners without scope filtering.
func (provider *providerTestScopeOnly) RegisterRules(
	registrar RuleRegistrar,
) error {
	return registrar.Register(provider.runners...)
}

// RegisterRulesByScope registers only runners matched by scope filter.
func (provider *providerTestScopeOnly) RegisterRulesByScope(
	registrar RuleRegistrar,
	scopes ...string,
) error {
	provider.scopes = append(provider.scopes[:0], scopes...)

	allowed := make(map[string]struct{}, len(scopes))
	for scopeIndex := range scopes {
		allowed[scopes[scopeIndex]] = struct{}{}
	}

	filtered := make([]RuleRunner, 0, len(provider.runners))
	for runnerIndex := range provider.runners {
		spec := provider.runners[runnerIndex].RuleSpec()
		if _, ok := allowed[spec.Scope]; !ok {
			continue
		}

		filtered = append(filtered, provider.runners[runnerIndex])
	}

	return registrar.Register(filtered...)
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

func TestRegisterRuleProvidersByScope(t *testing.T) {
	t.Parallel()

	scopeAware := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			},
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.preprocess.a002",
					Module:          "module_alpha",
					Scope:           "preprocess",
					Message:         "second",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByScope(
		&registrar,
		[]string{"parse"},
		scopeAware,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByScope() error: %v", err)
	}

	if !slices.Equal(scopeAware.scopes, []string{"parse"}) {
		t.Fatalf("scopeAware.scopes=%v, want [parse]", scopeAware.scopes)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "parse" {
		t.Fatalf("registered scope=%q, want parse", registrar.runners[0].RuleSpec().Scope)
	}
}

func TestRegisterRuleProvidersByStage(t *testing.T) {
	t.Parallel()

	scopeAware := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			},
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.preprocess.a002",
					Module:          "module_alpha",
					Scope:           "preprocess",
					Message:         "second",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByStage(
		&registrar,
		[]Stage{"preprocess"},
		scopeAware,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByStage() error: %v", err)
	}

	if !slices.Equal(scopeAware.scopes, []string{"preprocess"}) {
		t.Fatalf(
			"scopeAware.scopes=%v, want [preprocess]",
			scopeAware.scopes,
		)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "preprocess" {
		t.Fatalf(
			"registered scope=%q, want preprocess",
			registrar.runners[0].RuleSpec().Scope,
		)
	}
}

func TestRegisterRuleProvidersByStageWithModuleSpecWrapper(t *testing.T) {
	t.Parallel()

	scopeAware := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	wrapped := WithModuleSpec(
		scopeAware,
		ModuleSpec{
			ID:          "module_alpha",
			Name:        "Module Alpha",
			Description: "Rules for module_alpha.",
		},
	)

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByStage(
		&registrar,
		[]Stage{"parse"},
		wrapped,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByStage(wrapped) error: %v", err)
	}

	if len(registrar.modules) != 1 {
		t.Fatalf("registered modules=%d, want 1", len(registrar.modules))
	}

	if registrar.modules[0].ID != "module_alpha" {
		t.Fatalf("modules[0].ID=%q, want module_alpha", registrar.modules[0].ID)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}
}

func TestRegisterRuleProvidersByStageWithModuleSpecScopeFallback(t *testing.T) {
	t.Parallel()

	scopeOnly := &providerTestScopeOnly{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "first",
					DefaultSeverity: SeverityWarning,
				},
			},
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.preprocess.a002",
					Module:          "module_alpha",
					Scope:           "preprocess",
					Message:         "second",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	wrapped := WithModuleSpec(
		scopeOnly,
		ModuleSpec{
			ID:          "module_alpha",
			Name:        "Module Alpha",
			Description: "Rules for module_alpha.",
		},
	)

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByStage(
		&registrar,
		[]Stage{"parse"},
		wrapped,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByStage(scope fallback) error: %v", err)
	}

	if !slices.Equal(scopeOnly.scopes, []string{"parse"}) {
		t.Fatalf("scopeOnly.scopes=%v, want [parse]", scopeOnly.scopes)
	}

	if len(registrar.modules) != 1 {
		t.Fatalf("registered modules=%d, want 1", len(registrar.modules))
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "parse" {
		t.Fatalf("registered scope=%q, want parse", registrar.runners[0].RuleSpec().Scope)
	}
}

func TestComposeProvidersPreservesScopeFiltering(t *testing.T) {
	t.Parallel()

	left := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "left parse",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}
	right := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_beta.preprocess.b001",
					Module:          "module_beta",
					Scope:           "preprocess",
					Message:         "right preprocess",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	composed := ComposeProviders(left, right)
	if composed == nil {
		t.Fatal("ComposeProviders(left, right)=nil")
	}

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByScope(
		&registrar,
		[]string{"parse"},
		composed,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByScope(composed) error: %v", err)
	}

	if !slices.Equal(left.scopes, []string{"parse"}) {
		t.Fatalf("left.scopes=%v, want [parse]", left.scopes)
	}

	if !slices.Equal(right.scopes, []string{"parse"}) {
		t.Fatalf("right.scopes=%v, want [parse]", right.scopes)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "parse" {
		t.Fatalf("registered scope=%q, want parse", registrar.runners[0].RuleSpec().Scope)
	}
}

func TestComposeProvidersPreservesStageFiltering(t *testing.T) {
	t.Parallel()

	left := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_alpha.parse.a001",
					Module:          "module_alpha",
					Scope:           "parse",
					Message:         "left parse",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}
	right := &providerTestScopeAware{
		runners: []RuleRunner{
			providerTestRunner{
				spec: RuleSpec{
					ID:              "module_beta.preprocess.b001",
					Module:          "module_beta",
					Scope:           "preprocess",
					Message:         "right preprocess",
					DefaultSeverity: SeverityWarning,
				},
			},
		},
	}

	composed := ComposeProviders(left, right)
	if composed == nil {
		t.Fatal("ComposeProviders(left, right)=nil")
	}

	var registrar providerTestRegistrar
	err := RegisterRuleProvidersByStage(
		&registrar,
		[]Stage{"parse"},
		composed,
	)
	if err != nil {
		t.Fatalf("RegisterRuleProvidersByStage(composed) error: %v", err)
	}

	if !slices.Equal(left.scopes, []string{"parse"}) {
		t.Fatalf("left.scopes=%v, want [parse]", left.scopes)
	}

	if !slices.Equal(right.scopes, []string{"parse"}) {
		t.Fatalf("right.scopes=%v, want [parse]", right.scopes)
	}

	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if registrar.runners[0].RuleSpec().Scope != "parse" {
		t.Fatalf("registered scope=%q, want parse", registrar.runners[0].RuleSpec().Scope)
	}
}

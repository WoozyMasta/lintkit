// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
	"strings"
)

// ModuleRegistrar registers one or more lint module descriptors.
type ModuleRegistrar interface {
	// RegisterModule validates and registers one module metadata descriptor.
	RegisterModule(spec ModuleSpec) error
}

// RuleRegistrar registers one or more rule runners.
type RuleRegistrar interface {
	// Register validates and registers one or more runners.
	Register(runners ...RuleRunner) error
}

// RuleProvider registers module-owned rule runners into registrar.
type RuleProvider interface {
	// RegisterRules adds provider-owned rules to target registrar.
	RegisterRules(registrar RuleRegistrar) error
}

// ScopeRuleProvider registers provider-owned rules filtered by rule scope tokens.
type ScopeRuleProvider interface {
	// RegisterRulesByScope adds provider-owned rules for selected scopes.
	RegisterRulesByScope(registrar RuleRegistrar, scopes ...string) error
}

// StageRuleProvider registers provider-owned rules filtered by stage tokens.
type StageRuleProvider interface {
	// RegisterRulesByStage adds provider-owned rules for selected stages.
	RegisterRulesByStage(registrar RuleRegistrar, stages ...Stage) error
}

// ModuleProvider optionally provides module metadata for registration.
//
// RegisterRuleProviders will pass this metadata to registrars that implement
// ModuleRegistrar.
type ModuleProvider interface {
	// ModuleSpec returns stable provider module metadata descriptor.
	ModuleSpec() ModuleSpec
}

// RuleProviderFunc adapts function values to RuleProvider contract.
type RuleProviderFunc func(registrar RuleRegistrar) error

// RegisterRules invokes function-based provider registration.
func (fn RuleProviderFunc) RegisterRules(registrar RuleRegistrar) error {
	if fn == nil {
		return ErrNilRuleProvider
	}

	return fn(registrar)
}

// RegisterRuleProviders applies providers in declaration order.
func RegisterRuleProviders(
	registrar RuleRegistrar,
	providers ...RuleProvider,
) error {
	if registrar == nil {
		return ErrNilRuleRegistrar
	}

	for index := range providers {
		if providers[index] == nil {
			return fmt.Errorf("%w at index %d", ErrNilRuleProvider, index)
		}

		if err := registerProviderModuleMetadata(registrar, providers[index], index); err != nil {
			return err
		}

		if err := providers[index].RegisterRules(registrar); err != nil {
			return fmt.Errorf("provider[%d]: %w", index, err)
		}
	}

	return nil
}

// RegisterRuleProvidersByScope applies providers with optional scope filtering.
//
// Providers that implement ScopeRuleProvider receive filtered registration call.
// Other providers fall back to full RegisterRules behavior.
func RegisterRuleProvidersByScope(
	registrar RuleRegistrar,
	scopes []string,
	providers ...RuleProvider,
) error {
	if registrar == nil {
		return ErrNilRuleRegistrar
	}

	normalizedScopes := normalizeScopeFilters(scopes)

	for index := range providers {
		if providers[index] == nil {
			return fmt.Errorf("%w at index %d", ErrNilRuleProvider, index)
		}

		if err := registerProviderModuleMetadata(registrar, providers[index], index); err != nil {
			return err
		}

		if scopeProvider, ok := providers[index].(ScopeRuleProvider); ok {
			if err := scopeProvider.RegisterRulesByScope(
				registrar,
				normalizedScopes...,
			); err != nil {
				return fmt.Errorf("provider[%d] filtered scope registration: %w", index, err)
			}

			continue
		}

		if err := providers[index].RegisterRules(registrar); err != nil {
			return fmt.Errorf("provider[%d]: %w", index, err)
		}
	}

	return nil
}

// RegisterRuleProvidersByStage applies providers with optional stage filtering.
//
// Providers that implement StageRuleProvider receive filtered registration call.
// If provider has only ScopeRuleProvider, stage values are passed as scope tokens.
// Other providers fall back to full RegisterRules behavior.
func RegisterRuleProvidersByStage(
	registrar RuleRegistrar,
	stages []Stage,
	providers ...RuleProvider,
) error {
	if registrar == nil {
		return ErrNilRuleRegistrar
	}

	normalizedStages := normalizeStageFilters(stages)

	for index := range providers {
		if providers[index] == nil {
			return fmt.Errorf("%w at index %d", ErrNilRuleProvider, index)
		}

		if err := registerProviderModuleMetadata(registrar, providers[index], index); err != nil {
			return err
		}

		if stageProvider, ok := providers[index].(StageRuleProvider); ok {
			if err := stageProvider.RegisterRulesByStage(
				registrar,
				normalizedStages...,
			); err != nil {
				return fmt.Errorf("provider[%d] filtered stage registration: %w", index, err)
			}

			continue
		}

		if scopeProvider, ok := providers[index].(ScopeRuleProvider); ok {
			if err := scopeProvider.RegisterRulesByScope(
				registrar,
				stagesToScopeTokens(normalizedStages)...,
			); err != nil {
				return fmt.Errorf("provider[%d] filtered stage->scope registration: %w", index, err)
			}

			continue
		}

		if err := providers[index].RegisterRules(registrar); err != nil {
			return fmt.Errorf("provider[%d]: %w", index, err)
		}
	}

	return nil
}

// ComposeProviders combines multiple providers into one provider.
//
// Returned provider preserves declaration order and forwards optional module
// metadata from nested providers when registrar supports ModuleRegistrar.
// Returned provider also preserves optional scope and stage filter interfaces.
//
// Behavior:
//   - no input providers -> nil
//   - one provider -> same provider
//   - multiple providers -> composite provider
func ComposeProviders(providers ...RuleProvider) RuleProvider {
	if len(providers) == 0 {
		return nil
	}

	if len(providers) == 1 {
		return providers[0]
	}

	copied := make([]RuleProvider, len(providers))
	copy(copied, providers)

	return composedProvider{
		providers: copied,
	}
}

// composedProvider composes rule providers while preserving filter capabilities.
type composedProvider struct {
	// providers stores ordered child providers.
	providers []RuleProvider
}

// RegisterRules applies all child providers in declaration order.
func (provider composedProvider) RegisterRules(registrar RuleRegistrar) error {
	return RegisterRuleProviders(registrar, provider.providers...)
}

// RegisterRulesByScope applies all child providers with scope filtering.
func (provider composedProvider) RegisterRulesByScope(
	registrar RuleRegistrar,
	scopes ...string,
) error {
	return RegisterRuleProvidersByScope(registrar, scopes, provider.providers...)
}

// RegisterRulesByStage applies all child providers with stage filtering.
func (provider composedProvider) RegisterRulesByStage(
	registrar RuleRegistrar,
	stages ...Stage,
) error {
	return RegisterRuleProvidersByStage(registrar, stages, provider.providers...)
}

// providerWithModuleSpec wraps provider and attaches module metadata.
type providerWithModuleSpec struct {
	// provider stores wrapped rule provider.
	provider RuleProvider

	// module stores attached module metadata descriptor.
	module ModuleSpec
}

// RegisterRules delegates runner registration to wrapped provider.
func (provider providerWithModuleSpec) RegisterRules(registrar RuleRegistrar) error {
	if provider.provider == nil {
		return ErrNilRuleProvider
	}

	return provider.provider.RegisterRules(registrar)
}

// RegisterRulesByScope delegates filtered scope registration to wrapped provider.
func (provider providerWithModuleSpec) RegisterRulesByScope(
	registrar RuleRegistrar,
	scopes ...string,
) error {
	if provider.provider == nil {
		return ErrNilRuleProvider
	}

	scopeProvider, ok := provider.provider.(ScopeRuleProvider)
	if !ok {
		return provider.provider.RegisterRules(registrar)
	}

	return scopeProvider.RegisterRulesByScope(registrar, scopes...)
}

// RegisterRulesByStage delegates filtered stage registration to wrapped provider.
func (provider providerWithModuleSpec) RegisterRulesByStage(
	registrar RuleRegistrar,
	stages ...Stage,
) error {
	if provider.provider == nil {
		return ErrNilRuleProvider
	}

	stageProvider, ok := provider.provider.(StageRuleProvider)
	if ok {
		return stageProvider.RegisterRulesByStage(registrar, stages...)
	}

	scopeProvider, ok := provider.provider.(ScopeRuleProvider)
	if ok {
		return scopeProvider.RegisterRulesByScope(
			registrar,
			stagesToScopeTokens(stages)...,
		)
	}

	return provider.provider.RegisterRules(registrar)
}

// ModuleSpec returns attached module metadata descriptor.
func (provider providerWithModuleSpec) ModuleSpec() ModuleSpec {
	return provider.module
}

// WithModuleSpec attaches module metadata to provider registration flow.
func WithModuleSpec(provider RuleProvider, module ModuleSpec) RuleProvider {
	if provider == nil {
		return nil
	}

	return providerWithModuleSpec{
		provider: provider,
		module:   module,
	}
}

// registerProviderModuleMetadata attaches module metadata when registrar supports it.
func registerProviderModuleMetadata(
	registrar RuleRegistrar,
	provider RuleProvider,
	index int,
) error {
	moduleProvider, ok := provider.(ModuleProvider)
	if !ok {
		return nil
	}

	moduleRegistrar, ok := registrar.(ModuleRegistrar)
	if !ok {
		return nil
	}

	if err := moduleRegistrar.RegisterModule(moduleProvider.ModuleSpec()); err != nil {
		return fmt.Errorf("provider[%d] module metadata: %w", index, err)
	}

	return nil
}

// normalizeScopeFilters returns unique non-empty scope tokens.
func normalizeScopeFilters(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for index := range scopes {
		scope := strings.TrimSpace(scopes[index])
		if scope == "" {
			continue
		}

		if _, ok := seen[scope]; ok {
			continue
		}

		seen[scope] = struct{}{}
		out = append(out, scope)
	}

	return out
}

// normalizeStageFilters returns unique non-empty stage tokens.
func normalizeStageFilters(stages []Stage) []Stage {
	if len(stages) == 0 {
		return nil
	}

	seen := make(map[Stage]struct{}, len(stages))
	out := make([]Stage, 0, len(stages))
	for index := range stages {
		stage := Stage(strings.TrimSpace(string(stages[index])))
		if stage == "" {
			continue
		}

		if _, ok := seen[stage]; ok {
			continue
		}

		seen[stage] = struct{}{}
		out = append(out, stage)
	}

	return out
}

// stagesToScopeTokens converts stages to scope tokens.
func stagesToScopeTokens(stages []Stage) []string {
	if len(stages) == 0 {
		return nil
	}

	out := make([]string, 0, len(stages))
	for index := range stages {
		out = append(out, string(stages[index]))
	}

	return out
}

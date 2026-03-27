// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
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

		if moduleProvider, ok := providers[index].(ModuleProvider); ok {
			if moduleRegistrar, ok := registrar.(ModuleRegistrar); ok {
				if err := moduleRegistrar.RegisterModule(moduleProvider.ModuleSpec()); err != nil {
					return fmt.Errorf("provider[%d] module metadata: %w", index, err)
				}
			}
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

	return RuleProviderFunc(func(registrar RuleRegistrar) error {
		return RegisterRuleProviders(registrar, copied...)
	})
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

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
	"strings"
)

const (
	// UnknownCodeKeep keeps diagnostics with unknown code in run index.
	UnknownCodeKeep UnknownCodePolicy = "keep"

	// UnknownCodeDrop drops diagnostics with unknown code from run index.
	UnknownCodeDrop UnknownCodePolicy = "drop"
)

// UnknownCodePolicy controls how unknown catalog codes are handled.
type UnknownCodePolicy string

// CodeCatalogBindingConfig stores code-catalog binding setup.
type CodeCatalogBindingConfig[ItemDiagnostic any] struct {

	// CodeFromDiagnostic extracts code key from one diagnostic item.
	CodeFromDiagnostic func(diagnostic ItemDiagnostic) Code

	// DiagnosticToLint converts one module diagnostic into shared lint diagnostic.
	DiagnosticToLint func(diagnostic ItemDiagnostic) Diagnostic

	// RunValueKey stores run context key for grouped diagnostics map.
	RunValueKey string

	// UnknownCodePolicy controls unknown code handling; default is UnknownCodeKeep.
	UnknownCodePolicy UnknownCodePolicy

	// Catalog stores initialized code catalog helper.
	Catalog CodeCatalog
}

// CodeCatalogBinding stores reusable provider+attach wiring for one code catalog.
type CodeCatalogBinding[ItemDiagnostic any] struct {

	// codeFromDiagnostic extracts code key from one diagnostic item.
	codeFromDiagnostic func(diagnostic ItemDiagnostic) Code

	// runValueKey stores run context key for grouped diagnostics map.
	runValueKey string

	// unknownCodePolicy stores unknown code handling mode.
	unknownCodePolicy UnknownCodePolicy

	// catalog stores initialized code catalog helper.
	catalog CodeCatalog

	// provider stores prebuilt catalog-backed rule provider.
	provider CatalogProvider[CodeSpec, ItemDiagnostic, Code]
}

// NewCodeCatalogBinding builds reusable code-catalog rule binding.
func NewCodeCatalogBinding[ItemDiagnostic any](
	config CodeCatalogBindingConfig[ItemDiagnostic],
) (CodeCatalogBinding[ItemDiagnostic], error) {
	normalizedKey := strings.TrimSpace(config.RunValueKey)
	if !ValidateRunValueKey(normalizedKey) {
		return CodeCatalogBinding[ItemDiagnostic]{}, fmt.Errorf(
			"%w: invalid run value key",
			ErrInvalidCatalogProvider,
		)
	}

	if config.CodeFromDiagnostic == nil {
		return CodeCatalogBinding[ItemDiagnostic]{}, fmt.Errorf(
			"%w: nil diagnostic code mapper",
			ErrInvalidCatalogProvider,
		)
	}

	if config.DiagnosticToLint == nil {
		return CodeCatalogBinding[ItemDiagnostic]{}, fmt.Errorf(
			"%w: nil diagnostic converter",
			ErrInvalidCatalogProvider,
		)
	}

	policy := normalizeUnknownCodePolicy(config.UnknownCodePolicy)

	provider, err := NewCodeCatalogProvider(
		normalizedKey,
		config.Catalog,
		config.DiagnosticToLint,
	)
	if err != nil {
		return CodeCatalogBinding[ItemDiagnostic]{}, err
	}

	return CodeCatalogBinding[ItemDiagnostic]{
		runValueKey:        normalizedKey,
		catalog:            config.Catalog,
		codeFromDiagnostic: config.CodeFromDiagnostic,
		provider:           provider,
		unknownCodePolicy:  policy,
	}, nil
}

// RegisterRules registers binding rule runners into registrar.
func (binding CodeCatalogBinding[ItemDiagnostic]) RegisterRules(
	registrar RuleRegistrar,
) error {
	return binding.provider.RegisterRules(registrar)
}

// RegisterRulesByScope registers binding rules filtered by scope tokens.
func (binding CodeCatalogBinding[ItemDiagnostic]) RegisterRulesByScope(
	registrar RuleRegistrar,
	scopes ...string,
) error {
	return binding.provider.RegisterRulesByScope(registrar, scopes...)
}

// RegisterRulesByStage registers binding rules filtered by stage tokens.
func (binding CodeCatalogBinding[ItemDiagnostic]) RegisterRulesByStage(
	registrar RuleRegistrar,
	stages ...Stage,
) error {
	return binding.provider.RegisterRulesByStage(registrar, stages...)
}

// ModuleSpec returns inferred module metadata for registered rules.
func (binding CodeCatalogBinding[ItemDiagnostic]) ModuleSpec() ModuleSpec {
	return binding.provider.ModuleSpec()
}

// Attach indexes diagnostics in run context by code with configured policy.
func (binding CodeCatalogBinding[ItemDiagnostic]) Attach(
	run *RunContext,
	diagnostics []ItemDiagnostic,
) bool {
	items := diagnostics
	if binding.unknownCodePolicy == UnknownCodeDrop {
		items = filterKnownCatalogDiagnostics(
			binding.catalog,
			diagnostics,
			binding.codeFromDiagnostic,
		)
	}

	return AttachCatalogDiagnostics(
		run,
		binding.runValueKey,
		items,
		binding.codeFromDiagnostic,
	)
}

// filterKnownCatalogDiagnostics keeps diagnostics with catalog-known codes only.
func filterKnownCatalogDiagnostics[ItemDiagnostic any](
	catalog CodeCatalog,
	diagnostics []ItemDiagnostic,
	codeFromDiagnostic func(diagnostic ItemDiagnostic) Code,
) []ItemDiagnostic {
	if len(diagnostics) == 0 {
		return nil
	}

	filtered := make([]ItemDiagnostic, 0, len(diagnostics))
	for index := range diagnostics {
		code := codeFromDiagnostic(diagnostics[index])
		if _, ok := catalog.ByCode(code); !ok {
			continue
		}

		filtered = append(filtered, diagnostics[index])
	}

	return filtered
}

// normalizeUnknownCodePolicy normalizes unknown code policy token.
func normalizeUnknownCodePolicy(policy UnknownCodePolicy) UnknownCodePolicy {
	switch policy {
	case UnknownCodeDrop:
		return UnknownCodeDrop
	case UnknownCodeKeep:
		return UnknownCodeKeep
	default:
		return UnknownCodeKeep
	}
}

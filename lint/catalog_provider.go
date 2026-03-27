// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"fmt"
	"strings"
)

// CatalogProvider registers generic catalog-backed rule runners.
type CatalogProvider[Catalog any, ItemDiagnostic any, Code comparable] struct {
	// module stores optional inferred module metadata descriptor.
	module *ModuleSpec

	// runners stores prebuilt runners mapped from catalog entries.
	runners []RuleRunner
}

// NewCatalogProvider builds generic catalog-backed provider.
func NewCatalogProvider[Catalog any, ItemDiagnostic any, Code comparable](
	runValueKey string,
	catalog []Catalog,
	ruleSpecFromCatalog func(item Catalog) RuleSpec,
	codeFromCatalog func(item Catalog) Code,
	convertDiagnostic func(diagnostic ItemDiagnostic) Diagnostic,
) (CatalogProvider[Catalog, ItemDiagnostic, Code], error) {
	normalizedKey := strings.TrimSpace(runValueKey)
	if normalizedKey == "" {
		return CatalogProvider[Catalog, ItemDiagnostic, Code]{}, fmt.Errorf(
			"%w: empty run value key",
			ErrInvalidCatalogProvider,
		)
	}

	if ruleSpecFromCatalog == nil {
		return CatalogProvider[Catalog, ItemDiagnostic, Code]{}, fmt.Errorf(
			"%w: nil rule spec mapper",
			ErrInvalidCatalogProvider,
		)
	}

	if codeFromCatalog == nil {
		return CatalogProvider[Catalog, ItemDiagnostic, Code]{}, fmt.Errorf(
			"%w: nil catalog code mapper",
			ErrInvalidCatalogProvider,
		)
	}

	if convertDiagnostic == nil {
		return CatalogProvider[Catalog, ItemDiagnostic, Code]{}, fmt.Errorf(
			"%w: nil diagnostic converter",
			ErrInvalidCatalogProvider,
		)
	}

	runners := make([]RuleRunner, 0, len(catalog))
	moduleSpec := ModuleSpec{}
	runValueKeyPtr := &normalizedKey
	for itemIndex := range catalog {
		item := catalog[itemIndex]
		spec := ruleSpecFromCatalog(item)
		moduleSpec = mergeInferredModuleSpec(moduleSpec, spec)
		runners = append(runners, catalogRuleRunner[ItemDiagnostic, Code]{
			code:              codeFromCatalog(item),
			spec:              &spec,
			runValueKey:       runValueKeyPtr,
			convertDiagnostic: convertDiagnostic,
		})
	}

	provider := CatalogProvider[Catalog, ItemDiagnostic, Code]{
		runners: runners,
	}
	if strings.TrimSpace(moduleSpec.ID) != "" {
		provider.module = &moduleSpec
	}

	return provider, nil
}

// RegisterRules adds prebuilt catalog runners into target registrar.
func (provider CatalogProvider[Catalog, ItemDiagnostic, Code]) RegisterRules(
	registrar RuleRegistrar,
) error {
	if registrar == nil {
		return ErrNilRuleRegistrar
	}

	if moduleRegistrar, ok := registrar.(ModuleRegistrar); ok {
		module := provider.ModuleSpec()
		if strings.TrimSpace(module.ID) != "" {
			if err := moduleRegistrar.RegisterModule(module); err != nil {
				return err
			}
		}
	}

	if len(provider.runners) == 0 {
		return nil
	}

	return registrar.Register(provider.runners...)
}

// ModuleSpec returns inferred or attached provider module metadata.
func (provider CatalogProvider[Catalog, ItemDiagnostic, Code]) ModuleSpec() ModuleSpec {
	if provider.module == nil {
		return ModuleSpec{}
	}

	return *provider.module
}

// AttachCatalogDiagnostics indexes diagnostics and stores grouped map in context.
func AttachCatalogDiagnostics[ItemDiagnostic any, Code comparable](
	run *RunContext,
	runValueKey string,
	diagnostics []ItemDiagnostic,
	codeFromDiagnostic func(diagnostic ItemDiagnostic) Code,
) bool {
	if codeFromDiagnostic == nil {
		return false
	}

	return SetIndexedByCode(run, runValueKey, diagnostics, codeFromDiagnostic)
}

// catalogRuleRunner emits diagnostics matched by one catalog code.
type catalogRuleRunner[ItemDiagnostic any, Code comparable] struct {
	// spec stores stable rule metadata for current runner.
	spec *RuleSpec

	// runValueKey stores key with grouped diagnostics map.
	runValueKey *string

	// convertDiagnostic converts one module diagnostic into shared model.
	convertDiagnostic func(diagnostic ItemDiagnostic) Diagnostic

	// code stores current catalog row code.
	code Code
}

// RuleSpec returns one stable metadata descriptor for current rule.
func (runner catalogRuleRunner[ItemDiagnostic, Code]) RuleSpec() RuleSpec {
	if runner.spec == nil {
		return RuleSpec{}
	}

	return *runner.spec
}

// Check emits diagnostics for current catalog code from run context index.
func (runner catalogRuleRunner[ItemDiagnostic, Code]) Check(
	_ context.Context,
	run *RunContext,
	emit DiagnosticEmit,
) error {
	diagnosticsByCode, ok := GetIndexedByCode[ItemDiagnostic, Code](
		run,
		derefString(runner.runValueKey),
	)
	if !ok || len(diagnosticsByCode) == 0 {
		return nil
	}

	diagnostics := diagnosticsByCode[runner.code]
	for itemIndex := range diagnostics {
		converted := runner.convertDiagnostic(diagnostics[itemIndex])
		if runner.spec != nil {
			converted.RuleID = runner.spec.ID
		}

		emit(converted)
	}

	return nil
}

// derefString returns string pointer value or empty string for nil pointers.
func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

// mergeInferredModuleSpec updates inferred module metadata from one rule spec.
func mergeInferredModuleSpec(current ModuleSpec, spec RuleSpec) ModuleSpec {
	if strings.TrimSpace(current.ID) != "" {
		return current
	}

	moduleID := strings.TrimSpace(spec.Module)
	if moduleID == "" {
		return current
	}

	return ModuleSpec{ID: moduleID}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"fmt"
	"strings"
)

// CodeCatalogConfig stores code catalog behavior settings.
type CodeCatalogConfig struct {
	// ScopeDescriptions stores required scope documentation by scope token.
	ScopeDescriptions map[Stage]string

	// Module is required stable module namespace for rule IDs.
	Module string

	// CodePrefix is required stable exported code prefix.
	// Example: "RVCFG" => exported codes "RVCFG2001", "RVCFG3017".
	CodePrefix string

	// ModuleName is optional human-readable module display name.
	ModuleName string

	// ModuleDescription is optional module-level documentation text.
	ModuleDescription string
}

// CodeCatalog stores module-local stable code metadata helpers.
type CodeCatalog struct {
	// byCode stores code metadata rows by normalized code.
	byCode map[Code]CodeSpec

	// codePrefix stores optional exported code prefix.
	codePrefix string

	// moduleDescription stores optional module-level documentation text.
	moduleDescription string

	// moduleName stores optional module display name.
	moduleName string

	// module stores stable module namespace for rule IDs.
	module string

	// scopeDescriptions stores normalized scope descriptions by scope token.
	scopeDescriptions map[Stage]string

	// specs stores stable catalog rows in source order.
	specs []CodeSpec
}

// NewCodeCatalog builds reusable catalog helper from config and code metadata.
func NewCodeCatalog(
	config CodeCatalogConfig,
	specs []CodeSpec,
) (CodeCatalog, error) {
	normalizedConfig, err := normalizeCodeCatalogConfig(config)
	if err != nil {
		return CodeCatalog{}, err
	}

	copiedSpecs := make([]CodeSpec, len(specs))
	copy(copiedSpecs, specs)

	byCode := make(map[Code]CodeSpec, len(copiedSpecs))

	for index := range copiedSpecs {
		code := copiedSpecs[index].Code
		if code == 0 {
			continue
		}

		byCode[code] = copiedSpecs[index]
	}

	if err := validateCodeCatalogSpecs(normalizedConfig, copiedSpecs); err != nil {
		return CodeCatalog{}, err
	}

	return CodeCatalog{
		byCode:            byCode,
		codePrefix:        normalizedConfig.CodePrefix,
		module:            normalizedConfig.Module,
		moduleName:        normalizedConfig.ModuleName,
		moduleDescription: normalizedConfig.ModuleDescription,
		scopeDescriptions: normalizedConfig.ScopeDescriptions,
		specs:             copiedSpecs,
	}, nil
}

// CodeSpecs returns stable code metadata rows copy.
func (catalog CodeCatalog) CodeSpecs() []CodeSpec {
	out := make([]CodeSpec, len(catalog.specs))
	copy(out, catalog.specs)

	return out
}

// ByCode returns code metadata by normalized stable code token.
func (catalog CodeCatalog) ByCode(code Code) (CodeSpec, bool) {
	if code == 0 {
		return CodeSpec{}, false
	}

	spec, ok := catalog.byCode[code]
	return spec, ok
}

// RuleID returns stable global rule ID mapped from module-local code.
func (catalog CodeCatalog) RuleID(code Code) (string, error) {
	if code == 0 {
		return "", fmt.Errorf("%w: %d", ErrUnknownCodeCatalogCode, code)
	}

	spec, ok := catalog.byCode[code]
	if !ok {
		return "", fmt.Errorf("%w: %d", ErrUnknownCodeCatalogCode, code)
	}

	return BuildRuleID(
		catalog.module,
		spec.Stage,
		spec.Message,
		code,
	), nil
}

// RuleSpec converts one code metadata row into lint rule metadata.
func (catalog CodeCatalog) RuleSpec(spec CodeSpec) RuleSpec {
	ruleID := BuildRuleID(
		catalog.module,
		spec.Stage,
		spec.Message,
		spec.Code,
	)

	base := RuleSpec{
		ID:               ruleID,
		Module:           catalog.module,
		Scope:            strings.TrimSpace(string(spec.Stage)),
		ScopeDescription: catalog.scopeDescription(spec.Stage),
		Code:             catalog.PublicCode(spec.Code),
		Message:          strings.TrimSpace(spec.Message),
		Description:      spec.Description,
		DefaultSeverity:  spec.Severity,
		DefaultEnabled:   cloneBool(spec.Enabled),
		DefaultOptions:   spec.Options,
	}

	return mergeCodeRuleOverride(base, spec.Rule)
}

// ModuleSpec returns catalog module metadata descriptor.
func (catalog CodeCatalog) ModuleSpec() ModuleSpec {
	return ModuleSpec{
		ID:          catalog.module,
		Name:        catalog.moduleName,
		Description: catalog.moduleDescription,
	}
}

// PublicCode returns exported code token with optional catalog prefix.
func (catalog CodeCatalog) PublicCode(code Code) string {
	return applyCodePrefix(catalog.codePrefix, code)
}

// RuleSpecs returns deterministic lint rule specs from all catalog rows.
func (catalog CodeCatalog) RuleSpecs() []RuleSpec {
	specs := make([]RuleSpec, 0, len(catalog.specs))

	for index := range catalog.specs {
		specs = append(specs, catalog.RuleSpec(catalog.specs[index]))
	}

	return specs
}

// NewCodeCatalogProvider builds catalog provider from CodeCatalog helper.
func NewCodeCatalogProvider[ItemDiagnostic any](
	runValueKey string,
	catalog CodeCatalog,
	diagnosticToLint func(item ItemDiagnostic) Diagnostic,
) (CatalogProvider[CodeSpec, ItemDiagnostic, Code], error) {
	provider, err := NewCatalogProvider(
		runValueKey,
		catalog.CodeSpecs(),
		catalog.RuleSpec,
		func(spec CodeSpec) Code {
			return spec.Code
		},
		diagnosticToLint,
	)
	if err != nil {
		return CatalogProvider[CodeSpec, ItemDiagnostic, Code]{}, err
	}

	module := catalog.ModuleSpec()
	if strings.TrimSpace(module.ID) != "" {
		provider.module = &module
	}

	return provider, nil
}

// cloneBool returns detached bool pointer copy.
func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}

	out := *value
	return &out
}

// mergeCodeRuleOverride merges narrow catalog override fields into rule spec.
func mergeCodeRuleOverride(base RuleSpec, override *CodeRuleOverride) RuleSpec {
	if override == nil {
		return base
	}

	out := base
	item := *override

	if strings.TrimSpace(item.ID) != "" {
		out.ID = strings.TrimSpace(item.ID)
	}

	if strings.TrimSpace(item.Scope) != "" {
		out.Scope = strings.TrimSpace(item.Scope)
	}

	if strings.TrimSpace(item.ScopeDescription) != "" {
		out.ScopeDescription = strings.TrimSpace(item.ScopeDescription)
	}

	if strings.TrimSpace(item.Code) != "" {
		out.Code = strings.TrimSpace(item.Code)
	}

	if len(item.FileKinds) > 0 {
		out.FileKinds = NormalizeFileKinds(item.FileKinds)
	}

	if item.Deprecated {
		out.Deprecated = true
	}

	return out
}

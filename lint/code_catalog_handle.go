// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"strings"
	"sync"
)

// CodeCatalogHandle stores lazy code catalog initialization state.
type CodeCatalogHandle struct {

	// err stores catalog construction error.
	err error
	// config stores source catalog configuration.
	config CodeCatalogConfig

	// catalog stores constructed helper.
	catalog CodeCatalog

	// specs stores source catalog specs.
	specs []CodeSpec

	// once guards one-time catalog construction.
	once sync.Once
}

// NewCodeCatalogHandle builds lazy code catalog handle.
func NewCodeCatalogHandle(
	config CodeCatalogConfig,
	specs []CodeSpec,
) *CodeCatalogHandle {
	copiedSpecs := make([]CodeSpec, len(specs))
	copy(copiedSpecs, specs)

	return &CodeCatalogHandle{
		config: config,
		specs:  copiedSpecs,
	}
}

// Catalog returns initialized code catalog helper.
func (handle *CodeCatalogHandle) Catalog() (CodeCatalog, error) {
	if handle == nil {
		return CodeCatalog{}, ErrNilCodeCatalogHandle
	}

	handle.once.Do(func() {
		handle.catalog, handle.err = NewCodeCatalog(handle.config, handle.specs)
	})

	if handle.err != nil {
		return CodeCatalog{}, handle.err
	}

	return handle.catalog, nil
}

// RuleSpec converts one diagnostic spec into lint rule metadata.
func (handle *CodeCatalogHandle) RuleSpec(spec CodeSpec) (RuleSpec, error) {
	catalog, err := handle.Catalog()
	if err != nil {
		return RuleSpec{}, err
	}

	return catalog.RuleSpec(spec), nil
}

// RuleID returns stable rule id for code from initialized catalog.
func (handle *CodeCatalogHandle) RuleID(code Code) (string, error) {
	catalog, err := handle.Catalog()
	if err != nil {
		return "", err
	}

	return catalog.RuleID(code)
}

// RuleIDOrUnknown returns stable rule id or "<module>.unknown" fallback.
func (handle *CodeCatalogHandle) RuleIDOrUnknown(code Code) string {
	module := strings.TrimSpace(handle.module())
	if module == "" {
		module = "unknown"
	}

	ruleID, err := handle.RuleID(code)
	if err != nil {
		return module + ".unknown"
	}

	return ruleID
}

// CodeSpecs returns catalog code specs copy.
func (handle *CodeCatalogHandle) CodeSpecs() []CodeSpec {
	catalog, err := handle.Catalog()
	if err != nil {
		return nil
	}

	return catalog.CodeSpecs()
}

// ByCode returns catalog code spec for code token.
func (handle *CodeCatalogHandle) ByCode(code Code) (CodeSpec, bool) {
	catalog, err := handle.Catalog()
	if err != nil {
		return CodeSpec{}, false
	}

	return catalog.ByCode(code)
}

// RuleSpecs returns deterministic lint rule specs from catalog rows.
func (handle *CodeCatalogHandle) RuleSpecs() []RuleSpec {
	catalog, err := handle.Catalog()
	if err != nil {
		return nil
	}

	return catalog.RuleSpecs()
}

// ModuleSpec returns catalog module metadata descriptor.
func (handle *CodeCatalogHandle) ModuleSpec() ModuleSpec {
	catalog, err := handle.Catalog()
	if err != nil {
		moduleID := strings.TrimSpace(handle.module())
		if moduleID == "" {
			return ModuleSpec{}
		}

		return ModuleSpec{ID: moduleID}
	}

	return catalog.ModuleSpec()
}

// module returns configured module id from source config.
func (handle *CodeCatalogHandle) module() string {
	if handle == nil {
		return ""
	}

	return strings.TrimSpace(handle.config.Module)
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/woozymasta/lintkit/lint"
)

// Registry stores registered lint rules and provides deterministic lookups.
type Registry struct {
	// byID stores rule metadata by stable rule ID.
	byID map[string]lint.RuleSpec

	// codeOwners stores non-empty rule codes and owning rule IDs.
	codeOwners map[string]string

	// byModule stores normalized module metadata by stable module ID.
	byModule map[string]lint.ModuleSpec

	// cachedRules stores cached sorted rule list for read-heavy queries.
	cachedRules []lint.RuleSpec

	// cachedModules stores cached sorted module list for read-heavy queries.
	cachedModules []lint.ModuleSpec

	// cachedSnapshot stores cached deterministic snapshot payload.
	cachedSnapshot lint.RegistrySnapshot

	// hasCachedSnapshot reports whether cachedSnapshot is initialized.
	hasCachedSnapshot bool

	// mu guards concurrent registry reads and writes.
	mu sync.RWMutex
}

// NewRegistry creates an empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		byID:       make(map[string]lint.RuleSpec),
		codeOwners: make(map[string]string),
		byModule:   make(map[string]lint.ModuleSpec),
	}
}

// Register validates and registers one rule metadata descriptor.
func (registry *Registry) Register(spec lint.RuleSpec) error {
	if registry == nil {
		return ErrNilRegistry
	}

	normalized, err := normalizeRuleSpec(spec)
	if err != nil {
		return err
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.byID[normalized.ID]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateRuleID, normalized.ID)
	}

	if normalized.Code != "" {
		if owner, exists := registry.codeOwners[normalized.Code]; exists {
			return fmt.Errorf(
				"%w: %q used by %q and %q",
				ErrDuplicateRuleCode,
				normalized.Code,
				owner,
				normalized.ID,
			)
		}
	}

	if err := registry.registerModuleLocked(lint.ModuleSpec{
		ID: normalized.Module,
	}); err != nil {
		return err
	}

	registry.byID[normalized.ID] = normalized
	if normalized.Code != "" {
		registry.codeOwners[normalized.Code] = normalized.ID
	}
	registry.invalidateCachesLocked()

	return nil
}

// RegisterMany validates and registers multiple rule specs as one batch.
func (registry *Registry) RegisterMany(specs ...lint.RuleSpec) error {
	if registry == nil {
		return ErrNilRegistry
	}

	if len(specs) == 0 {
		return nil
	}

	normalized := make([]lint.RuleSpec, 0, len(specs))
	seen := make(map[string]struct{}, len(specs))
	seenCodes := make(map[string]string, len(specs))
	for index := range specs {
		item, err := normalizeRuleSpec(specs[index])
		if err != nil {
			return fmt.Errorf("spec[%d]: %w", index, err)
		}

		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRuleID, item.ID)
		}

		if item.Code != "" {
			if ownerID, exists := seenCodes[item.Code]; exists {
				return fmt.Errorf(
					"%w: %q used by %q and %q",
					ErrDuplicateRuleCode,
					item.Code,
					ownerID,
					item.ID,
				)
			}

			seenCodes[item.Code] = item.ID
		}

		seen[item.ID] = struct{}{}
		normalized = append(normalized, item)
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	for index := range normalized {
		if _, exists := registry.byID[normalized[index].ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRuleID, normalized[index].ID)
		}

		if normalized[index].Code != "" {
			if ownerID, exists := registry.codeOwners[normalized[index].Code]; exists {
				return fmt.Errorf(
					"%w: %q used by %q and %q",
					ErrDuplicateRuleCode,
					normalized[index].Code,
					ownerID,
					normalized[index].ID,
				)
			}
		}
	}

	for index := range normalized {
		registry.byID[normalized[index].ID] = normalized[index]
		if normalized[index].Code != "" {
			registry.codeOwners[normalized[index].Code] = normalized[index].ID
		}

		if err := registry.registerModuleLocked(lint.ModuleSpec{
			ID: normalized[index].Module,
		}); err != nil {
			return err
		}
	}
	registry.invalidateCachesLocked()

	return nil
}

// RegisterModule validates and registers one lint module metadata descriptor.
func (registry *Registry) RegisterModule(spec lint.ModuleSpec) error {
	if registry == nil {
		return ErrNilRegistry
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	return registry.registerModuleLocked(spec)
}

// registerModuleLocked validates and upserts module metadata.
//
// Caller must hold registry.mu write lock.
func (registry *Registry) registerModuleLocked(spec lint.ModuleSpec) error {
	normalized, err := normalizeModuleSpec(spec)
	if err != nil {
		return err
	}

	existing, exists := registry.byModule[normalized.ID]
	if !exists {
		registry.byModule[normalized.ID] = normalized
		registry.invalidateCachesLocked()
		return nil
	}

	merged, err := mergeModuleSpecs(existing, normalized)
	if err != nil {
		return err
	}

	registry.byModule[normalized.ID] = merged
	registry.invalidateCachesLocked()
	return nil
}

// Module returns one registered module metadata descriptor by ID.
func (registry *Registry) Module(id string) (lint.ModuleSpec, bool) {
	if registry == nil {
		return lint.ModuleSpec{}, false
	}

	key := strings.TrimSpace(id)
	if key == "" {
		return lint.ModuleSpec{}, false
	}

	registry.mu.RLock()
	defer registry.mu.RUnlock()

	spec, ok := registry.byModule[key]
	return spec, ok
}

// Modules returns deterministic sorted copy of registered module metadata.
func (registry *Registry) Modules() []lint.ModuleSpec {
	if registry == nil {
		return nil
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.cachedModules == nil {
		cached := make([]lint.ModuleSpec, 0, len(registry.byModule))
		for _, spec := range registry.byModule {
			cached = append(cached, spec)
		}

		sortModuleSpecs(cached)
		registry.cachedModules = cached
	}

	return cloneModuleSpecs(registry.cachedModules)
}

// Rule returns one registered rule by stable rule ID.
func (registry *Registry) Rule(id string) (lint.RuleSpec, bool) {
	if registry == nil {
		return lint.RuleSpec{}, false
	}

	key := strings.TrimSpace(id)
	if key == "" {
		return lint.RuleSpec{}, false
	}

	registry.mu.RLock()
	defer registry.mu.RUnlock()

	spec, ok := registry.byID[key]
	if !ok {
		return lint.RuleSpec{}, false
	}

	return cloneRuleSpec(spec), true
}

// Rules returns deterministic sorted copy of all registered rule specs.
func (registry *Registry) Rules() []lint.RuleSpec {
	return cloneRuleSpecs(registry.rulesView())
}

// rulesView returns internal cached sorted rule specs for read-only usage.
func (registry *Registry) rulesView() []lint.RuleSpec {
	if registry == nil {
		return nil
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.cachedRules == nil {
		cached := make([]lint.RuleSpec, 0, len(registry.byID))
		for _, spec := range registry.byID {
			cached = append(cached, spec)
		}

		sortRuleSpecs(cached)
		registry.cachedRules = cached
	}

	return registry.cachedRules
}

// RulesByModule returns deterministic sorted rules by module namespace.
func (registry *Registry) RulesByModule(module string) []lint.RuleSpec {
	if registry == nil {
		return nil
	}

	moduleKey := strings.TrimSpace(module)
	if moduleKey == "" {
		return nil
	}

	registry.mu.RLock()
	defer registry.mu.RUnlock()

	out := make([]lint.RuleSpec, 0, len(registry.byID))
	for _, spec := range registry.byID {
		if spec.Module != moduleKey {
			continue
		}

		out = append(out, spec)
	}

	sortRuleSpecs(out)
	return out
}

// Snapshot returns deterministic export payload for all registered rules.
func (registry *Registry) Snapshot() lint.RegistrySnapshot {
	if registry == nil {
		return lint.RegistrySnapshot{}
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if !registry.hasCachedSnapshot {
		if registry.cachedModules == nil {
			cachedModules := make([]lint.ModuleSpec, 0, len(registry.byModule))
			for _, spec := range registry.byModule {
				cachedModules = append(cachedModules, spec)
			}

			sortModuleSpecs(cachedModules)
			registry.cachedModules = cachedModules
		}

		if registry.cachedRules == nil {
			cachedRules := make([]lint.RuleSpec, 0, len(registry.byID))
			for _, spec := range registry.byID {
				cachedRules = append(cachedRules, spec)
			}

			sortRuleSpecs(cachedRules)
			registry.cachedRules = cachedRules
		}

		registry.cachedSnapshot = lint.RegistrySnapshot{
			Modules: cloneModuleSpecs(registry.cachedModules),
			Rules:   cloneRuleSpecs(registry.cachedRules),
		}
		registry.hasCachedSnapshot = true
	}

	return lint.RegistrySnapshot{
		Modules: cloneModuleSpecs(registry.cachedSnapshot.Modules),
		Rules:   cloneRuleSpecs(registry.cachedSnapshot.Rules),
	}
}

// invalidateCachesLocked clears sorted list and snapshot caches.
//
// Caller must hold registry.mu write lock.
func (registry *Registry) invalidateCachesLocked() {
	registry.cachedRules = nil
	registry.cachedModules = nil
	registry.cachedSnapshot = lint.RegistrySnapshot{}
	registry.hasCachedSnapshot = false
}

// cloneRuleSpecs returns copy of rule specs slice.
func cloneRuleSpecs(specs []lint.RuleSpec) []lint.RuleSpec {
	if len(specs) == 0 {
		return nil
	}

	out := make([]lint.RuleSpec, len(specs))
	for index := range specs {
		out[index] = cloneRuleSpec(specs[index])
	}

	return out
}

// cloneModuleSpecs returns copy of module specs slice.
func cloneModuleSpecs(specs []lint.ModuleSpec) []lint.ModuleSpec {
	if len(specs) == 0 {
		return nil
	}

	out := make([]lint.ModuleSpec, len(specs))
	copy(out, specs)
	return out
}

// normalizeRuleSpec validates and normalizes one rule metadata descriptor.
func normalizeRuleSpec(spec lint.RuleSpec) (lint.RuleSpec, error) {
	out := spec
	out.ID = trimASCIIOrSpace(out.ID)
	out.Module = trimASCIIOrSpace(out.Module)
	out.Scope = trimASCIIOrSpace(out.Scope)
	out.ScopeDescription = trimASCIIOrSpace(out.ScopeDescription)
	out.Message = trimASCIIOrSpace(out.Message)
	out.Code = trimASCIIOrSpace(out.Code)

	if out.DefaultSeverity == "" {
		out.DefaultSeverity = lint.SeverityWarning
	}

	if out.ID == "" {
		return lint.RuleSpec{}, fmt.Errorf("%w: empty id", ErrInvalidRuleSpec)
	}

	if !isValidRuleID(out.ID) {
		return lint.RuleSpec{}, fmt.Errorf("%w: invalid id %q", ErrInvalidRuleSpec, out.ID)
	}

	if out.Module == "" {
		return lint.RuleSpec{}, fmt.Errorf("%w: empty module for %q", ErrInvalidRuleSpec, out.ID)
	}

	if !isValidModule(out.Module) {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: invalid module %q",
			ErrInvalidRuleSpec,
			out.Module,
		)
	}

	if !strings.HasPrefix(out.ID, out.Module+".") {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: id %q must start with module prefix %q",
			ErrInvalidRuleSpec,
			out.ID,
			out.Module+".",
		)
	}

	if out.Message == "" {
		return lint.RuleSpec{}, fmt.Errorf("%w: empty message for %q", ErrInvalidRuleSpec, out.ID)
	}

	if !isSupportedSeverity(out.DefaultSeverity) {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: invalid severity %q for %q",
			ErrInvalidRuleSpec,
			out.DefaultSeverity,
			out.ID,
		)
	}

	if out.Code != "" && !isValidCode(out.Code) {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: invalid code %q for %q",
			ErrInvalidRuleSpec,
			out.Code,
			out.ID,
		)
	}

	if out.Scope != "" && !isValidScope(out.Scope) {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: invalid scope %q for %q",
			ErrInvalidRuleSpec,
			out.Scope,
			out.ID,
		)
	}

	if out.Scope != "" && out.ScopeDescription == "" {
		return lint.RuleSpec{}, fmt.Errorf(
			"%w: empty scope_description for %q",
			ErrInvalidRuleSpec,
			out.ID,
		)
	}

	out.DefaultEnabled = cloneBoolPtr(out.DefaultEnabled)
	out.DefaultOptions = cloneDynamicValue(out.DefaultOptions)
	out.FileKinds = lint.NormalizeFileKinds(out.FileKinds)
	return out, nil
}

// normalizeModuleSpec validates and normalizes one module metadata descriptor.
func normalizeModuleSpec(spec lint.ModuleSpec) (lint.ModuleSpec, error) {
	out := spec
	out.ID = trimASCIIOrSpace(out.ID)
	out.Name = trimASCIIOrSpace(out.Name)
	out.Description = trimASCIIOrSpace(out.Description)

	if out.ID == "" {
		return lint.ModuleSpec{}, fmt.Errorf("%w: empty id", ErrInvalidModuleSpec)
	}

	if !isValidModule(out.ID) {
		return lint.ModuleSpec{}, fmt.Errorf(
			"%w: invalid id %q",
			ErrInvalidModuleSpec,
			out.ID,
		)
	}

	return out, nil
}

// mergeModuleSpecs merges two metadata descriptors for the same module ID.
func mergeModuleSpecs(
	base lint.ModuleSpec,
	update lint.ModuleSpec,
) (lint.ModuleSpec, error) {
	if trimASCIIOrSpace(base.ID) != trimASCIIOrSpace(update.ID) {
		return lint.ModuleSpec{}, fmt.Errorf(
			"%w: module id mismatch %q vs %q",
			ErrInvalidModuleSpec,
			base.ID,
			update.ID,
		)
	}

	merged := base
	if update.Name != "" {
		if merged.Name != "" && merged.Name != update.Name {
			return lint.ModuleSpec{}, fmt.Errorf(
				"%w: id %q name %q vs %q",
				ErrConflictingModuleSpec,
				base.ID,
				merged.Name,
				update.Name,
			)
		}

		merged.Name = update.Name
	}

	if update.Description != "" {
		if merged.Description != "" && merged.Description != update.Description {
			return lint.ModuleSpec{}, fmt.Errorf(
				"%w: id %q description conflict",
				ErrConflictingModuleSpec,
				base.ID,
			)
		}

		merged.Description = update.Description
	}

	return merged, nil
}

// trimASCIIOrSpace trims string only when boundary spaces are present.
func trimASCIIOrSpace(value string) string {
	if value == "" {
		return ""
	}

	if !isBoundarySpace(value[0]) && !isBoundarySpace(value[len(value)-1]) {
		return value
	}

	return strings.TrimSpace(value)
}

// isBoundarySpace reports whether byte can be treated as ASCII boundary space.
func isBoundarySpace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

// isSupportedSeverity reports whether severity value is supported.
func isSupportedSeverity(severity lint.Severity) bool {
	return lint.IsSupportedSeverity(severity)
}

// sortRuleSpecs applies deterministic rule ordering for exports and listings.
func sortRuleSpecs(specs []lint.RuleSpec) {
	slices.SortStableFunc(specs, func(left lint.RuleSpec, right lint.RuleSpec) int {
		if left.Module != right.Module {
			if left.Module < right.Module {
				return -1
			}

			return 1
		}

		if left.ID != right.ID {
			if left.ID < right.ID {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Message, right.Message)
	})
}

// sortModuleSpecs applies deterministic module ordering for exports/listings.
func sortModuleSpecs(specs []lint.ModuleSpec) {
	slices.SortStableFunc(specs, func(left lint.ModuleSpec, right lint.ModuleSpec) int {
		if left.ID != right.ID {
			if left.ID < right.ID {
				return -1
			}

			return 1
		}

		if left.Name != right.Name {
			if left.Name < right.Name {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Description, right.Description)
	})
}

// isValidRuleID reports whether rule ID has stable dot-separated token form.
func isValidRuleID(value string) bool {
	if value == "" {
		return false
	}

	if value[0] == '.' || value[len(value)-1] == '.' {
		return false
	}

	segmentStart := 0
	segmentCount := 0
	for index := 0; index < len(value); index++ {
		if value[index] != '.' {
			continue
		}

		if !isValidIDToken(value[segmentStart:index]) {
			return false
		}

		segmentCount++
		segmentStart = index + 1
	}

	if !isValidIDToken(value[segmentStart:]) {
		return false
	}

	segmentCount++
	return segmentCount >= 2
}

// isValidModule reports whether module namespace token is valid.
func isValidModule(value string) bool {
	return lint.IsValidModuleToken(value)
}

// isValidCode reports whether short code token is valid.
func isValidCode(value string) bool {
	code, ok := lint.ParsePublicCode(value)
	return ok && code != 0
}

// isValidScope reports whether scope token is valid.
func isValidScope(value string) bool {
	return lint.IsValidScopeToken(value)
}

// isValidIDToken reports whether token matches [A-Za-z][A-Za-z0-9_-]*.
func isValidIDToken(value string) bool {
	return lint.IsValidScopeToken(value)
}

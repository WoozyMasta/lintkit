// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

// Package collector discovers and collects lint rule providers from
// Go packages using go toolchain.
package collector

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// lintkitProviderImportPath stores built-in lintkit provider package path.
	lintkitProviderImportPath = "github.com/woozymasta/lintkit/linting"
)

// CollectSnapshot resolves providers and returns collected registry snapshot.
func CollectSnapshot(options Options) (lint.RegistrySnapshot, error) {
	resolvedOptions, err := normalizeOptions(options)
	if err != nil {
		return lint.RegistrySnapshot{}, err
	}

	packages, err := ResolvePackages(
		resolvedOptions.WorkDir,
		resolvedOptions.Modules,
		resolvedOptions.IncludeLintkitRules,
	)
	if err != nil {
		return lint.RegistrySnapshot{}, fmt.Errorf("%w: %w", ErrProviderDiscovery, err)
	}

	payload, err := RunCollector(
		resolvedOptions.WorkDir,
		resolvedOptions.CollectorTempDir,
		resolvedOptions.KeepCollector,
		packages,
		resolvedOptions.StrictProviders,
		resolvedOptions.Scopes,
		resolvedOptions.Stages,
	)
	if err != nil {
		return lint.RegistrySnapshot{}, fmt.Errorf("%w: %w", ErrProviderCollection, err)
	}

	var snapshot lint.RegistrySnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return lint.RegistrySnapshot{}, fmt.Errorf(
			"decode collected snapshot: %w",
			err,
		)
	}

	return snapshot, nil
}

// ResolvePackages returns unique sorted provider package import paths.
func ResolvePackages(
	workDir string,
	modules []string,
	includeLintkitRules bool,
) ([]string, error) {
	packages := normalizeModuleImportPaths(modules)
	if len(packages) > 0 {
		return packages, nil
	}

	if err := ensureGoModuleInWorkDir(workDir); err != nil {
		return nil, err
	}

	discovered, err := discoverImportPaths(workDir)
	if err != nil {
		return nil, err
	}

	packages = append(
		packages,
		filterDiscoveredProviders(discovered, includeLintkitRules)...,
	)
	packages = normalizeModuleImportPaths(packages)
	if len(packages) == 0 {
		return nil, ErrNoProviderPackages
	}

	return packages, nil
}

// discoverImportPaths resolves provider packages from dependency graph.
func discoverImportPaths(workDir string) ([]string, error) {
	dependencyPackages, err := listDependencyPackages(workDir)
	if err != nil {
		return nil, err
	}

	return detectProviderImportPaths(dependencyPackages), nil
}

// filterDiscoveredProviders applies discovery-time provider exclusions.
func filterDiscoveredProviders(
	packages []string,
	includeLintkitRules bool,
) []string {
	if includeLintkitRules || len(packages) == 0 {
		return packages
	}

	filtered := make([]string, 0, len(packages))
	for index := range packages {
		if packages[index] == lintkitProviderImportPath {
			continue
		}

		filtered = append(filtered, packages[index])
	}

	slices.Sort(filtered)
	return filtered
}

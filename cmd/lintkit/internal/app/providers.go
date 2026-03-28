// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"errors"
	"fmt"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/collector"
	"github.com/woozymasta/lintkit/lint"
)

// collectRegistrySnapshot resolves provider packages and collects rule snapshot.
func collectRegistrySnapshot(options ProviderCollectOptions) (lint.RegistrySnapshot, error) {
	snapshot, err := collector.CollectSnapshot(collector.Options{
		WorkDir:         options.WorkDir,
		Modules:         options.Modules,
		StrictProviders: !options.SoftProviders,
		Scopes:          options.Scopes,
		Stages:          options.Stages,
	})
	if err == nil {
		return snapshot, nil
	}

	if errors.Is(err, collector.ErrNoProviderPackages) {
		return lint.RegistrySnapshot{}, ErrNoProviderPackages
	}

	if errors.Is(err, collector.ErrProviderDiscovery) {
		return lint.RegistrySnapshot{}, fmt.Errorf("%w: %w", ErrProviderDiscoveryFailed, err)
	}

	if errors.Is(err, collector.ErrProviderCollection) {
		return lint.RegistrySnapshot{}, fmt.Errorf("%w: %w", ErrProviderCollectionFailed, err)
	}

	return lint.RegistrySnapshot{}, err
}

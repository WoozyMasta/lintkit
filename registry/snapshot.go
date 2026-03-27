// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package registry

import (
	"fmt"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

// SnapshotFromProviders builds deterministic snapshot from provider list.
func SnapshotFromProviders(providers ...lint.RuleProvider) (lint.RegistrySnapshot, error) {
	engine := linting.NewEngine()
	if err := RegisterProviders(engine, providers...); err != nil {
		return lint.RegistrySnapshot{}, err
	}

	return engine.Snapshot(), nil
}

// RegisterProviders registers providers into existing engine.
func RegisterProviders(engine *linting.Engine, providers ...lint.RuleProvider) error {
	if engine == nil {
		return ErrNilEngine
	}

	for index := range providers {
		if providers[index] == nil {
			return fmt.Errorf("%w: provider[%d]", ErrNilProvider, index)
		}

		if err := providers[index].RegisterRules(engine); err != nil {
			return fmt.Errorf("register provider[%d]: %w", index, err)
		}
	}

	return nil
}

// SnapshotFromEngine returns engine registry snapshot.
func SnapshotFromEngine(engine *linting.Engine) (lint.RegistrySnapshot, error) {
	if engine == nil {
		return lint.RegistrySnapshot{}, ErrNilEngine
	}

	return engine.Snapshot(), nil
}

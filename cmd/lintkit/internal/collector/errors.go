// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import "errors"

var (
	// ErrNoProviderPackages indicates empty provider package resolution.
	ErrNoProviderPackages = errors.New("no provider packages resolved")

	// ErrNoGoModuleInWorkDir indicates missing go.mod in selected workdir.
	ErrNoGoModuleInWorkDir = errors.New("workdir has no go.mod")

	// ErrInvalidWorkDir indicates invalid provider collection working directory.
	ErrInvalidWorkDir = errors.New("invalid provider collection workdir")

	// ErrInvalidCollectorTempDir indicates invalid collector temporary directory.
	ErrInvalidCollectorTempDir = errors.New("invalid collector temporary directory")

	// ErrProviderDiscovery indicates provider discovery failure.
	ErrProviderDiscovery = errors.New("provider discovery failed")

	// ErrProviderCollection indicates provider runtime collection failure.
	ErrProviderCollection = errors.New("provider collection failed")

	// ErrConflictingRuleFilters indicates mixed stage and scope filter input.
	ErrConflictingRuleFilters = errors.New("conflicting provider rule filters")
)

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

// Package app implements lintkit CLI command execution logic.
package app

import "errors"

var (
	// ErrNoProviderPackages indicates that provider collection resolved no modules.
	ErrNoProviderPackages = errors.New("no provider packages resolved; use --module and/or set --workdir to project root")

	// ErrProviderDiscoveryFailed indicates provider discovery failure.
	ErrProviderDiscoveryFailed = errors.New("provider discovery failed")

	// ErrProviderCollectionFailed indicates provider collection runtime failure.
	ErrProviderCollectionFailed = errors.New("provider collection failed")

	// ErrCheckRequiresOutputPath indicates --check mode without file output path.
	ErrCheckRequiresOutputPath = errors.New("--check requires output file path")

	// ErrEmptyStdinSnapshot indicates empty snapshot payload from stdin.
	ErrEmptyStdinSnapshot = errors.New("read stdin: empty input")

	// ErrInvalidPolicySchemaProperties indicates missing object properties node.
	ErrInvalidPolicySchemaProperties = errors.New("invalid policy schema: missing object properties")

	// ErrInvalidPolicySchemaProperty indicates missing required property node.
	ErrInvalidPolicySchemaProperty = errors.New("invalid policy schema: missing object property")

	// ErrInvalidPolicySchemaRulesItems indicates invalid rules items node.
	ErrInvalidPolicySchemaRulesItems = errors.New("invalid policy schema: rules.items is not an object")

	// ErrUnsupportedSchemaRef indicates unsupported schema ref style.
	ErrUnsupportedSchemaRef = errors.New("unsupported schema ref")

	// ErrInvalidSchemaRefPath indicates unresolved local schema ref.
	ErrInvalidSchemaRefPath = errors.New("invalid schema ref path")
)

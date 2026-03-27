// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "errors"

var (
	// ErrInvalidCatalogProvider indicates invalid catalog provider config.
	ErrInvalidCatalogProvider = errors.New("invalid catalog provider")

	// ErrInvalidCodePrefix indicates unsupported exported code prefix token.
	ErrInvalidCodePrefix = errors.New("invalid code prefix")

	// ErrInvalidCodeCatalogConfig indicates unsupported code catalog config.
	ErrInvalidCodeCatalogConfig = errors.New("invalid code catalog config")

	// ErrUnknownCodeCatalogCode indicates unknown code token in catalog lookup.
	ErrUnknownCodeCatalogCode = errors.New("unknown code catalog code")

	// ErrNilRuleRegistrar indicates nil rule registrar in provider registration.
	ErrNilRuleRegistrar = errors.New("rule registrar is nil")

	// ErrNilRuleProvider indicates nil rule provider argument.
	ErrNilRuleProvider = errors.New("rule provider is nil")

	// ErrUnknownOverrideRuleID indicates unknown rule id in override map.
	ErrUnknownOverrideRuleID = errors.New("unknown override rule id")
)

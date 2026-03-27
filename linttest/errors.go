// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linttest

import "errors"

var (
	// ErrEmptyCatalogModule indicates empty lint module token in contract checks.
	ErrEmptyCatalogModule = errors.New("catalog contract: empty module")

	// ErrNilRuleIDMapper indicates nil rule id mapper callback.
	ErrNilRuleIDMapper = errors.New("catalog contract: nil rule id mapper")
)

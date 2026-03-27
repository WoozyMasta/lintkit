// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package registry

import "errors"

var (
	// ErrNilProvider indicates nil rule provider in provider list.
	ErrNilProvider = errors.New("rule provider is nil")

	// ErrNilEngine indicates nil linting engine input.
	ErrNilEngine = errors.New("linting engine is nil")
)

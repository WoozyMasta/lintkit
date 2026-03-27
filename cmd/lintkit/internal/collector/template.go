// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import _ "embed"

// collectorProgramTemplateText stores embedded go-run helper template.
//
//go:embed templates/main.go.gotmpl
var collectorProgramTemplateText string

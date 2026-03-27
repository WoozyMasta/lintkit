// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import _ "embed"

//go:embed templates/list.gotmpl
var markdownListTemplate string

//go:embed templates/table.gotmpl
var markdownTableTemplate string

//go:embed templates/html.gotmpl
var markdownHTMLTemplate string

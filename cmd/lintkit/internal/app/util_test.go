// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import "testing"

func TestResolveOutputFormat(t *testing.T) {
	t.Parallel()

	if got := ResolveOutputFormat("yaml", "rules.snapshot.json"); got != "yaml" {
		t.Fatalf("ResolveOutputFormat(explicit)=%q, want yaml", got)
	}

	if got := ResolveOutputFormat("", "rules.snapshot.yaml"); got != "yaml" {
		t.Fatalf("ResolveOutputFormat(yaml ext)=%q, want yaml", got)
	}

	if got := ResolveOutputFormat("", "rules.snapshot"); got != "json" {
		t.Fatalf("ResolveOutputFormat(default)=%q, want json", got)
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "testing"

func TestNormalizeFileKind(t *testing.T) {
	if got := NormalizeFileKind("  Source.CFG  "); got != "source.cfg" {
		t.Fatalf("NormalizeFileKind()=%q, want source.cfg", got)
	}

	if got := NormalizeFileKind("  "); got != "" {
		t.Fatalf("NormalizeFileKind(blank)=%q, want empty", got)
	}
}

func TestNormalizeFileKinds(t *testing.T) {
	got := NormalizeFileKinds([]FileKind{
		" source.cfg ",
		"module_beta",
		"MODULE_BETA",
		"",
	})

	if len(got) != 2 {
		t.Fatalf("NormalizeFileKinds() len=%d, want 2", len(got))
	}

	if got[0] != "module_beta" || got[1] != "source.cfg" {
		t.Fatalf("NormalizeFileKinds()=%v, want [module_beta source.cfg]", got)
	}
}

func TestSupportsFileKind(t *testing.T) {
	if !SupportsFileKind(nil, "source.cfg") {
		t.Fatal("SupportsFileKind(nil, source.cfg)=false, want true")
	}

	if !SupportsFileKind([]FileKind{"source.cfg"}, "source.cfg") {
		t.Fatal("SupportsFileKind(source.cfg, source.cfg)=false, want true")
	}

	if SupportsFileKind([]FileKind{"module_beta"}, "source.cfg") {
		t.Fatal("SupportsFileKind(module_beta, source.cfg)=true, want false")
	}

	if !SupportsFileKind([]FileKind{FileKindAny}, "source.cfg") {
		t.Fatal("SupportsFileKind(*, source.cfg)=false, want true")
	}
}

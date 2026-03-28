// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"errors"
	"testing"
)

func TestCodeCatalogHandleNil(t *testing.T) {
	t.Parallel()

	var handle *CodeCatalogHandle

	_, err := handle.Catalog()
	if !errors.Is(err, ErrNilCodeCatalogHandle) {
		t.Fatalf("Catalog() error=%v, want ErrNilCodeCatalogHandle", err)
	}

	if got := handle.RuleIDOrUnknown(1001); got != "unknown.unknown" {
		t.Fatalf("RuleIDOrUnknown()=%q, want unknown.unknown", got)
	}
}

func TestCodeCatalogHandleHappyPath(t *testing.T) {
	t.Parallel()

	handle := NewCodeCatalogHandle(
		CodeCatalogConfig{
			Module:     "alpha",
			CodePrefix: "ALPHA",
			ScopeDescriptions: map[Stage]string{
				"parse": "Parser diagnostics.",
			},
		},
		[]CodeSpec{
			ErrorCodeSpec(2001, "parse", "unexpected token"),
		},
	)

	catalog, err := handle.Catalog()
	if err != nil {
		t.Fatalf("Catalog() error: %v", err)
	}

	if _, ok := catalog.ByCode(2001); !ok {
		t.Fatal("Catalog().ByCode(2001) not found")
	}

	if got := handle.RuleIDOrUnknown(2001); got != "alpha.parse.unexpected-token" {
		t.Fatalf(
			"RuleIDOrUnknown(2001)=%q, want alpha.parse.unexpected-token",
			got,
		)
	}

	if got := handle.RuleIDOrUnknown(9999); got != "alpha.unknown" {
		t.Fatalf("RuleIDOrUnknown(9999)=%q, want alpha.unknown", got)
	}

	if module := handle.ModuleSpec(); module.ID != "alpha" {
		t.Fatalf("ModuleSpec().ID=%q, want alpha", module.ID)
	}

	if len(handle.CodeSpecs()) != 1 {
		t.Fatalf("len(CodeSpecs())=%d, want 1", len(handle.CodeSpecs()))
	}

	if len(handle.RuleSpecs()) != 1 {
		t.Fatalf("len(RuleSpecs())=%d, want 1", len(handle.RuleSpecs()))
	}

	spec, err := handle.RuleSpec(ErrorCodeSpec(2001, "parse", "unexpected token"))
	if err != nil {
		t.Fatalf("RuleSpec() error: %v", err)
	}

	if spec.ID != "alpha.parse.unexpected-token" {
		t.Fatalf("RuleSpec().ID=%q, want alpha.parse.unexpected-token", spec.ID)
	}
}

func TestCodeCatalogHandleInvalidConfig(t *testing.T) {
	t.Parallel()

	handle := NewCodeCatalogHandle(
		CodeCatalogConfig{
			Module:     "alpha",
			CodePrefix: "3D",
		},
		nil,
	)

	_, err := handle.Catalog()
	if !errors.Is(err, ErrInvalidCodeCatalogConfig) {
		t.Fatalf("Catalog() error=%v, want ErrInvalidCodeCatalogConfig", err)
	}

	if got := handle.RuleIDOrUnknown(2001); got != "alpha.unknown" {
		t.Fatalf("RuleIDOrUnknown()=%q, want alpha.unknown", got)
	}

	_, err = handle.RuleSpec(ErrorCodeSpec(2001, "parse", "unexpected token"))
	if !errors.Is(err, ErrInvalidCodeCatalogConfig) {
		t.Fatalf("RuleSpec() error=%v, want ErrInvalidCodeCatalogConfig", err)
	}
}

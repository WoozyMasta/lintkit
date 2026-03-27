// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package registryio

import (
	"errors"
	"strings"
	"testing"
)

func TestReadSnapshotYAMLFromStdin(t *testing.T) {
	t.Parallel()

	snapshot, err := ReadSnapshot(
		"-",
		strings.NewReader("rules:\n  - id: module_alpha.rule\n    module: module_alpha\n    title: Rule\n    message: Rule\n"),
		errors.New("empty stdin"),
	)
	if err != nil {
		t.Fatalf("ReadSnapshot(yaml stdin) error: %v", err)
	}

	if len(snapshot.Rules) != 1 {
		t.Fatalf("len(snapshot.Rules)=%d, want 1", len(snapshot.Rules))
	}
}

func TestReadSnapshotEmptyStdinError(t *testing.T) {
	t.Parallel()

	emptyErr := errors.New("empty stdin")
	_, err := ReadSnapshot("-", strings.NewReader("   "), emptyErr)
	if !errors.Is(err, emptyErr) {
		t.Fatalf("ReadSnapshot(empty stdin) error=%v, want emptyErr", err)
	}
}

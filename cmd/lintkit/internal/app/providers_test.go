// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCollectRegistrySnapshotNoProviders(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	goModText := "module example.com/test\n\ngo 1.26\n"
	if err := os.WriteFile(goModPath, []byte(goModText), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error: %v", err)
	}

	_, err := collectRegistrySnapshot(ProviderCollectOptions{
		WorkDir: workDir,
	})
	if !errors.Is(err, ErrNoProviderPackages) {
		t.Fatalf("collectRegistrySnapshot() error=%v, want ErrNoProviderPackages", err)
	}
}

func TestCollectRegistrySnapshotProviderCollectionFailure(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	goModText := "module example.com/test\n\ngo 1.26\n"
	if err := os.WriteFile(goModPath, []byte(goModText), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error: %v", err)
	}

	_, err := collectRegistrySnapshot(ProviderCollectOptions{
		WorkDir: workDir,
		Modules: []string{"not a valid import"},
	})
	if !errors.Is(err, ErrProviderCollectionFailed) {
		t.Fatalf(
			"collectRegistrySnapshot(invalid module) error=%v, want ErrProviderCollectionFailed",
			err,
		)
	}
}

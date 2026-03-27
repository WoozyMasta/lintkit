// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryExportGoldenFiles(t *testing.T) {
	t.Parallel()

	registry := buildTestRegistry(t)

	tests := []struct {
		name     string
		fileName string
		export   func() ([]byte, error)
	}{
		{
			name:     "json",
			fileName: "export_registry.golden.json",
			export: func() ([]byte, error) {
				return registry.ExportJSON(true)
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			expectedPath := filepath.Join("..", "testdata", test.fileName)
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("ReadFile(%s) error: %v", expectedPath, err)
			}

			got, err := test.export()
			if err != nil {
				t.Fatalf("export() error: %v", err)
			}

			gotText := normalizeLineEndings(string(got))
			expectedText := normalizeLineEndings(string(expected))
			if gotText != expectedText {
				t.Fatalf(
					"export mismatch for %s\n--- got ---\n%s\n--- expected ---\n%s",
					test.fileName,
					gotText,
					expectedText,
				)
			}
		})
	}
}

// normalizeLineEndings normalizes text to LF for cross-platform golden checks.
func normalizeLineEndings(value string) string {
	normalized := strings.ReplaceAll(value, "\r\n", "\n")
	normalized = strings.TrimPrefix(normalized, "\uFEFF")
	normalized = strings.TrimSuffix(normalized, "\n")

	return normalized
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveOutputFormat resolves output format from flag or file extension.
func ResolveOutputFormat(raw string, outputPath string) string {
	format := strings.ToLower(strings.TrimSpace(raw))
	if format == "json" || format == "yaml" {
		return format
	}

	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(outputPath)))
	switch extension {
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "json"
	}
}

// checkOutput compares rendered content with target output file contents.
func checkOutput(rendered []byte, outputPath string) error {
	current, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("read check target %q: %w", outputPath, err)
	}

	if bytes.Equal(current, rendered) {
		return nil
	}

	return fmt.Errorf("output differs: %s", outputPath)
}

// writeOutput writes rendered output to selected file path.
func writeOutput(path string, data []byte) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return fmt.Errorf("create output dir %q: %w", dirPath, err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	return nil
}

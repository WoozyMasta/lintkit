// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// RunCollector executes generated helper and returns snapshot payload bytes.
func RunCollector(
	workDir string,
	collectorTempDir string,
	keepCollector bool,
	packages []string,
	strictProviders bool,
	scopes []string,
	stages []string,
) ([]byte, error) {
	source, err := BuildCollectorProgram(
		packages,
		strictProviders,
		scopes,
		stages,
	)
	if err != nil {
		return nil, err
	}

	tempFile, err := os.CreateTemp(collectorTempDir, "lintkit-collect-*.go")
	if err != nil {
		return nil, fmt.Errorf("%w: create temporary collector: %w", ErrProviderCollection, err)
	}

	tempPath := tempFile.Name()
	if !keepCollector {
		defer func() {
			_ = os.Remove(tempPath)
		}()
	}

	if _, err := tempFile.Write(source); err != nil {
		_ = tempFile.Close()
		return nil, fmt.Errorf("%w: write temporary collector: %w", ErrProviderCollection, err)
	}

	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("%w: close temporary collector: %w", ErrProviderCollection, err)
	}

	payload, err := runGoProgram(workDir, tempPath)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// normalizeOptions validates and normalizes provider collection options.
func normalizeOptions(options Options) (Options, error) {
	options.WorkDir = strings.TrimSpace(options.WorkDir)
	if options.WorkDir == "" {
		options.WorkDir = "."
	}

	absWorkDir, err := filepath.Abs(options.WorkDir)
	if err != nil {
		return Options{}, fmt.Errorf("%w: resolve %q: %w", ErrInvalidWorkDir, options.WorkDir, err)
	}

	options.WorkDir = absWorkDir
	options.CollectorTempDir = strings.TrimSpace(options.CollectorTempDir)
	if options.CollectorTempDir != "" {
		resolvedCollectorTempDir := options.CollectorTempDir
		if !filepath.IsAbs(resolvedCollectorTempDir) {
			resolvedCollectorTempDir = filepath.Join(
				absWorkDir,
				resolvedCollectorTempDir,
			)
		}

		if mkdirErr := os.MkdirAll(resolvedCollectorTempDir, 0o750); mkdirErr != nil {
			return Options{}, fmt.Errorf(
				"%w: prepare %q: %w",
				ErrInvalidCollectorTempDir,
				options.CollectorTempDir,
				mkdirErr,
			)
		}

		collectorInfo, statErr := os.Stat(resolvedCollectorTempDir)
		if statErr != nil {
			return Options{}, fmt.Errorf(
				"%w: resolve %q: %w",
				ErrInvalidCollectorTempDir,
				options.CollectorTempDir,
				statErr,
			)
		}

		if !collectorInfo.IsDir() {
			return Options{}, fmt.Errorf(
				"%w: %q is not a directory",
				ErrInvalidCollectorTempDir,
				options.CollectorTempDir,
			)
		}

		options.CollectorTempDir = resolvedCollectorTempDir
	}

	options.Scopes = normalizeFilterTokens(options.Scopes)
	options.Stages = normalizeFilterTokens(options.Stages)
	if len(options.Scopes) > 0 && len(options.Stages) > 0 {
		return Options{}, ErrConflictingRuleFilters
	}

	return options, nil
}

// runGoProgram executes one generated go source and returns stdout payload.
func runGoProgram(workDir string, sourcePath string) ([]byte, error) {
	command := exec.Command("go", "run", sourcePath)
	command.Dir = workDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}

		return nil, fmt.Errorf("%w: %s", ErrProviderCollection, detail)
	}

	return stdout.Bytes(), nil
}

// normalizeFilterTokens trims, deduplicates and sorts filter tokens.
func normalizeFilterTokens(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for index := range values {
		item := strings.TrimSpace(values[index])
		if item == "" {
			continue
		}

		if _, exists := seen[item]; exists {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	slices.Sort(out)
	return out
}

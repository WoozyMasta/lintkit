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
	"strings"
)

// RunCollector executes generated helper and returns snapshot payload bytes.
func RunCollector(
	workDir string,
	packages []string,
	strictProviders bool,
) ([]byte, error) {
	source, err := BuildCollectorProgram(packages, strictProviders)
	if err != nil {
		return nil, err
	}

	tempFile, err := os.CreateTemp(workDir, "lintkit-collect-*.go")
	if err != nil {
		return nil, fmt.Errorf("%w: create temporary collector: %w", ErrProviderCollection, err)
	}

	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

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

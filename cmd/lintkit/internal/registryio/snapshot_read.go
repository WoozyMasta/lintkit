// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

// Package registryio provides snapshot IO helpers for lintkit CLI internals.
package registryio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/woozymasta/lintkit/lint"
)

// ReadSnapshot reads one snapshot from file path or stdin marker.
func ReadSnapshot(
	path string,
	stdin io.Reader,
	emptyStdinErr error,
) (lint.RegistrySnapshot, error) {
	payload, err := readSnapshotPayload(path, stdin, emptyStdinErr)
	if err != nil {
		return lint.RegistrySnapshot{}, err
	}

	return decodeSnapshot(payload)
}

// readSnapshotPayload reads raw snapshot bytes from file path or stdin.
func readSnapshotPayload(
	path string,
	stdin io.Reader,
	emptyStdinErr error,
) ([]byte, error) {
	path = strings.TrimSpace(path)

	if path == "" || path == "-" {
		payload, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}

		if len(bytes.TrimSpace(payload)) == 0 {
			return nil, emptyStdinErr
		}

		return payload, nil
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	return payload, nil
}

// decodeSnapshot decodes json/yaml snapshot payload to registry snapshot.
func decodeSnapshot(payload []byte) (lint.RegistrySnapshot, error) {
	var snapshot lint.RegistrySnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		if yamlErr := yaml.Unmarshal(payload, &snapshot); yamlErr != nil {
			return lint.RegistrySnapshot{}, fmt.Errorf(
				"decode snapshot json/yaml: json=%v yaml=%v",
				err,
				yamlErr,
			)
		}
	}

	return snapshot, nil
}

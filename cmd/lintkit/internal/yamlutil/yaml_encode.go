// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

// Package yamlutil provides YAML encoding helpers for lintkit CLI.
package yamlutil

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Marshal renders value as YAML with stable 2-space indentation.
func Marshal(value any) ([]byte, error) {
	var out bytes.Buffer

	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("encode yaml: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("close yaml encoder: %w", err)
	}

	return out.Bytes(), nil
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/registryio"
	"github.com/woozymasta/lintkit/cmd/lintkit/internal/yamlutil"
	"github.com/woozymasta/lintkit/lint"
)

// policySchemaBaseJSON stores generated baseline schema for RunPolicyConfig.
//
// It is generated via workspace-level `schemadoc build` pipeline.
//
//go:embed schema.json
var policySchemaBaseJSON []byte

// RunSchema renders RunPolicyConfig schema from one registry snapshot input.
func (runner *Runner) RunSchema(
	inputPath string,
	outputPath string,
	flags SchemaCommandOptions,
) error {
	snapshot, err := registryio.ReadSnapshot(
		inputPath,
		runner.Stdin,
		ErrEmptyStdinSnapshot,
	)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	rendered, err := renderPolicySchema(
		snapshot,
		ResolveOutputFormat(flags.Format, outputPath),
		flags.Selector,
	)
	if err != nil {
		return err
	}

	outputPath = strings.TrimSpace(outputPath)
	if flags.Check {
		if outputPath == "" || outputPath == "-" {
			return ErrCheckRequiresOutputPath
		}

		return checkOutput(rendered, outputPath)
	}

	if outputPath == "" || outputPath == "-" {
		if _, err := runner.Stdout.Write(rendered); err != nil {
			return fmt.Errorf("write rendered output to stdout: %w", err)
		}

		return nil
	}

	if err := writeOutput(outputPath, rendered); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(runner.Stderr, "written %s (%d bytes)\n", outputPath, len(rendered))
	return nil
}

// renderPolicySchema builds policy schema and serializes it in selected format.
func renderPolicySchema(
	snapshot lint.RegistrySnapshot,
	format string,
	selectors []string,
) ([]byte, error) {
	root, err := buildPolicySchemaRoot(snapshot, selectors)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(root, "", "  ")
	case "yaml":
		data, err := yamlutil.Marshal(root)
		if err != nil {
			return nil, fmt.Errorf("marshal schema yaml: %w", err)
		}

		return data, nil
	default:
		return nil, fmt.Errorf("unsupported schema format %q", format)
	}
}

// buildPolicySchemaRoot loads embedded base schema and injects selector enum.
func buildPolicySchemaRoot(
	snapshot lint.RegistrySnapshot,
	selectors []string,
) (map[string]any, error) {
	payload := append([]byte(nil), policySchemaBaseJSON...)
	root := make(map[string]any)
	if err := json.Unmarshal(payload, &root); err != nil {
		return nil, fmt.Errorf("decode embedded schema map: %w", err)
	}

	selectorEnum, enabled := buildSelectorEnum(snapshot, selectors)
	if !enabled {
		if err := clearSchemaSelectorEnum(root); err != nil {
			return nil, err
		}

		return root, nil
	}

	if err := setSchemaSelectorEnum(root, selectorEnum); err != nil {
		return nil, err
	}

	return root, nil
}

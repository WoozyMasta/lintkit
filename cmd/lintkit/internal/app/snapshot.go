// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/linting"
)

// RunSnapshot collects provider rules and renders registry snapshot output.
func (runner *Runner) RunSnapshot(
	outputPath string,
	providerFlags ProviderCollectOptions,
	flags SnapshotCommandOptions,
) error {
	snapshot, err := collectRegistrySnapshot(providerFlags)
	if err != nil {
		return err
	}

	if serviceDiagnosticsEnabled(flags) {
		snapshot = linting.AppendServiceRules(snapshot)
	}

	outputPath = strings.TrimSpace(outputPath)
	format := ResolveOutputFormat(flags.Format, outputPath)
	rendered, err := renderSnapshot(snapshot, renderOptions{
		Format: format,
		Pretty: true,
	})
	if err != nil {
		return fmt.Errorf("render collected snapshot (%s): %w", format, err)
	}

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

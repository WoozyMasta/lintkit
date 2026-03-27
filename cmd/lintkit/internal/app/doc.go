// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"fmt"
	"strings"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/registryio"
)

// RunDoc renders markdown documentation from one registry snapshot input.
func (runner *Runner) RunDoc(
	inputPath string,
	outputPath string,
	flags DocCommandOptions,
) error {
	snapshot, err := registryio.ReadSnapshot(
		inputPath,
		runner.Stdin,
		ErrEmptyStdinSnapshot,
	)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	rendered, err := renderSnapshot(snapshot, renderOptions{
		Format:              "markdown",
		TemplateName:        strings.TrimSpace(flags.Markdown.TemplateName),
		TemplatePath:        strings.TrimSpace(flags.Markdown.TemplatePath),
		DocumentTitle:       strings.TrimSpace(flags.Markdown.DocumentTitle),
		DocumentDescription: strings.TrimSpace(flags.Markdown.DocumentDescription),
		ExampleFormat:       strings.ToLower(strings.TrimSpace(flags.Markdown.ExampleFormat)),
		TOCMode:             strings.ToLower(strings.TrimSpace(flags.Markdown.TOCMode)),
		WrapWidth:           flags.Markdown.WrapWidth,
		FooterToolName:      strings.TrimSpace(flags.Markdown.FooterToolName),
		FooterToolURL:       strings.TrimSpace(flags.Markdown.FooterToolURL),
		FooterVersion:       strings.TrimSpace(flags.Markdown.FooterVersion),
		FooterCommit:        strings.TrimSpace(flags.Markdown.FooterCommit),
	})
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

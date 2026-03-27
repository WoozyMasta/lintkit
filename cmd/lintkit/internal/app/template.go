// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/render"
)

// RunTemplate writes selected built-in markdown template to file or stdout.
func (runner *Runner) RunTemplate(name string, outputPath string) error {
	text, err := render.BuiltinTemplate(name)
	if err != nil {
		return err
	}

	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" || outputPath == "-" {
		if _, err := io.WriteString(runner.Stdout, text); err != nil {
			return fmt.Errorf("write template to stdout: %w", err)
		}

		return nil
	}

	return writeOutput(outputPath, []byte(text))
}

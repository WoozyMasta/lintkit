// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/lintkit/cmd/lintkit/internal/app"
)

// cliRunner executes CLI operations with custom IO streams.
type cliRunner struct {
	app         *app.Runner
	stdout      io.Writer
	stderr      io.Writer
	programName string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes CLI logic and returns process exit code.
func run(args []string, stdout io.Writer, stderr io.Writer) int {
	return runWithIO(args, os.Stdin, stdout, stderr)
}

// runWithIO executes CLI logic with custom stdin, for tests.
func runWithIO(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	programName := filepath.Base(strings.TrimSpace(os.Args[0]))
	if programName == "" {
		programName = "lintkit"
	}

	runner := &cliRunner{
		app: &app.Runner{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		stdout:      stdout,
		stderr:      stderr,
		programName: programName,
	}

	return runner.run(args)
}

// run parses CLI args and maps errors to process exit codes.
func (runner *cliRunner) run(args []string) int {
	err := parseCLIArgs(args, runner)
	if err == nil {
		return 0
	}

	var flagsErr *flags.Error
	if errors.As(err, &flagsErr) {
		if flagsErr.Type == flags.ErrHelp {
			writeCLIError(runner.stdout, err)
			return 0
		}

		writeCLIError(runner.stderr, err)
		return 2
	}

	writeCLIError(runner.stderr, err)
	return 1
}

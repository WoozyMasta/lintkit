// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"fmt"
	"io"

	"github.com/jessevdk/go-flags"
)

// parseCLIArgs parses CLI args and triggers selected subcommand execution.
func parseCLIArgs(args []string, runner *cliRunner) error {
	options := &cliOptions{}
	options.Snapshot.runner = runner
	options.Doc.runner = runner
	options.Schema.runner = runner
	options.Template.runner = runner

	parser := flags.NewParser(options, flags.HelpFlag)
	parser.Name = runner.programName
	applyCommandLongDescriptions(parser, runner.programName)

	_, err := parser.ParseArgs(args)
	if err != nil {
		return err
	}

	return nil
}

// applyCommandLongDescriptions configures detailed command help text.
func applyCommandLongDescriptions(parser *flags.Parser, programName string) {
	descriptions := commandLongDescriptions(programName)
	for commandName, description := range descriptions {
		command := parser.Find(commandName)
		if command == nil {
			continue
		}

		command.LongDescription = description
	}
}

// writeCLIError writes plain CLI error text to selected stream.
func writeCLIError(output io.Writer, err error) {
	if err == nil {
		return
	}

	_, _ = fmt.Fprintln(output, err.Error())
}

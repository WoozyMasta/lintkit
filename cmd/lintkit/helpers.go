// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/woozymasta/flags"
)

// parseCLIArgs parses CLI args and triggers selected subcommand execution.
func parseCLIArgs(args []string, runner *cliRunner) error {
	options := &cliOptions{}
	options.Snapshot.runner = runner
	options.Doc.runner = runner
	options.Schema.runner = runner
	options.Template.runner = runner

	parser := flags.NewParser(
		options,
		flags.HelpFlag|
			flags.VersionFlag|
			flags.HelpCommand|
			flags.VersionCommand|
			flags.CompletionCommand|
			flags.DocsCommand|
			flags.KeepDescriptionWhitespace|
			flags.PrintHelpOnInputErrors|
			flags.ShowRepeatableInHelp|
			flags.DetectShellFlagStyle|
			flags.DetectShellEnvStyle,
	)
	parser.Name = runner.programName
	fields := flags.VersionFieldsCore
	if BuildTime.IsZero() {
		fields &^= flags.VersionFieldBuilt
	}

	parser.SetVersionFields(fields)
	parser.SetVersionInfo(flags.VersionInfo{
		File:         os.Args[0],
		Version:      Version,
		Revision:     Commit,
		RevisionTime: BuildTime,
		URL:          URL,
	})
	if err := parser.EnsureBuiltinCommands(); err != nil {
		return err
	}

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

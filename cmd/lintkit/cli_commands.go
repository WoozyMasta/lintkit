// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"github.com/woozymasta/lintkit/cmd/lintkit/internal/app"
)

// cliOptions describes lintkit CLI flags and subcommands.
type cliOptions struct {
	Version  versionCommand  `command:"version" description:"Print version information"`
	Doc      docCommand      `command:"doc" description:"Render documentation from registry snapshot"`
	Schema   schemaCommand   `command:"schema" description:"Render JSON Schema from registry snapshot"`
	Template templateCommand `command:"template" description:"Print built-in documentation template"`
	Snapshot snapshotCommand `command:"snapshot" description:"Collect rules from modules or dependency graph and render registry snapshot"`
}

// markdownRenderFlags groups markdown template and layout flags.
type markdownRenderFlags struct {
	TemplateName  string `short:"t" long:"template" description:"Built-in documentation template style" choice:"list" choice:"table" choice:"html" default:"list"`
	TemplatePath  string `short:"p" long:"template-file" description:"Path to external template file (.gotmpl), overrides built-in template"`
	Title         string `short:"T" long:"title" description:"Custom document title in rendered output"`
	Description   string `short:"D" long:"description" description:"Custom document description in rendered output"`
	ExampleFormat string `short:"e" long:"example-format" description:"Append policy example block to markdown output" choice:"json" choice:"yaml"`
	TOCMode       string `short:"o" long:"toc" description:"TOC mode for markdown output" choice:"auto" choice:"always" choice:"off" default:"auto"`
	WrapWidth     int    `short:"w" long:"wrap" description:"Wrap width for plain text fields in markdown output" default:"80"`
}

// checkFlags groups diff-only output validation flag.
type checkFlags struct {
	Check bool `short:"c" long:"check" description:"Check rendered output against output file and exit non-zero on diff"`
}

// docFlags groups markdown doc render options.
type docFlags struct {
	Markdown markdownRenderFlags `group:"Markdown Render"`
	Check    checkFlags          `group:"Output Check"`
}

// snapshotFlags groups registry snapshot render and provider collect options.
type snapshotFlags struct {
	Format        string     `short:"f" long:"format" description:"Snapshot output format (inferred from snapshot extension when omitted)" choice:"json" choice:"yaml"`
	WorkDir       string     `short:"r" long:"workdir" description:"Working directory for provider collection commands" default:"."`
	Modules       []string   `short:"m" long:"module" description:"Go package import path with LintRulesProvider (repeatable)"`
	SoftProviders bool       `short:"s" long:"soft-providers" description:"Allow duplicate provider rule conflicts and keep first registered rule"`
	Check         checkFlags `group:"Output Check"`
}

// schemaFlags groups policy schema rendering options.
type schemaFlags struct {
	Format   string     `short:"f" long:"format" description:"Schema output format (inferred from output extension when omitted)" choice:"json" choice:"yaml"`
	Selector []string   `short:"s" long:"selector" description:"Selector enum kinds (repeatable)" choice:"all" choice:"none" choice:"module" choice:"id" choice:"code"`
	Check    checkFlags `group:"Output Check"`
}

// snapshotCommand collects providers and renders registry snapshot.
type snapshotCommand struct {
	runner *cliRunner

	Args struct {
		Output string `positional-arg-name:"output" description:"Output snapshot path (optional; stdout when omitted)"`
	} `positional-args:"yes"`
	SnapshotFlags snapshotFlags `group:"Snapshot"`
}

// Execute runs snapshot subcommand.
func (command *snapshotCommand) Execute(_ []string) error {
	return command.runner.app.RunSnapshot(
		command.Args.Output,
		app.ProviderCollectOptions{
			WorkDir:       command.SnapshotFlags.WorkDir,
			Modules:       command.SnapshotFlags.Modules,
			SoftProviders: command.SnapshotFlags.SoftProviders,
		},
		app.SnapshotCommandOptions{
			Format: command.SnapshotFlags.Format,
			Check:  command.SnapshotFlags.Check.Check,
		},
	)
}

// docCommand renders documentation from snapshot input.
type docCommand struct {
	runner *cliRunner

	Args struct {
		Input  string `positional-arg-name:"input" description:"Input registry snapshot path (optional; stdin when omitted)"`
		Output string `positional-arg-name:"output" description:"Output documentation path (optional; stdout when omitted)"`
	} `positional-args:"yes"`

	DocFlags docFlags `group:"Doc Render"`
}

// Execute runs doc subcommand.
func (command *docCommand) Execute(_ []string) error {
	return command.runner.app.RunDoc(
		command.Args.Input,
		command.Args.Output,
		app.DocCommandOptions{
			Markdown: app.MarkdownRenderOptions{
				TemplateName:        command.DocFlags.Markdown.TemplateName,
				TemplatePath:        command.DocFlags.Markdown.TemplatePath,
				DocumentTitle:       command.DocFlags.Markdown.Title,
				DocumentDescription: command.DocFlags.Markdown.Description,
				ExampleFormat:       command.DocFlags.Markdown.ExampleFormat,
				TOCMode:             command.DocFlags.Markdown.TOCMode,
				WrapWidth:           command.DocFlags.Markdown.WrapWidth,
				FooterToolName:      "lintkit",
				FooterToolURL:       URL,
				FooterVersion:       Version,
				FooterCommit:        Commit,
			},
			Check: command.DocFlags.Check.Check,
		},
	)
}

// schemaCommand renders policy JSON Schema from snapshot input.
type schemaCommand struct {
	runner *cliRunner

	Args struct {
		Input  string `positional-arg-name:"input" description:"Input registry snapshot path (optional; stdin when omitted)"`
		Output string `positional-arg-name:"output" description:"Output schema path (optional; stdout when omitted)"`
	} `positional-args:"yes"`

	SchemaFlags schemaFlags `group:"Schema Render"`
}

// Execute runs schema subcommand.
func (command *schemaCommand) Execute(_ []string) error {
	return command.runner.app.RunSchema(
		command.Args.Input,
		command.Args.Output,
		app.SchemaCommandOptions{
			Format:   command.SchemaFlags.Format,
			Selector: command.SchemaFlags.Selector,
			Check:    command.SchemaFlags.Check.Check,
		},
	)
}

// templateCommand writes selected built-in documentation template.
type templateCommand struct {
	runner *cliRunner

	Args struct {
		Output string `positional-arg-name:"output" description:"Output template file path (optional; stdout when omitted)"`
	} `positional-args:"yes"`

	TemplateType string `short:"t" long:"template" description:"Built-in documentation template style" choice:"list" choice:"table" choice:"html" default:"list"`
}

// Execute runs template subcommand.
func (command *templateCommand) Execute(_ []string) error {
	return command.runner.app.RunTemplate(command.TemplateType, command.Args.Output)
}

// versionCommand prints version information.
type versionCommand struct{}

// Execute runs version subcommand.
func (command *versionCommand) Execute(_ []string) error {
	printVersionInfo()
	return nil
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"io"
)

// Runner executes CLI operations with custom IO streams.
type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// MarkdownRenderOptions groups markdown template and layout options.
type MarkdownRenderOptions struct {
	TemplateName        string
	TemplatePath        string
	DocumentTitle       string
	DocumentDescription string
	ExampleFormat       string
	TOCMode             string
	FooterToolName      string
	FooterToolURL       string
	FooterVersion       string
	FooterCommit        string
	WrapWidth           int
}

// ProviderCollectOptions groups provider collection input.
type ProviderCollectOptions struct {
	WorkDir       string
	Modules       []string
	SoftProviders bool
}

// SnapshotCommandOptions groups snapshot output behavior.
type SnapshotCommandOptions struct {
	EnableServiceDiagnostics *bool
	Format                   string
	Check                    bool
}

// DocCommandOptions groups markdown doc behavior.
type DocCommandOptions struct {
	Markdown MarkdownRenderOptions
	Check    bool
}

// SchemaCommandOptions groups policy schema output behavior.
type SchemaCommandOptions struct {
	Format   string
	Selector []string
	Check    bool
}

// renderOptions stores normalized render behavior.
type renderOptions struct {
	Format              string
	TemplateName        string
	TemplatePath        string
	DocumentTitle       string
	DocumentDescription string
	ExampleFormat       string
	TOCMode             string
	FooterToolName      string
	FooterToolURL       string
	FooterVersion       string
	FooterCommit        string
	WrapWidth           int
	Pretty              bool
}

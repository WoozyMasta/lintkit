// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import "github.com/woozymasta/lintkit/lint"

// View stores snapshot data prepared for markdown templates.
type View struct {
	// DocumentTitle is optional custom top-level document title.
	DocumentTitle string

	// DocumentDescription is optional custom top-level document description.
	DocumentDescription string

	// ExampleFormat is optional policy example format ("json" or "yaml").
	ExampleFormat string

	// ExampleContent is optional rendered policy example payload.
	ExampleContent string

	// FooterToolName is footer tool label.
	FooterToolName string

	// FooterToolURL is footer tool project URL.
	FooterToolURL string

	// FooterVersion is footer build version.
	FooterVersion string

	// FooterCommit is footer git revision.
	FooterCommit string

	// Rules is flat deterministic rules list used by templates.
	Rules []lint.RuleSpec

	// Modules is grouped rules view by module.
	Modules []Module

	// WrapWidth is plain text wrap width for markdown templates.
	WrapWidth int

	// ShowTOC enables module table of contents rendering.
	ShowTOC bool

	// ShowScopeTOC enables scope-level TOC rendering inside module sections.
	ShowScopeTOC bool
}

// Module stores one module section in markdown list view.
type Module struct {
	// ID is module identifier used for grouping and anchors.
	ID string

	// Name is optional human-readable module name.
	Name string

	// Description is optional module-level documentation text.
	Description string

	// Scopes is ordered module scope groups.
	Scopes []Scope

	// RuleCount is total number of module rules.
	RuleCount int
}

// Scope stores one module scope section with ordered rules.
type Scope struct {
	// ID is normalized scope identifier (for example "validate").
	ID string

	// Anchor is stable scope anchor token for TOC links.
	Anchor string

	// Description is human-readable scope documentation text.
	Description string

	// Rules is ordered rule list for this scope.
	Rules []lint.RuleSpec

	// RuleCount is total number of rules in this scope.
	RuleCount int
}

// Options stores markdown render options.
type Options struct {
	// TemplateName selects built-in template style.
	TemplateName string

	// TemplatePath points to optional external template override.
	TemplatePath string

	// DocumentTitle overrides top-level rendered document title.
	DocumentTitle string

	// DocumentDescription overrides top-level rendered description.
	DocumentDescription string

	// ExampleFormat enables optional policy example block.
	ExampleFormat string

	// TOCMode controls TOC behavior ("auto", "always", or "off").
	TOCMode string

	// FooterToolName is footer tool label.
	FooterToolName string

	// FooterToolURL is footer tool project URL.
	FooterToolURL string

	// FooterVersion is footer build version.
	FooterVersion string

	// FooterCommit is footer git revision.
	FooterCommit string

	// WrapWidth controls markdown plain text wrapping width.
	WrapWidth int
}

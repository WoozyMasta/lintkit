// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

// Options stores provider collection input.
type Options struct {
	// WorkDir is go toolchain working directory. Defaults to ".".
	WorkDir string

	// Modules is explicit provider package import list.
	Modules []string

	// StrictProviders fails on duplicate conflicts during provider registration.
	StrictProviders bool

	// Scopes selects provider rules by rule scope tokens.
	Scopes []string

	// Stages selects provider rules by stage tokens.
	Stages []string
}

// directoryState stores one source directory provider detection state.
type directoryState struct {
	// hasProviderType reports whether LintRulesProvider type exists in directory.
	hasProviderType bool

	// hasRegisterMethod reports whether RegisterRules method exists in directory.
	hasRegisterMethod bool
}

// collectorImport stores one generated provider import entry.
type collectorImport struct {
	// Alias stores deterministic generated import alias.
	Alias string

	// Path stores provider package import path.
	Path string
}

// collectorTemplateData stores template input for collector source.
type collectorTemplateData struct {
	// Imports stores generated provider package imports.
	Imports []collectorImport

	// StrictProviders controls duplicate conflict handling mode.
	StrictProviders bool

	// Scopes stores normalized scope filter tokens.
	Scopes []string

	// Stages stores normalized stage filter tokens.
	Stages []string
}

// listedPackage stores minimal go list package payload used by discovery.
type listedPackage struct {
	// ImportPath stores package import path.
	ImportPath string

	// Dir stores absolute package directory path.
	Dir string

	// GoFiles stores package .go file basenames, without _test.go files.
	GoFiles []string

	// Standard reports whether package belongs to Go standard library.
	Standard bool
}

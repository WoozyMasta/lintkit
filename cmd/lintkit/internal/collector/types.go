// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

// Options stores provider collection input.
type Options struct {
	// WorkDir is go toolchain working directory. Defaults to ".".
	WorkDir string

	// CollectorTempDir stores optional directory for generated collector source.
	// Empty value means system temporary directory.
	CollectorTempDir string

	// Modules is explicit provider package import list.
	Modules []string

	// Scopes selects provider rules by rule scope tokens.
	Scopes []string

	// Stages selects provider rules by stage tokens.
	Stages []string

	// KeepCollector keeps generated collector source file for diagnostics.
	KeepCollector bool

	// IncludeLintkitRules keeps built-in lintkit provider in auto-discovery.
	IncludeLintkitRules bool

	// StrictProviders fails on duplicate conflicts during provider registration.
	StrictProviders bool
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

	// Scopes stores normalized scope filter tokens.
	Scopes []string

	// Stages stores normalized stage filter tokens.
	Stages []string

	// StrictProviders controls duplicate conflict handling mode.
	StrictProviders bool
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

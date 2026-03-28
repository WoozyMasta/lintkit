// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectProviderImportPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	providerDir := filepath.Join(root, "module", "linting")
	if err := os.MkdirAll(providerDir, 0o750); err != nil {
		t.Fatalf("mkdir provider dir: %v", err)
	}

	providerFile := filepath.Join(providerDir, "lint_rules.go")
	providerBody := "package linting\n\n" +
		"type LintRulesProvider struct{}\n\n" +
		"func (provider LintRulesProvider) RegisterRules(_ any) error { return nil }\n"
	if err := os.WriteFile(providerFile, []byte(providerBody), 0o600); err != nil {
		t.Fatalf("write provider file: %v", err)
	}

	otherDir := filepath.Join(root, "module", "other")
	if err := os.MkdirAll(otherDir, 0o750); err != nil {
		t.Fatalf("mkdir other dir: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(otherDir, "x.go"),
		[]byte("package other\n"),
		0o600,
	); err != nil {
		t.Fatalf("write other file: %v", err)
	}

	got := detectProviderImportPaths([]listedPackage{
		{
			ImportPath: "example.com/test/module/linting",
			Dir:        providerDir,
			GoFiles:    []string{"lint_rules.go"},
		},
		{
			ImportPath: "example.com/test/module/other",
			Dir:        otherDir,
			GoFiles:    []string{"x.go"},
		},
	})

	if len(got) != 1 {
		t.Fatalf("len(detectProviderImportPaths())=%d, want 1", len(got))
	}

	if got[0] != "example.com/test/module/linting" {
		t.Fatalf(
			"detectProviderImportPaths()[0]=%q, want %q",
			got[0],
			"example.com/test/module/linting",
		)
	}
}

func TestResolvePackagesRequiresGoModuleForDiscovery(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	_, err := ResolvePackages(root, nil)
	if !errors.Is(err, ErrNoGoModuleInWorkDir) {
		t.Fatalf("ResolvePackages() error=%v, want ErrNoGoModuleInWorkDir", err)
	}
}

func TestResolvePackagesWithModulesSkipsDiscovery(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packages, err := ResolvePackages(root, []string{
		"github.com/example/a",
		"github.com/example/a",
		"github.com/example/b",
	})
	if err != nil {
		t.Fatalf("ResolvePackages(modules) error: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("len(ResolvePackages(modules))=%d, want 2", len(packages))
	}

	if packages[0] != "github.com/example/a" || packages[1] != "github.com/example/b" {
		t.Fatalf("ResolvePackages(modules)=%v, want sorted unique values", packages)
	}
}

func TestDetectProviderImportPathsPointerReceiver(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	providerDir := filepath.Join(root, "module", "linting")
	if err := os.MkdirAll(providerDir, 0o750); err != nil {
		t.Fatalf("mkdir provider dir: %v", err)
	}

	providerFile := filepath.Join(providerDir, "lint_rules.go")
	providerBody := "package linting\n\n" +
		"type LintRulesProvider struct{}\n\n" +
		"func (provider *LintRulesProvider) RegisterRules(_ any) error { return nil }\n"
	if err := os.WriteFile(providerFile, []byte(providerBody), 0o600); err != nil {
		t.Fatalf("write provider file: %v", err)
	}

	got := detectProviderImportPaths([]listedPackage{
		{
			ImportPath: "example.com/test/module/linting",
			Dir:        providerDir,
			GoFiles:    []string{"lint_rules.go"},
		},
	})
	if len(got) != 1 {
		t.Fatalf("len(detectProviderImportPaths())=%d, want 1", len(got))
	}

	if got[0] != "example.com/test/module/linting" {
		t.Fatalf(
			"detectProviderImportPaths()[0]=%q, want %q",
			got[0],
			"example.com/test/module/linting",
		)
	}
}

func TestBuildCollectorProgramSupportsPointerFallback(t *testing.T) {
	t.Parallel()

	program, err := BuildCollectorProgram(
		[]string{"example.com/test/module/linting"},
		true,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("BuildCollectorProgram() error: %v", err)
	}

	text := string(program)
	if !strings.Contains(text, "any(&provider_0Provider).(lint.RuleProvider)") {
		t.Fatalf("generated collector missing pointer fallback:\n%s", text)
	}

	if !strings.Contains(text, "reflect.DeepEqual(left.DefaultOptions, right.DefaultOptions)") {
		t.Fatalf("generated collector missing DefaultOptions compare:\n%s", text)
	}

	if !strings.Contains(text, "lint.RegisterRuleProviders(registrar, providers...)") {
		t.Fatalf("generated collector missing RegisterRuleProviders usage:\n%s", text)
	}
}

func TestBuildCollectorProgramScopeFilters(t *testing.T) {
	t.Parallel()

	program, err := BuildCollectorProgram(
		[]string{"example.com/test/module/linting"},
		true,
		[]string{"parse", "validate"},
		nil,
	)
	if err != nil {
		t.Fatalf("BuildCollectorProgram(scope) error: %v", err)
	}

	text := string(program)
	if !strings.Contains(text, "lint.RegisterRuleProvidersByScope(") {
		t.Fatalf("generated collector missing scope registration helper:\n%s", text)
	}

	if !strings.Contains(text, "\"parse\"") || !strings.Contains(text, "\"validate\"") {
		t.Fatalf("generated collector missing scope filters:\n%s", text)
	}
}

func TestBuildCollectorProgramStageFilters(t *testing.T) {
	t.Parallel()

	program, err := BuildCollectorProgram(
		[]string{"example.com/test/module/linting"},
		true,
		nil,
		[]string{"parse"},
	)
	if err != nil {
		t.Fatalf("BuildCollectorProgram(stage) error: %v", err)
	}

	text := string(program)
	if !strings.Contains(text, "lint.RegisterRuleProvidersByStage(") {
		t.Fatalf("generated collector missing stage registration helper:\n%s", text)
	}

	if !strings.Contains(text, "lint.Stage(\"parse\")") {
		t.Fatalf("generated collector missing stage filters:\n%s", text)
	}
}

func TestNormalizeOptionsConflictingFilters(t *testing.T) {
	t.Parallel()

	_, err := normalizeOptions(Options{
		WorkDir: ".",
		Scopes:  []string{"parse"},
		Stages:  []string{"parse"},
	})
	if !errors.Is(err, ErrConflictingRuleFilters) {
		t.Fatalf("normalizeOptions(conflicting filters) error=%v, want ErrConflictingRuleFilters", err)
	}
}

func TestNormalizeOptionsFilterTokens(t *testing.T) {
	t.Parallel()

	options, err := normalizeOptions(Options{
		WorkDir: ".",
		Scopes:  []string{" parse ", "validate", "parse"},
	})
	if err != nil {
		t.Fatalf("normalizeOptions(scopes) error: %v", err)
	}

	if len(options.Scopes) != 2 || options.Scopes[0] != "parse" || options.Scopes[1] != "validate" {
		t.Fatalf("normalizeOptions(scopes)=%v, want [parse validate]", options.Scopes)
	}
}

func TestDetectProviderImportPathsRejectsInvalidRegisterSignature(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	providerDir := filepath.Join(root, "module", "linting")
	if err := os.MkdirAll(providerDir, 0o750); err != nil {
		t.Fatalf("mkdir provider dir: %v", err)
	}

	providerFile := filepath.Join(providerDir, "lint_rules.go")
	providerBody := "package linting\n\n" +
		"type LintRulesProvider struct{}\n\n" +
		"func (provider LintRulesProvider) RegisterRules() {}\n"
	if err := os.WriteFile(providerFile, []byte(providerBody), 0o600); err != nil {
		t.Fatalf("write provider file: %v", err)
	}

	got := detectProviderImportPaths([]listedPackage{
		{
			ImportPath: "example.com/test/module/linting",
			Dir:        providerDir,
			GoFiles:    []string{"lint_rules.go"},
		},
	})
	if len(got) != 0 {
		t.Fatalf("len(detectProviderImportPaths())=%d, want 0", len(got))
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

// TestRunDocFromStdinToStdout verifies markdown list rendering from stdin.
func TestRunDocFromStdinToStdout(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-"},
		bytes.NewReader(input),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc stdin/stdout) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "# Lint Rules Registry") {
		t.Fatalf("expected markdown header, got: %q", text)
	}

	if !strings.Contains(text, "#### `RVCFG2020`") {
		t.Fatalf("expected code-based rule heading in markdown output, got: %q", text)
	}

	if !strings.Contains(text, "* Rule ID: `module_alpha.parse.trailing-comma`") {
		t.Fatalf("expected rule id attribute in markdown output, got: %q", text)
	}

	if strings.Contains(text, "## Contents") {
		t.Fatalf("did not expect TOC for single-module snapshot, got: %q", text)
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr output: %q", stderr.String())
	}
}

// TestRunDocTableTemplate verifies built-in markdown table rendering.
func TestRunDocTableTemplate(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--template", "table"},
		bytes.NewReader(input),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc table) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "| Field | Value |") {
		t.Fatalf("expected field table header, got: %q", text)
	}

	if !strings.Contains(text, "| Rule ID | `module_alpha.parse.trailing-comma` |") {
		t.Fatalf("expected rule id attribute row in markdown output, got: %q", text)
	}
}

// TestRunDocHTMLTemplate verifies built-in HTML rendering.
func TestRunDocHTMLTemplate(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--template", "html"},
		bytes.NewReader(input),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc html) exit=%d", exitCode)
	}

	text := strings.ToLower(stdout.String())
	if !strings.Contains(text, "<!doctype html>") {
		t.Fatalf("expected html doctype, got: %q", stdout.String())
	}

	if !strings.Contains(text, "lint rules registry") {
		t.Fatalf("expected html title text, got: %q", stdout.String())
	}
}

// TestRunDocHTMLTemplateRichDescription verifies code/list rich rendering.
func TestRunDocHTMLTemplateRichDescription(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSONWithRules(t, []lint.RuleSpec{
		{
			ID:              "module_alpha.validate.rule-rich",
			Module:          "module_alpha",
			Code:            "RVCFG2999",
			Message:         "rich description",
			Description:     "check `token` format\n- first item\n* second item\n1. third item",
			DefaultSeverity: lint.SeverityWarning,
		},
	})

	var stdout bytes.Buffer
	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--template", "html"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc html rich) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "<code>token</code>") {
		t.Fatalf("expected inline code in rich description, got: %q", text)
	}

	if !strings.Contains(text, "<ul><li>first item</li><li>second item</li><li>third item</li></ul>") {
		t.Fatalf("expected list rendering in rich description, got: %q", text)
	}
}

// TestRunDocWithJSONExample verifies markdown with embedded JSON example.
func TestRunDocWithJSONExample(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--example-format", "json"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc + json example) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "## Lint Policy Example (`json`)") {
		t.Fatalf("missing JSON example section: %q", text)
	}

	if !strings.Contains(text, "\"rules\":") {
		t.Fatalf("missing JSON example payload: %q", text)
	}
}

// TestRunDocWithYAMLExample verifies markdown with YAML example.
func TestRunDocWithYAMLExample(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--example-format", "yaml"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc + yaml example) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "## Lint Policy Example (`yaml`)") {
		t.Fatalf("missing YAML example section: %q", text)
	}

	if !strings.Contains(text, "rules:") {
		t.Fatalf("missing YAML example payload: %q", text)
	}

	if !strings.Contains(text, "rule: RVCFG2020") {
		t.Fatalf("missing code selector in YAML example: %q", text)
	}
}

// TestRunDocCustomTemplate verifies external markdown template rendering.
func TestRunDocCustomTemplate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	templatePath := filepath.Join(dir, "custom.gotmpl")
	if err := os.WriteFile(
		templatePath,
		[]byte("{{ len .Rules }} rules\n"),
		0o600,
	); err != nil {
		t.Fatalf("write custom template: %v", err)
	}

	input := testSnapshotJSON(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--template-file", templatePath},
		bytes.NewReader(input),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc custom template) exit=%d", exitCode)
	}

	if got := stdout.String(); got != "1 rules\n" {
		t.Fatalf("custom template output=%q, want %q", got, "1 rules\n")
	}
}

// TestRunDocTOCAlways verifies TOC rendering when enabled explicitly.
func TestRunDocTOCAlways(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSONTwoModules(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--toc", "always"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc + toc=always) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "* [module_alpha](#module_alpha)") {
		t.Fatalf("missing module_alpha TOC item, got: %q", text)
	}

	if !strings.Contains(text, "* [module_beta](#module_beta)") {
		t.Fatalf("missing module_beta TOC item, got: %q", text)
	}
}

// TestRunDocModuleMetadata verifies module name/description output.
func TestRunDocModuleMetadata(t *testing.T) {
	t.Parallel()

	snapshot := lint.RegistrySnapshot{
		Modules: []lint.ModuleSpec{
			{
				ID:          "module_alpha",
				Name:        "Module Alpha",
				Description: "Rules for module_alpha parser.",
			},
		},
		Rules: []lint.RuleSpec{
			{
				ID:               "module_alpha.parse.trailing-comma",
				Module:           "module_alpha",
				Scope:            "parse",
				ScopeDescription: "Parser diagnostics.",
				Code:             "RVCFG2020",
				Message:          "trailing comma",
				Description:      "Warn when trailing comma is used.",
				DefaultSeverity:  lint.SeverityWarning,
			},
		},
	}

	input, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := runWithIO(
		[]string{"doc", "-", "-"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, "## module_alpha") {
		t.Fatalf("missing module heading: %q", text)
	}

	if !strings.Contains(text, "Module Alpha") {
		t.Fatalf("missing module name: %q", text)
	}

	if !strings.Contains(text, "> Rules for module_alpha parser.") {
		t.Fatalf("missing module description: %q", text)
	}

	if !strings.Contains(text, "* Scope: `parse`") {
		t.Fatalf("missing rule scope field: %q", text)
	}
}

// TestRunDocWrap verifies wrap width handling for description fields.
func TestRunDocWrap(t *testing.T) {
	t.Parallel()

	description := "one two three four five six seven eight nine ten"
	expectedBlockQuote := strings.Join([]string{
		"> one two three four",
		"> five six seven eight",
		"> nine ten",
	}, "\n")

	input := testSnapshotJSONWithRules(t, []lint.RuleSpec{
		{
			ID:              "module_alpha.parse.long-description",
			Module:          "module_alpha",
			Code:            "RVCFG2021",
			Message:         "long description",
			Description:     description,
			DefaultSeverity: lint.SeverityWarning,
		},
	})
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"doc", "-", "-", "--wrap", "20"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc + wrap) exit=%d", exitCode)
	}

	text := stdout.String()
	if !strings.Contains(text, expectedBlockQuote) {
		t.Fatalf("expected wrapped description blockquote, got: %q", text)
	}
}

// TestRunDocWriteAndCheck verifies doc write and --check flow.
func TestRunDocWriteAndCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "snapshot.json")
	outputPath := filepath.Join(dir, "docs", "rules.md")
	if err := os.WriteFile(inputPath, testSnapshotJSON(t), 0o600); err != nil {
		t.Fatalf("write input snapshot: %v", err)
	}

	exitCode := runWithIO(
		[]string{"doc", inputPath, outputPath},
		bytes.NewReader(nil),
		bytes.NewBuffer(nil),
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc write) exit=%d", exitCode)
	}

	exitCode = runWithIO(
		[]string{"doc", inputPath, outputPath, "--check"},
		bytes.NewReader(nil),
		bytes.NewBuffer(nil),
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(doc check equal) exit=%d", exitCode)
	}

	if err := os.WriteFile(outputPath, []byte("changed\n"), 0o600); err != nil {
		t.Fatalf("mutate output file: %v", err)
	}

	exitCode = runWithIO(
		[]string{"doc", inputPath, outputPath, "--check"},
		bytes.NewReader(nil),
		bytes.NewBuffer(nil),
		bytes.NewBuffer(nil),
	)
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code on --check diff")
	}
}

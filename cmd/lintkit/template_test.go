// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRunTemplate verifies built-in template export command.
func TestRunTemplate(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"template", "--template", "table"},
		bytes.NewReader(nil),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(template) exit=%d", exitCode)
	}

	if !strings.Contains(stdout.String(), "| Field | Value |") {
		t.Fatalf("template output mismatch: %q", stdout.String())
	}
}

// TestRunTemplateHTML verifies built-in HTML template export command.
func TestRunTemplateHTML(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runWithIO(
		[]string{"template", "--template", "html"},
		bytes.NewReader(nil),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(template html) exit=%d", exitCode)
	}

	text := strings.ToLower(stdout.String())
	if !strings.Contains(text, "<!doctype html>") {
		t.Fatalf("html template output mismatch: %q", stdout.String())
	}
}

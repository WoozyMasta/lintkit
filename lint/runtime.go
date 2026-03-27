// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "context"

// DiagnosticEmit appends one diagnostic produced by one rule runner.
type DiagnosticEmit func(diagnostic Diagnostic)

// RunContext stores shared execution context for rule checks.
type RunContext struct {
	// Values stores optional caller-provided shared mutable state.
	Values map[string]any `json:"values,omitempty" yaml:"values,omitempty" jsonschema:"additionalProperties=true"`

	// RootDir is optional filesystem root for rules that need path resolution.
	RootDir string `json:"root_dir,omitempty" yaml:"root_dir,omitempty" jsonschema:"example=./"`

	// TargetPath is optional current target path.
	TargetPath string `json:"target_path,omitempty" yaml:"target_path,omitempty" jsonschema:"example=workspace/main/source.cfg"`

	// TargetKind is optional current target kind (for example source.cfg, module_beta).
	TargetKind FileKind `json:"target_kind,omitempty" yaml:"target_kind,omitempty" jsonschema:"example=source.cfg,example=module_beta"`

	// Content stores optional raw bytes for in-memory checks.
	Content []byte `json:"content,omitempty" yaml:"content,omitempty" jsonschema:"description=Raw in-memory content bytes for non-filesystem lint flows."`

	// TargetIsDir reports whether current target path points to directory.
	TargetIsDir bool `json:"target_is_dir,omitempty" yaml:"target_is_dir,omitempty" jsonschema:"default=false"`
}

// RuleRunner executes one rule against one runtime context.
type RuleRunner interface {
	// RuleSpec returns stable metadata descriptor for this rule.
	RuleSpec() RuleSpec

	// Check runs one rule and emits zero or more diagnostics.
	Check(ctx context.Context, run *RunContext, emit DiagnosticEmit) error
}

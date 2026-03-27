// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"errors"

	"github.com/woozymasta/lintkit/lint"
)

var (
	// ErrNilRegistry indicates nil registry receiver usage.
	ErrNilRegistry = errors.New("registry is nil")

	// ErrInvalidRuleSpec indicates malformed rule metadata.
	ErrInvalidRuleSpec = errors.New("invalid rule spec")

	// ErrInvalidModuleSpec indicates malformed module metadata.
	ErrInvalidModuleSpec = errors.New("invalid module spec")

	// ErrDuplicateRuleID indicates duplicate rule ID registration.
	ErrDuplicateRuleID = errors.New("duplicate rule id")

	// ErrDuplicateRuleCode indicates duplicate non-empty rule code registration.
	ErrDuplicateRuleCode = errors.New("duplicate rule code")

	// ErrConflictingModuleSpec indicates conflicting metadata for one module.
	ErrConflictingModuleSpec = errors.New("conflicting module spec")

	// ErrNilEngine indicates nil engine receiver usage.
	ErrNilEngine = errors.New("engine is nil")

	// ErrNilRuleRunner indicates nil rule runner registration.
	ErrNilRuleRunner = errors.New("rule runner is nil")

	// ErrRuleRunnerPanic indicates recovered panic from rule runner.
	ErrRuleRunnerPanic = errors.New("rule runner panic")

	// ErrNilRuleProvider indicates nil rule provider registration.
	ErrNilRuleProvider = lint.ErrNilRuleProvider

	// ErrUnknownRuleID indicates explicit run target references unknown rule.
	ErrUnknownRuleID = errors.New("unknown rule id")

	// ErrInvalidRunPolicy indicates malformed run policy settings.
	ErrInvalidRunPolicy = errors.New("invalid run policy")

	// ErrUnknownRuleSelector indicates unknown policy rule selector.
	ErrUnknownRuleSelector = errors.New("unknown rule selector")

	// ErrAmbiguousRuleSelector indicates selector that matches multiple rules.
	ErrAmbiguousRuleSelector = errors.New("ambiguous rule selector")

	// ErrNilRunPolicy indicates nil run policy receiver usage.
	ErrNilRunPolicy = errors.New("run policy is nil")

	// ErrNilPolicyOverride indicates nil policy override receiver usage.
	ErrNilPolicyOverride = errors.New("policy override is nil")

	// ErrNilPatternMatcherCompiler indicates missing matcher compiler callback.
	ErrNilPatternMatcherCompiler = errors.New("pattern matcher compiler is nil")

	// ErrInvalidFailSeverity indicates unsupported fail-on threshold severity.
	ErrInvalidFailSeverity = errors.New("invalid fail severity")
)

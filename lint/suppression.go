// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "strings"

const (
	// SuppressionScopeLine suppresses one diagnostic line span.
	SuppressionScopeLine SuppressionScope = "line"

	// SuppressionScopeBlock suppresses one multi-line block span.
	SuppressionScopeBlock SuppressionScope = "block"

	// SuppressionScopeFile suppresses diagnostics for entire file.
	SuppressionScopeFile SuppressionScope = "file"
)

// SuppressionScope defines normalized suppression scope granularity.
type SuppressionScope string

// SuppressionDecision stores suppression decision and optional audit metadata.
type SuppressionDecision struct {
	// Scope stores suppression scope for standardized audit/behavior.
	Scope SuppressionScope `json:"scope,omitempty" yaml:"scope,omitempty" jsonschema:"enum=line,enum=block,enum=file,default=line,example=line"`

	// Reason stores optional human-readable suppression reason.
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty" jsonschema:"example=legacy file, ignore until migration"`

	// Source stores optional suppression source (inline comment/rule file).
	Source string `json:"source,omitempty" yaml:"source,omitempty" jsonschema:"example=inline,example=.lintignore"`

	// ExpiresAt stores optional suppression expiration timestamp (RFC3339).
	ExpiresAt string `json:"expires_at,omitempty" yaml:"expires_at,omitempty" jsonschema:"format=date-time,example=2026-12-31T23:59:59Z"`
	// Suppressed reports whether diagnostic should be filtered out.
	Suppressed bool `json:"suppressed" yaml:"suppressed" jsonschema:"required,default=false"`
}

// SuppressionHit stores one suppressed diagnostic audit event.
type SuppressionHit struct {
	// RuleID is suppressed rule identifier.
	RuleID string `json:"rule_id,omitempty" yaml:"rule_id,omitempty" jsonschema:"example=module_alpha.parse.macro-redefinition"`

	// Path is suppressed diagnostic path.
	Path string `json:"path,omitempty" yaml:"path,omitempty" jsonschema:"example=workspace/main/source.cfg"`

	// Scope stores normalized suppression scope for audit/reporting.
	Scope SuppressionScope `json:"scope,omitempty" yaml:"scope,omitempty" jsonschema:"enum=line,enum=block,enum=file,default=line,example=line"`

	// Reason stores optional human-readable suppression reason.
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty" jsonschema:"example=legacy file, ignore until migration"`

	// Source stores optional suppression source (inline comment/rule file).
	Source string `json:"source,omitempty" yaml:"source,omitempty" jsonschema:"example=inline,example=.lintignore"`

	// ExpiresAt stores optional suppression expiration timestamp (RFC3339).
	ExpiresAt string `json:"expires_at,omitempty" yaml:"expires_at,omitempty" jsonschema:"format=date-time,example=2026-12-31T23:59:59Z"`

	// Start is optional start position.
	Start Position `json:"start,omitzero" yaml:"start,omitempty"`

	// End is optional end position.
	End Position `json:"end,omitzero" yaml:"end,omitempty"`
}

// SuppressionSet checks whether one diagnostic is suppressed by context.
type SuppressionSet interface {
	// DecideSuppression returns suppression decision for one diagnostic.
	DecideSuppression(
		ruleID string,
		path string,
		start Position,
		end Position,
	) SuppressionDecision
}

// SuppressionSetFunc adapts bool callback values to SuppressionSet interface.
type SuppressionSetFunc func(
	ruleID string,
	path string,
	start Position,
	end Position,
) bool

// DecideSuppression delegates suppression decision to bool callback.
func (fn SuppressionSetFunc) DecideSuppression(
	ruleID string,
	path string,
	start Position,
	end Position,
) SuppressionDecision {
	if fn == nil {
		return SuppressionDecision{}
	}

	suppressed := fn(ruleID, path, start, end)
	if !suppressed {
		return SuppressionDecision{}
	}

	return SuppressionDecision{
		Suppressed: true,
		Scope:      SuppressionScopeLine,
	}
}

// SuppressionDecisionFunc adapts decision callbacks to SuppressionSet interface.
type SuppressionDecisionFunc func(
	ruleID string,
	path string,
	start Position,
	end Position,
) SuppressionDecision

// DecideSuppression delegates suppression decision callback.
func (fn SuppressionDecisionFunc) DecideSuppression(
	ruleID string,
	path string,
	start Position,
	end Position,
) SuppressionDecision {
	if fn == nil {
		return SuppressionDecision{}
	}

	decision := fn(ruleID, path, start, end)
	if !decision.Suppressed {
		return decision
	}

	decision.Scope = NormalizeSuppressionScope(decision.Scope)
	return decision
}

// NormalizeSuppressionScope returns normalized suppression scope value.
func NormalizeSuppressionScope(scope SuppressionScope) SuppressionScope {
	normalized := strings.ToLower(strings.TrimSpace(string(scope)))
	switch SuppressionScope(normalized) {
	case SuppressionScopeLine:
		return SuppressionScopeLine
	case SuppressionScopeBlock:
		return SuppressionScopeBlock
	case SuppressionScopeFile:
		return SuppressionScopeFile
	default:
		return SuppressionScopeLine
	}
}

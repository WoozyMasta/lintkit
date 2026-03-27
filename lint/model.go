// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

const (
	// SeverityError marks hard lint failures.
	SeverityError Severity = "error"

	// SeverityWarning marks non-fatal findings.
	SeverityWarning Severity = "warning"

	// SeverityInfo marks informational findings.
	SeverityInfo Severity = "info"

	// SeverityNotice marks low-priority notices.
	SeverityNotice Severity = "notice"
)

// Severity defines normalized lint diagnostic severity.
type Severity string

// Code defines module-local stable numeric lint code token.
type Code uint32

// Stage defines optional lint pipeline stage token.
type Stage string

// CodeSpec stores reusable lint code metadata shared by module catalogs.
type CodeSpec struct {
	// Rule stores optional explicit catalog-level rule metadata override.
	// Only non-runtime metadata fields are allowed here.
	Rule *CodeRuleOverride `json:"rule,omitempty" yaml:"rule,omitempty"`

	// Options stores optional module-defined settings payload for this code.
	Options any `json:"options,omitempty" yaml:"options,omitempty" jsonschema:"description=Arbitrary module-defined code options payload."`

	// Enabled controls default rule enable state for this code.
	// Nil means enabled by default.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty" jsonschema:"default=true,example=true"`

	// Stage is optional lint pipeline stage token.
	Stage Stage `json:"stage,omitempty" yaml:"stage,omitempty"`

	// Severity is default lint level for this code.
	Severity Severity `json:"severity,omitempty" yaml:"severity,omitempty"`

	// Message is short human-readable diagnostic text.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`

	// Description is optional detailed rule description for documentation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Code is stable machine-readable lint code token.
	Code Code `json:"code,omitempty" yaml:"code,omitempty" jsonschema:"minimum=1,example=2001"`
}

// CodeRuleOverride stores optional narrow rule metadata overrides for catalog rows.
type CodeRuleOverride struct {
	// ID overrides generated stable rule identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty" jsonschema:"example=module_alpha.parse.macro-redefinition"`

	// Scope overrides generated scope token.
	Scope string `json:"scope,omitempty" yaml:"scope,omitempty" jsonschema:"example=parse,example=validate"`

	// ScopeDescription overrides generated scope description.
	ScopeDescription string `json:"scope_description,omitempty" yaml:"scope_description,omitempty" jsonschema:"example=Parser diagnostics.,example=Semantic validation diagnostics."`

	// Code overrides generated public code token.
	Code string `json:"code,omitempty" yaml:"code,omitempty" jsonschema:"example=RVCFG2001,example=RVMAT2028"`

	// FileKinds overrides supported file kind list.
	FileKinds []FileKind `json:"file_kinds,omitempty" yaml:"file_kinds,omitempty" jsonschema:"uniqueItems=true,example=source.cfg,example=module_beta"`

	// Deprecated marks rule as deprecated.
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty" jsonschema:"default=false"`
}

// Position stores one source position in file-oriented diagnostics.
type Position struct {
	// File is source file path or logical source name.
	File string `json:"file,omitempty" yaml:"file,omitempty" jsonschema:"example=workspace/main/source.cfg"`

	// Line is 1-based source line.
	Line int `json:"line,omitempty" yaml:"line,omitempty" jsonschema:"minimum=1,example=12"`

	// Column is 1-based source column.
	Column int `json:"column,omitempty" yaml:"column,omitempty" jsonschema:"minimum=1,example=5"`

	// Offset is 0-based byte offset from source start.
	Offset int `json:"offset,omitempty" yaml:"offset,omitempty" jsonschema:"minimum=0,example=128"`
}

// Diagnostic stores one normalized lint finding.
type Diagnostic struct {
	// RuleID is stable machine-readable rule identifier.
	RuleID string `json:"rule_id,omitempty" yaml:"rule_id,omitempty" jsonschema:"minLength=3,example=module_alpha.parse.macro-redefinition"`

	// Code is optional short diagnostic code token.
	Code string `json:"code,omitempty" yaml:"code,omitempty" jsonschema:"example=RVCFG2001,example=LINTKIT1001"`

	// Severity is effective lint level for this finding.
	Severity Severity `json:"severity,omitempty" yaml:"severity,omitempty" jsonschema:"enum=error,enum=warning,enum=info,enum=notice,example=warning"`

	// Message is human-readable finding text.
	Message string `json:"message,omitempty" yaml:"message,omitempty" jsonschema:"minLength=1,example=macro redefinition"`

	// Path is optional file path related to finding.
	Path string `json:"path,omitempty" yaml:"path,omitempty" jsonschema:"example=workspace/main/source.cfg"`

	// Start is optional start position.
	Start Position `json:"start,omitzero" yaml:"start,omitempty"`

	// End is optional end position.
	End Position `json:"end,omitzero" yaml:"end,omitempty"`
}

// ModuleSpec stores one lint module metadata descriptor.
type ModuleSpec struct {
	// ID is stable machine-readable module namespace token.
	ID string `json:"id,omitempty" yaml:"id,omitempty" jsonschema:"required,minLength=1,example=module_alpha,example=module_gamma"`

	// Name is optional human-readable module display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty" jsonschema:"example=Module Alpha Parser"`

	// Description is optional module-level documentation text.
	Description string `json:"description,omitempty" yaml:"description,omitempty" jsonschema:"example=Rules for module_alpha parse and preprocess flows."`
}

// RuleSpec stores one stable lint rule metadata descriptor.
type RuleSpec struct {
	// ID is globally-unique stable rule identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty" jsonschema:"required,minLength=3,example=module_alpha.parse.macro-redefinition,example=module_gamma.validate.asset-exists"`

	// Module is owner module namespace.
	Module string `json:"module,omitempty" yaml:"module,omitempty" jsonschema:"required,minLength=1,example=module_alpha,example=module_gamma"`

	// Scope is optional module-defined rule scope or stage token.
	Scope string `json:"scope,omitempty" yaml:"scope,omitempty" jsonschema:"example=parse,example=validate"`

	// ScopeDescription is human-readable scope documentation text.
	// When Scope is set, ScopeDescription must also be set.
	ScopeDescription string `json:"scope_description,omitempty" yaml:"scope_description,omitempty" jsonschema:"example=Parser diagnostics.,example=Semantic validation diagnostics."`

	// Code is optional exported lint code token.
	Code string `json:"code,omitempty" yaml:"code,omitempty" jsonschema:"example=RVCFG2001,example=RVMAT2028"`

	// Message is short human-readable diagnostic message.
	Message string `json:"message,omitempty" yaml:"message,omitempty" jsonschema:"required,minLength=1,example=macro redefinition"`

	// Description is detailed rule description.
	Description string `json:"description,omitempty" yaml:"description,omitempty" jsonschema:"example=Warn when macro is redefined."`

	// DefaultSeverity is default lint level when no override is configured.
	DefaultSeverity Severity `json:"default_severity,omitempty" yaml:"default_severity,omitempty" jsonschema:"default=warning,enum=error,enum=warning,enum=info,enum=notice"`

	// DefaultEnabled controls default runtime rule enable state.
	// Nil means enabled by default.
	DefaultEnabled *bool `json:"default_enabled,omitempty" yaml:"default_enabled,omitempty" jsonschema:"default=true,example=true"`

	// DefaultOptions stores optional default rule options payload.
	DefaultOptions any `json:"default_options,omitempty" yaml:"default_options,omitempty" jsonschema:"description=Arbitrary default rule options payload for runner-specific behavior."`

	// FileKinds lists supported file kinds for this rule.
	FileKinds []FileKind `json:"file_kinds,omitempty" yaml:"file_kinds,omitempty" jsonschema:"uniqueItems=true,example=source.cfg,example=module_beta"`

	// Deprecated marks deprecated rule definitions.
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty" jsonschema:"default=false"`
}

// RegistrySnapshot is a deterministic export payload of registry rules.
type RegistrySnapshot struct {
	// Modules stores sorted registry module metadata.
	Modules []ModuleSpec `json:"modules,omitempty" yaml:"modules,omitempty"`

	// Rules stores sorted registry rule metadata.
	Rules []RuleSpec `json:"rules,omitempty" yaml:"rules,omitempty"`
}

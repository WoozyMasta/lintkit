// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"testing"
)

func TestApplyAndRebaseCodePrefix(t *testing.T) {
	t.Parallel()

	if got := ApplyCodePrefix("RVCFG", 2001); got != "RVCFG2001" {
		t.Fatalf("ApplyCodePrefix()=%q, want RVCFG2001", got)
	}

	if got := RebaseCodePrefix("MYAPP", "RVCFG2001"); got != "MYAPP2001" {
		t.Fatalf("RebaseCodePrefix(prefixed)=%q, want MYAPP2001", got)
	}

	if got := RebaseCodePrefix("MYAPP", "2001"); got != "MYAPP2001" {
		t.Fatalf("RebaseCodePrefix(raw)=%q, want MYAPP2001", got)
	}

	if got := RebaseCodePrefix("MYAPP", "P3D9001"); got != "MYAPP9001" {
		t.Fatalf("RebaseCodePrefix(P3D9001)=%q, want MYAPP9001", got)
	}
}

func TestParsePublicCode(t *testing.T) {
	t.Parallel()

	code, ok := ParsePublicCode("RVCFG2001")
	if !ok || code != 2001 {
		t.Fatalf("ParsePublicCode(RVCFG2001)=%d,%v, want 2001,true", code, ok)
	}

	code, ok = ParsePublicCode("2002")
	if !ok || code != 2002 {
		t.Fatalf("ParsePublicCode(2002)=%d,%v, want 2002,true", code, ok)
	}

	code, ok = ParsePublicCode("P3D9001")
	if !ok || code != 9001 {
		t.Fatalf("ParsePublicCode(P3D9001)=%d,%v, want 9001,true", code, ok)
	}

	code, ok = ParsePublicCode("RVCFG#2001")
	if ok || code != 0 {
		t.Fatalf("ParsePublicCode(RVCFG#2001)=%d,%v, want 0,false", code, ok)
	}
}

func TestSeverityHelpers(t *testing.T) {
	t.Parallel()

	if !IsSupportedSeverity(SeverityError) {
		t.Fatal("IsSupportedSeverity(error)=false, want true")
	}

	if IsSupportedSeverity(Severity("fatal")) {
		t.Fatal("IsSupportedSeverity(fatal)=true, want false")
	}

	if SeverityRank(SeverityError) <= SeverityRank(SeverityWarning) {
		t.Fatal("SeverityRank(error) <= SeverityRank(warning), want higher")
	}
}

func TestValidateCodePrefix(t *testing.T) {
	t.Parallel()

	if err := ValidateCodePrefix("RVCFG"); err != nil {
		t.Fatalf("ValidateCodePrefix(RVCFG) error: %v", err)
	}

	if err := ValidateCodePrefix("P3D"); err != nil {
		t.Fatalf("ValidateCodePrefix(P3D) error: %v", err)
	}

	err := ValidateCodePrefix("3D")
	if !errors.Is(err, ErrInvalidCodePrefix) {
		t.Fatalf("ValidateCodePrefix(3D) error=%v, want ErrInvalidCodePrefix", err)
	}
}

func TestNewCodeCatalogRejectsInvalidPrefix(t *testing.T) {
	t.Parallel()

	_, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "3D",
	}, nil)
	if !errors.Is(err, ErrInvalidCodeCatalogConfig) {
		t.Fatalf("NewCodeCatalog() error=%v, want ErrInvalidCodeCatalogConfig", err)
	}
}

func TestNewCodeCatalogRejectsZeroCode(t *testing.T) {
	t.Parallel()

	_, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
	}, []CodeSpec{
		WarningCodeSpec(0, "parse", "missing code"),
	})
	if !errors.Is(err, ErrInvalidCodeCatalogConfig) {
		t.Fatalf("NewCodeCatalog() error=%v, want ErrInvalidCodeCatalogConfig", err)
	}
}

func TestNewCodeCatalogRejectsDuplicateCode(t *testing.T) {
	t.Parallel()

	_, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{
		WarningCodeSpec(2001, "parse", "first"),
		WarningCodeSpec(2001, "parse", "second"),
	})
	if !errors.Is(err, ErrInvalidCodeCatalogConfig) {
		t.Fatalf("NewCodeCatalog() error=%v, want ErrInvalidCodeCatalogConfig", err)
	}
}

func TestCodeCatalogByCodeAndCopy(t *testing.T) {
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{
		ErrorCodeSpec(2001, "parse", "unexpected token"),
	})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	spec, ok := catalog.ByCode(2001)
	if !ok {
		t.Fatal("ByCode(2001) not found")
	}

	if spec.Code != 2001 {
		t.Fatalf("ByCode().Code=%d, want 2001", spec.Code)
	}

	copySpecs := catalog.CodeSpecs()
	if len(copySpecs) != 1 {
		t.Fatalf("CodeSpecs() len=%d, want 1", len(copySpecs))
	}

	copySpecs[0].Message = "changed"
	spec, ok = catalog.ByCode(2001)
	if !ok {
		t.Fatal("ByCode(2001) not found after copy update")
	}

	if spec.Message != "unexpected token" {
		t.Fatalf("ByCode().Message=%q, want unchanged", spec.Message)
	}
}

func TestCodeCatalogWithPrefixExportsPrefixedCodes(t *testing.T) {
	t.Parallel()

	catalog, err := NewCodeCatalog(
		CodeCatalogConfig{
			Module:     "rvcfg",
			CodePrefix: "RVCFG",
			ScopeDescriptions: map[Stage]string{
				"parse": "Parser diagnostics.",
			},
		},
		[]CodeSpec{ErrorCodeSpec(2001, "parse", "unexpected token")},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	spec := catalog.RuleSpec(ErrorCodeSpec(2001, "parse", "unexpected token"))
	if spec.Code != "RVCFG2001" {
		t.Fatalf("RuleSpec().Code=%q, want RVCFG2001", spec.Code)
	}

	if _, ok := catalog.ByCode(2001); !ok {
		t.Fatal("ByCode(2001) not found")
	}
}

func TestCodeCatalogModuleSpec(t *testing.T) {
	t.Parallel()

	catalog, err := NewCodeCatalog(
		CodeCatalogConfig{
			Module:            "module_alpha",
			CodePrefix:        "ALPHA",
			ModuleName:        "Module Alpha",
			ModuleDescription: "Rules for module_alpha.",
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	module := catalog.ModuleSpec()
	if module.ID != "module_alpha" {
		t.Fatalf("ModuleSpec().ID=%q, want module_alpha", module.ID)
	}

	if module.Name != "Module Alpha" {
		t.Fatalf("ModuleSpec().Name=%q, want Module Alpha", module.Name)
	}

	if module.Description != "Rules for module_alpha." {
		t.Fatalf(
			"ModuleSpec().Description=%q, want Rules for module_alpha.",
			module.Description,
		)
	}
}

func TestCodeCatalogRuleID(t *testing.T) {
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{
		ErrorCodeSpec(2001, "parse", "unexpected token"),
	})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	got, err := catalog.RuleID(2001)
	if err != nil {
		t.Fatalf("RuleID(2001) error: %v", err)
	}

	if got != "alpha.parse.unexpected-token" {
		t.Fatalf("RuleID(2001)=%q, want alpha.parse.unexpected-token", got)
	}

	_, err = catalog.RuleID(9999)
	if !errors.Is(err, ErrUnknownCodeCatalogCode) {
		t.Fatalf("RuleID(9999) error=%v, want ErrUnknownCodeCatalogCode", err)
	}
}

func TestCodeCatalogRuleSpec(t *testing.T) {
	spec := WarningCodeSpec(2020, "parse", "rule 2020")
	spec.Rule = &CodeRuleOverride{
		FileKinds:  []FileKind{"alpha.cfg", "beta.cfg"},
		Deprecated: true,
	}
	enabled := false
	spec.Enabled = &enabled
	spec.Options = map[string]any{"mode": "strict"}
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{spec})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	ruleSpec := catalog.RuleSpec(spec)
	if ruleSpec.ID != "alpha.parse.rule-2020" {
		t.Fatalf("RuleSpec().ID=%q, want alpha.parse.rule-2020", ruleSpec.ID)
	}

	if ruleSpec.Scope != "parse" {
		t.Fatalf("RuleSpec().Scope=%q, want parse", ruleSpec.Scope)
	}

	if ruleSpec.DefaultEnabled == nil || *ruleSpec.DefaultEnabled {
		t.Fatal("RuleSpec().DefaultEnabled mismatch, want false")
	}

	if ruleSpec.Code != "ALPHA2020" {
		t.Fatalf("RuleSpec().Code=%q, want ALPHA2020", ruleSpec.Code)
	}

	if len(ruleSpec.FileKinds) != 2 {
		t.Fatalf("RuleSpec().FileKinds len=%d, want 2", len(ruleSpec.FileKinds))
	}

	if !ruleSpec.Deprecated {
		t.Fatal("RuleSpec().Deprecated=false, want true")
	}

	options, ok := ruleSpec.DefaultOptions.(map[string]any)
	if !ok {
		t.Fatalf("RuleSpec().DefaultOptions type=%T, want map[string]any", ruleSpec.DefaultOptions)
	}

	if options["mode"] != "strict" {
		t.Fatalf("RuleSpec().DefaultOptions[mode]=%v, want strict", options["mode"])
	}
}

func TestCodeCatalogRuleSpecOverride(t *testing.T) {
	t.Parallel()

	spec := WarningCodeSpec(2021, "parse", "base title")
	spec.Description = "override description"
	spec.Rule = &CodeRuleOverride{}
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{spec})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	ruleSpec := catalog.RuleSpec(spec)
	if ruleSpec.Description != "override description" {
		t.Fatalf(
			"RuleSpec().Description=%q, want override description",
			ruleSpec.Description,
		)
	}
}

func TestCodeCatalogRuleSpecs(t *testing.T) {
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{
		ErrorCodeSpec(2001, "parse", "unexpected token"),
		WarningCodeSpec(2020, "parse", "autofix inserted semicolon"),
	})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	specs := catalog.RuleSpecs()
	if len(specs) != 2 {
		t.Fatalf("RuleSpecs() len=%d, want 2", len(specs))
	}

	if specs[0].ID != "alpha.parse.unexpected-token" {
		t.Fatalf("RuleSpecs()[0].ID=%q", specs[0].ID)
	}
	if specs[1].ID != "alpha.parse.autofix-inserted-semicolon" {
		t.Fatalf("RuleSpecs()[1].ID=%q", specs[1].ID)
	}
}

func TestNewCodeCatalogProvider(t *testing.T) {
	catalog, err := NewCodeCatalog(CodeCatalogConfig{
		Module:     "alpha",
		CodePrefix: "ALPHA",
		ScopeDescriptions: map[Stage]string{
			"parse": "Parser diagnostics.",
		},
	}, []CodeSpec{
		ErrorCodeSpec(2001, "parse", "unexpected token"),
	})
	if err != nil {
		t.Fatalf("NewCodeCatalog() error: %v", err)
	}

	provider, err := NewCodeCatalogProvider(
		"test.catalog.by_code",
		catalog,
		func(item catalogProviderTestDiagnostic) Diagnostic {
			code, _ := ParseCode(item.Code)
			ruleID, err := catalog.RuleID(code)
			if err != nil {
				ruleID = ""
			}

			return Diagnostic{
				RuleID:   ruleID,
				Severity: SeverityError,
				Message:  item.Message,
			}
		},
	)
	if err != nil {
		t.Fatalf("NewCodeCatalogProvider() error: %v", err)
	}

	var registrar catalogProviderTestRegistrar
	if err := provider.RegisterRules(&registrar); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}
	if len(registrar.runners) != 1 {
		t.Fatalf("registered runners=%d, want 1", len(registrar.runners))
	}

	if module := provider.ModuleSpec(); module.ID != "alpha" {
		t.Fatalf("provider.ModuleSpec().ID=%q, want alpha", module.ID)
	}

	run := RunContext{TargetPath: "test.cfg", TargetKind: "source.cfg"}
	ok := AttachCatalogDiagnostics(
		&run,
		"test.catalog.by_code",
		[]catalogProviderTestDiagnostic{{Code: "2001", Message: "boom"}},
		func(item catalogProviderTestDiagnostic) Code {
			code, _ := ParseCode(item.Code)
			return code
		},
	)
	if !ok {
		t.Fatal("AttachCatalogDiagnostics() returned false")
	}

	var diagnostics []Diagnostic
	err = registrar.runners[0].Check(
		context.Background(),
		&run,
		func(item Diagnostic) {
			diagnostics = append(diagnostics, item)
		},
	)
	if err != nil {
		t.Fatalf("runner.Check() error: %v", err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("len(diagnostics)=%d, want 1", len(diagnostics))
	}
	if diagnostics[0].RuleID != "alpha.parse.unexpected-token" {
		t.Fatalf("diagnostics[0].RuleID=%q", diagnostics[0].RuleID)
	}
}

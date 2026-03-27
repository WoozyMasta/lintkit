// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"errors"
	"testing"
)

func TestWithRuleCodePrefixRejectsInvalidPrefix(t *testing.T) {
	t.Parallel()

	_, err := WithRuleCodePrefix(RuleProviderFunc(func(registrar RuleRegistrar) error {
		return nil
	}), "3D")
	if !errors.Is(err, ErrInvalidCodePrefix) {
		t.Fatalf("WithRuleCodePrefix() error=%v, want ErrInvalidCodePrefix", err)
	}
}

func TestWithRuleCodeOverrides(t *testing.T) {
	t.Parallel()

	runner := stubRuleRunner{
		spec: RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Code:            "ALPHA001",
			Message:         "rule",
			DefaultSeverity: SeverityWarning,
		},
	}

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(runner)
	})
	overridden := WithRuleCodeOverrides(provider, map[string]string{
		"module_alpha.parse.rule": "OVR1001",
	})

	capture := captureRegistrar{}
	if err := overridden.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if len(capture.registered) != 1 {
		t.Fatalf("len(registered)=%d, want 1", len(capture.registered))
	}

	if got := capture.registered[0].RuleSpec().Code; got != "OVR1001" {
		t.Fatalf("RuleSpec().Code=%q, want OVR1001", got)
	}
}

func TestWithRuleCodeOverridesRejectsUnknownRuleID(t *testing.T) {
	t.Parallel()

	runner := stubRuleRunner{
		spec: RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Code:            "ALPHA001",
			Message:         "rule",
			DefaultSeverity: SeverityWarning,
		},
	}

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(runner)
	})
	overridden := WithRuleCodeOverrides(provider, map[string]string{
		"module_beta.parse.rule": "OVR1001",
	})

	capture := captureRegistrar{}
	err := overridden.RegisterRules(&capture)
	if !errors.Is(err, ErrUnknownOverrideRuleID) {
		t.Fatalf("RegisterRules() error=%v, want ErrUnknownOverrideRuleID", err)
	}
}

func TestWithRuleCodeOverridesSoftIgnoresUnknownRuleID(t *testing.T) {
	t.Parallel()

	runner := stubRuleRunner{
		spec: RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Code:            "ALPHA001",
			Message:         "rule",
			DefaultSeverity: SeverityWarning,
		},
	}

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(runner)
	})
	overridden := WithRuleCodeOverridesSoft(provider, map[string]string{
		"module_beta.parse.rule": "OVR1001",
	})

	capture := captureRegistrar{}
	if err := overridden.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if got := capture.registered[0].RuleSpec().Code; got != "ALPHA001" {
		t.Fatalf("RuleSpec().Code=%q, want ALPHA001", got)
	}
}

func TestWithRuleCodePrefix(t *testing.T) {
	t.Parallel()

	runner := stubRuleRunner{
		spec: RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Code:            "ALPHA001",
			Message:         "rule",
			DefaultSeverity: SeverityWarning,
		},
	}

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(runner)
	})
	prefixed, err := WithRuleCodePrefix(provider, "MYAPP")
	if err != nil {
		t.Fatalf("WithRuleCodePrefix() error: %v", err)
	}

	capture := captureRegistrar{}
	if err := prefixed.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if got := capture.registered[0].RuleSpec().Code; got != "MYAPP001" {
		t.Fatalf("RuleSpec().Code=%q, want MYAPP001", got)
	}
}

func TestWithRuleCodePrefixRebasesExistingPrefix(t *testing.T) {
	t.Parallel()

	runner := stubRuleRunner{
		spec: RuleSpec{
			ID:              "module_alpha.parse.rule",
			Module:          "module_alpha",
			Code:            "RVCFG2001",
			Message:         "rule",
			DefaultSeverity: SeverityWarning,
		},
	}

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		return registrar.Register(runner)
	})
	prefixed, err := WithRuleCodePrefix(provider, "MYAPP")
	if err != nil {
		t.Fatalf("WithRuleCodePrefix() error: %v", err)
	}

	capture := captureRegistrar{}
	if err := prefixed.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if got := capture.registered[0].RuleSpec().Code; got != "MYAPP2001" {
		t.Fatalf("RuleSpec().Code=%q, want MYAPP2001", got)
	}
}

type captureRegistrar struct {
	registered []RuleRunner
	modules    []ModuleSpec
}

func (registrar *captureRegistrar) Register(runners ...RuleRunner) error {
	registrar.registered = append(registrar.registered, runners...)
	return nil
}

func (registrar *captureRegistrar) RegisterModule(spec ModuleSpec) error {
	registrar.modules = append(registrar.modules, spec)
	return nil
}

type stubRuleRunner struct {
	spec RuleSpec
}

func (runner stubRuleRunner) RuleSpec() RuleSpec {
	return runner.spec
}

func (runner stubRuleRunner) Check(
	_ context.Context,
	_ *RunContext,
	_ DiagnosticEmit,
) error {
	return nil
}

func TestWithRuleCodeOverridesPreservesModuleRegistration(t *testing.T) {
	t.Parallel()

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		moduleRegistrar, ok := registrar.(ModuleRegistrar)
		if !ok {
			t.Fatal("registrar does not implement ModuleRegistrar")
		}

		if err := moduleRegistrar.RegisterModule(ModuleSpec{
			ID:   "module_alpha",
			Name: "Module Alpha",
		}); err != nil {
			return err
		}

		return registrar.Register(stubRuleRunner{
			spec: RuleSpec{
				ID:              "module_alpha.parse.rule",
				Module:          "module_alpha",
				Code:            "ALPHA001",
				Message:         "rule",
				DefaultSeverity: SeverityWarning,
			},
		})
	})

	wrapped := WithRuleCodeOverrides(provider, map[string]string{
		"module_alpha.parse.rule": "ALPHA2001",
	})

	capture := captureRegistrar{}
	if err := wrapped.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if len(capture.modules) != 1 || capture.modules[0].ID != "module_alpha" {
		t.Fatalf("modules=%v, want one module_alpha registration", capture.modules)
	}
}

func TestWithRuleCodePrefixPreservesModuleRegistration(t *testing.T) {
	t.Parallel()

	provider := RuleProviderFunc(func(registrar RuleRegistrar) error {
		moduleRegistrar, ok := registrar.(ModuleRegistrar)
		if !ok {
			t.Fatal("registrar does not implement ModuleRegistrar")
		}

		if err := moduleRegistrar.RegisterModule(ModuleSpec{
			ID:   "module_alpha",
			Name: "Module Alpha",
		}); err != nil {
			return err
		}

		return registrar.Register(stubRuleRunner{
			spec: RuleSpec{
				ID:              "module_alpha.parse.rule",
				Module:          "module_alpha",
				Code:            "ALPHA001",
				Message:         "rule",
				DefaultSeverity: SeverityWarning,
			},
		})
	})

	wrapped, err := WithRuleCodePrefix(provider, "APP")
	if err != nil {
		t.Fatalf("WithRuleCodePrefix() error: %v", err)
	}

	capture := captureRegistrar{}
	if err := wrapped.RegisterRules(&capture); err != nil {
		t.Fatalf("RegisterRules() error: %v", err)
	}

	if len(capture.modules) != 1 || capture.modules[0].ID != "module_alpha" {
		t.Fatalf("modules=%v, want one module_alpha registration", capture.modules)
	}
}

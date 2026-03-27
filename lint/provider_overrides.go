// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import (
	"context"
	"fmt"
	"strings"
)

// WithRuleCodeOverrides applies rule code overrides to one provider.
//
// Override keys are exact rule IDs, values are target codes.
// Empty rule IDs are ignored. Empty override code clears rule code.
// Unknown override rule IDs return an error during provider registration.
func WithRuleCodeOverrides(
	provider RuleProvider,
	overrides map[string]string,
) RuleProvider {
	return withRuleCodeOverrides(provider, overrides, true)
}

// WithRuleCodeOverridesSoft applies rule code overrides in relaxed mode.
//
// Unknown override rule IDs are ignored.
func WithRuleCodeOverridesSoft(
	provider RuleProvider,
	overrides map[string]string,
) RuleProvider {
	return withRuleCodeOverrides(provider, overrides, false)
}

// withRuleCodeOverrides applies rule code overrides to one provider.
func withRuleCodeOverrides(
	provider RuleProvider,
	overrides map[string]string,
	strict bool,
) RuleProvider {
	if provider == nil || len(overrides) == 0 {
		return provider
	}

	normalized := make(map[string]string, len(overrides))
	for ruleID, code := range overrides {
		id := strings.TrimSpace(ruleID)
		if id == "" {
			continue
		}

		normalized[id] = normalizeRuleCodeToken(code)
	}

	if len(normalized) == 0 {
		return provider
	}

	return RuleProviderFunc(func(registrar RuleRegistrar) error {
		if registrar == nil {
			return ErrNilRuleRegistrar
		}

		return provider.RegisterRules(codeOverrideRegistrar{
			base:      registrar,
			overrides: normalized,
			strict:    strict,
		})
	})
}

// WithRuleCodePrefix rewrites all non-empty provider rule codes to one prefix.
//
// Examples:
//   - "2001" + "RVCFG" => "RVCFG2001"
//   - "RVCFG2001" + "MYAPP" => "MYAPP2001"
func WithRuleCodePrefix(provider RuleProvider, prefix string) (RuleProvider, error) {
	if provider == nil {
		return provider, nil
	}

	normalizedPrefix := normalizeCodePrefix(prefix)
	if normalizedPrefix == "" {
		return provider, nil
	}
	if err := ValidateCodePrefix(normalizedPrefix); err != nil {
		return nil, err
	}

	return RuleProviderFunc(func(registrar RuleRegistrar) error {
		if registrar == nil {
			return ErrNilRuleRegistrar
		}

		return provider.RegisterRules(codePrefixRegistrar{
			base:   registrar,
			prefix: normalizedPrefix,
		})
	}), nil
}

// codeOverrideRegistrar intercepts provider registration and rewrites rule specs.
type codeOverrideRegistrar struct {
	// base stores downstream registrar receiving rewritten runners.
	base RuleRegistrar

	// overrides stores ruleID->code rewrite map.
	overrides map[string]string

	// strict enables unknown override rule ID validation.
	strict bool
}

// codePrefixRegistrar intercepts provider registration and rewrites code prefix.
type codePrefixRegistrar struct {
	// base stores downstream registrar receiving rewritten runners.
	base RuleRegistrar

	// prefix stores target exported code prefix.
	prefix string
}

// RegisterModule delegates module metadata registration when supported.
func (registrar codeOverrideRegistrar) RegisterModule(spec ModuleSpec) error {
	moduleRegistrar, ok := registrar.base.(ModuleRegistrar)
	if !ok {
		return nil
	}

	return moduleRegistrar.RegisterModule(spec)
}

// Register rewrites rule codes and delegates to base registrar.
func (registrar codeOverrideRegistrar) Register(runners ...RuleRunner) error {
	rewritten := make([]RuleRunner, 0, len(runners))
	applied := make(map[string]struct{}, len(registrar.overrides))
	for index := range runners {
		if runners[index] == nil {
			rewritten = append(rewritten, nil)
			continue
		}

		spec := runners[index].RuleSpec()
		if code, exists := registrar.overrides[strings.TrimSpace(spec.ID)]; exists {
			spec.Code = normalizeRuleCodeToken(code)
			applied[strings.TrimSpace(spec.ID)] = struct{}{}
		}

		rewritten = append(rewritten, ruleSpecOverrideRunner{
			base: runners[index],
			spec: spec,
		})
	}

	if registrar.strict {
		for overrideRuleID := range registrar.overrides {
			if _, ok := applied[overrideRuleID]; ok {
				continue
			}

			return fmt.Errorf("%w: %q", ErrUnknownOverrideRuleID, overrideRuleID)
		}
	}

	return registrar.base.Register(rewritten...)
}

// Register rewrites runner codes to one target prefix and delegates to base registrar.
func (registrar codePrefixRegistrar) Register(runners ...RuleRunner) error {
	rewritten := make([]RuleRunner, 0, len(runners))
	for index := range runners {
		if runners[index] == nil {
			rewritten = append(rewritten, nil)
			continue
		}

		spec := runners[index].RuleSpec()
		if spec.Code != "" {
			spec.Code = RebaseCodePrefix(registrar.prefix, spec.Code)
		}

		rewritten = append(rewritten, ruleSpecOverrideRunner{
			base: runners[index],
			spec: spec,
		})
	}

	return registrar.base.Register(rewritten...)
}

// RegisterModule delegates module metadata registration when supported.
func (registrar codePrefixRegistrar) RegisterModule(spec ModuleSpec) error {
	moduleRegistrar, ok := registrar.base.(ModuleRegistrar)
	if !ok {
		return nil
	}

	return moduleRegistrar.RegisterModule(spec)
}

// ruleSpecOverrideRunner delegates checks and returns rewritten static spec.
type ruleSpecOverrideRunner struct {
	// base stores original runner.
	base RuleRunner

	// spec stores rewritten stable rule spec.
	spec RuleSpec
}

// RuleSpec returns rewritten metadata.
func (runner ruleSpecOverrideRunner) RuleSpec() RuleSpec {
	return runner.spec
}

// Check delegates to original runner implementation.
func (runner ruleSpecOverrideRunner) Check(
	ctx context.Context,
	run *RunContext,
	emit DiagnosticEmit,
) error {
	return runner.base.Check(ctx, run, emit)
}

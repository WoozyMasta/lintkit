// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import "github.com/woozymasta/lintkit/lint"

// NewEngineWithProviders creates engine and applies providers in order.
func NewEngineWithProviders(providers ...lint.RuleProvider) (*Engine, error) {
	engine := NewEngine()
	if err := RegisterProviders(engine, providers...); err != nil {
		return nil, err
	}

	return engine, nil
}

// RegisterProviders applies providers in declaration order.
func RegisterProviders(engine *Engine, providers ...lint.RuleProvider) error {
	if engine == nil {
		return ErrNilEngine
	}

	return lint.RegisterRuleProviders(engine, providers...)
}

// RegisterProviders applies providers in declaration order.
func (engine *Engine) RegisterProviders(providers ...lint.RuleProvider) error {
	return RegisterProviders(engine, providers...)
}

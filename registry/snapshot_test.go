// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

func TestSnapshotFromProviders(t *testing.T) {
	t.Parallel()

	snapshot, err := SnapshotFromProviders(testProvider{})
	if err != nil {
		t.Fatalf("SnapshotFromProviders() error: %v", err)
	}

	if len(snapshot.Rules) != 1 {
		t.Fatalf("len(snapshot.Rules)=%d, want 1", len(snapshot.Rules))
	}

	if snapshot.Rules[0].ID != "test.module.rule" {
		t.Fatalf("snapshot.Rules[0].ID=%q", snapshot.Rules[0].ID)
	}
}

func TestRegisterProvidersNilEngine(t *testing.T) {
	t.Parallel()

	err := RegisterProviders(nil, testProvider{})
	if !errors.Is(err, ErrNilEngine) {
		t.Fatalf("RegisterProviders(nil) error=%v, want ErrNilEngine", err)
	}
}

func TestRegisterProvidersNilProvider(t *testing.T) {
	t.Parallel()

	engine := linting.NewEngine()
	err := RegisterProviders(engine, nil)
	if !errors.Is(err, ErrNilProvider) {
		t.Fatalf("RegisterProviders(nil provider) error=%v, want ErrNilProvider", err)
	}
}

type testProvider struct{}

func (testProvider) RegisterRules(registrar lint.RuleRegistrar) error {
	return registrar.Register(testRunner{})
}

type testRunner struct{}

func (testRunner) RuleSpec() lint.RuleSpec {
	return lint.RuleSpec{
		ID:              "test.module.rule",
		Module:          "test",
		Code:            "TEST1001",
		Message:         "test rule",
		DefaultSeverity: lint.SeverityWarning,
	}
}

func (testRunner) Check(_ context.Context, _ *lint.RunContext, _ lint.DiagnosticEmit) error {
	return nil
}

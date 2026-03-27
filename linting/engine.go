// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/woozymasta/lintkit/lint"
)

// RunOptions controls runtime rule filtering and error behavior.
type RunOptions struct {
	// EnableServiceDiagnostics controls built-in lintkit policy diagnostics.
	// Nil means enabled by default.
	EnableServiceDiagnostics *bool `json:"enable_service_diagnostics,omitempty" yaml:"enable_service_diagnostics,omitempty" jsonschema:"default=true,example=true"`

	// CompiledPolicy stores precompiled rule policy for repeated runs.
	CompiledPolicy *CompiledRunPolicy `json:"-" yaml:"-" jsonschema:"-"`

	// Policy stores global and path-scoped rule settings.
	Policy *RunPolicy `json:"-" yaml:"-" jsonschema:"-"`

	// Suppressions stores optional inline or external suppression index.
	// It is applied centrally in emit pipeline after diagnostic normalization.
	Suppressions lint.SuppressionSet `json:"-" yaml:"-" jsonschema:"-"`

	// RuleIDs filters active rules by IDs.
	// Empty means "run all registered rules".
	RuleIDs []string `json:"rule_ids,omitempty" yaml:"rule_ids,omitempty" jsonschema:"example=module_alpha.parse.unexpected-token,module_beta.validate.asset-exists"`

	// StopOnRuleError stops execution after the first rule runtime error.
	StopOnRuleError bool `json:"stop_on_rule_error,omitempty" yaml:"stop_on_rule_error,omitempty" jsonschema:"default=false"`
}

// RuleError stores one runtime rule execution error.
type RuleError struct {
	// Cause is runtime execution error.
	Cause error `json:"-" yaml:"-" jsonschema:"-"`

	// RuleID is source rule ID for this runtime error.
	RuleID string `json:"rule_id,omitempty" yaml:"rule_id,omitempty" jsonschema:"example=module_alpha.parse.unexpected-token"`
}

// Error returns compact rule runtime error text.
func (item RuleError) Error() string {
	if item.RuleID == "" {
		if item.Cause == nil {
			return ""
		}

		return item.Cause.Error()
	}

	if item.Cause == nil {
		return item.RuleID
	}

	return item.RuleID + ": " + item.Cause.Error()
}

// RunResult stores diagnostics and runtime errors from one engine run.
type RunResult struct {
	// Diagnostics are normalized findings from executed rules.
	Diagnostics []lint.Diagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`

	// Suppressed stores suppressed diagnostic audit events.
	Suppressed []lint.SuppressionHit `json:"suppressed,omitempty" yaml:"suppressed,omitempty"`

	// RuleErrors stores non-fatal per-rule runtime failures.
	RuleErrors []RuleError `json:"rule_errors,omitempty" yaml:"rule_errors,omitempty"`
}

// Engine stores rule runners and shared registry metadata.
type Engine struct {
	// runners stores registered runners by stable rule ID.
	runners map[string]lint.RuleRunner

	// specs stores normalized rule metadata by stable rule ID.
	specs map[string]*lint.RuleSpec

	// registry stores normalized rule metadata.
	registry *Registry

	// orderedRuntimeRules stores deterministic runtime execution plan.
	orderedRuntimeRules []runtimeRule

	// mu guards concurrent reads and writes.
	mu sync.RWMutex
}

// runtimeRule stores one resolved runtime rule pair.
type runtimeRule struct {
	// Runner stores registered rule runner instance.
	Runner lint.RuleRunner

	// Spec stores normalized rule metadata pointer.
	Spec *lint.RuleSpec
}

// policyResolver resolves path enablement and rule settings for one run.
type policyResolver interface {
	PathEnabled(path string, isDir bool) bool
	Resolve(rule lint.RuleSpec, path string, isDir bool) RuleDecision
}

// NewEngine creates an empty lint rule engine.
func NewEngine() *Engine {
	return &Engine{
		runners:  make(map[string]lint.RuleRunner),
		specs:    make(map[string]*lint.RuleSpec),
		registry: NewRegistry(),
	}
}

// Register validates and registers one or more rule runners.
func (engine *Engine) Register(runners ...lint.RuleRunner) error {
	if engine == nil {
		return ErrNilEngine
	}

	if len(runners) == 0 {
		return nil
	}

	type runnerPair struct {
		Runner lint.RuleRunner
		Spec   lint.RuleSpec
	}

	normalized := make([]runnerPair, 0, len(runners))
	seen := make(map[string]struct{}, len(runners))
	for index := range runners {
		if runners[index] == nil {
			return fmt.Errorf("%w at index %d", ErrNilRuleRunner, index)
		}

		spec, err := normalizeRuleSpec(runners[index].RuleSpec())
		if err != nil {
			return fmt.Errorf("runner[%d]: %w", index, err)
		}

		if _, exists := seen[spec.ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRuleID, spec.ID)
		}

		seen[spec.ID] = struct{}{}
		normalized = append(normalized, runnerPair{
			Runner: runners[index],
			Spec:   spec,
		})
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	for index := range normalized {
		if _, exists := engine.runners[normalized[index].Spec.ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRuleID, normalized[index].Spec.ID)
		}
	}

	specs := make([]lint.RuleSpec, 0, len(normalized))
	for index := range normalized {
		specs = append(specs, normalized[index].Spec)
	}

	if err := engine.registry.RegisterMany(specs...); err != nil {
		return err
	}

	for index := range normalized {
		engine.runners[normalized[index].Spec.ID] = normalized[index].Runner
		spec := normalized[index].Spec
		engine.specs[spec.ID] = &spec
	}

	engine.refreshOrderedRuleIDsLocked()

	return nil
}

// RegisterModule validates and registers one lint module metadata descriptor.
func (engine *Engine) RegisterModule(spec lint.ModuleSpec) error {
	if engine == nil {
		return ErrNilEngine
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	return engine.registry.RegisterModule(spec)
}

// Rule returns one registered rule metadata by ID.
func (engine *Engine) Rule(id string) (lint.RuleSpec, bool) {
	if engine == nil {
		return lint.RuleSpec{}, false
	}

	key := strings.TrimSpace(id)
	if key == "" {
		return lint.RuleSpec{}, false
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()

	spec, ok := engine.specs[key]
	if !ok || spec == nil {
		return lint.RuleSpec{}, false
	}

	return cloneRuleSpec(*spec), true
}

// Rules returns deterministic sorted list of all registered rule specs.
func (engine *Engine) Rules() []lint.RuleSpec {
	if engine == nil {
		return nil
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()

	return engine.registry.Rules()
}

// Snapshot returns deterministic export snapshot.
func (engine *Engine) Snapshot() lint.RegistrySnapshot {
	if engine == nil {
		return lint.RegistrySnapshot{}
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()

	return engine.registry.Snapshot()
}

// ExportJSON returns engine registry snapshot encoded as JSON.
func (engine *Engine) ExportJSON(pretty bool) ([]byte, error) {
	if engine == nil {
		return nil, ErrNilEngine
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()

	return engine.registry.ExportJSON(pretty)
}

// Run executes selected rules against one runtime context.
func (engine *Engine) Run(
	ctx context.Context,
	runContext lint.RunContext,
	options *RunOptions,
) (RunResult, error) {
	if engine == nil {
		return RunResult{}, ErrNilEngine
	}

	if ctx == nil {
		ctx = context.Background()
	}

	selectedRules, err := engine.selectRuntimeRules(options)
	if err != nil {
		return RunResult{}, err
	}

	resolver, err := engine.resolvePolicy(options)
	if err != nil {
		return RunResult{}, err
	}

	if resolver != nil {
		if !resolver.PathEnabled(runContext.TargetPath, runContext.TargetIsDir) {
			return RunResult{
				Diagnostics: make([]lint.Diagnostic, 0),
				RuleErrors:  make([]RuleError, 0),
			}, nil
		}
	}

	if runContext.Values == nil {
		runContext.Values = make(map[string]any)
	}

	suppressions := lookupSuppressions(options)
	result := RunResult{
		Diagnostics: make([]lint.Diagnostic, 0, suggestedDiagnosticCapacity(len(selectedRules))),
		Suppressed:  make([]lint.SuppressionHit, 0, suggestedSuppressedCapacity(suppressions, len(selectedRules))),
		RuleErrors:  make([]RuleError, 0, 2),
	}
	var policyDiagnostics []lint.Diagnostic
	if options != nil && serviceDiagnosticsEnabled(options) {
		if options.CompiledPolicy != nil {
			policyDiagnostics = cloneDiagnostics(options.CompiledPolicy.ServiceDiagnostics)
		} else if options.Policy != nil {
			policyDiagnostics = collectPolicyServiceDiagnostics(
				options.Policy,
				engine.registry.rulesView(),
			)
		}
	}

	appendUniquePolicyDiagnostics(&runContext, &result.Diagnostics, policyDiagnostics)

	emitter := newRunEmitter(
		&result.Diagnostics,
		&result.Suppressed,
		&runContext,
		suppressions,
	)
	emit := emitter.EmitNoSuppressions
	if suppressions != nil {
		emit = emitter.Emit
	}
	targetKind := lint.NormalizeFileKind(runContext.TargetKind)
	ruleOptionsSet := false

	for index := range selectedRules {
		item := selectedRules[index]
		ruleID := ""
		if item.Spec != nil {
			ruleID = item.Spec.ID
		}

		if item.Runner == nil || item.Spec == nil {
			// Defensive branch for impossible state: registry and runners should
			// stay in sync after successful Register calls.
			result.RuleErrors = append(result.RuleErrors, RuleError{
				RuleID: ruleID,
				Cause:  fmt.Errorf("%w: %q", ErrUnknownRuleID, ruleID),
			})
			continue
		}

		if !lint.SupportsNormalizedFileKind(item.Spec.FileKinds, targetKind) {
			continue
		}

		effectiveSeverity := item.Spec.DefaultSeverity
		var effectiveOptions any
		if item.Spec.DefaultEnabled != nil && !*item.Spec.DefaultEnabled && resolver == nil {
			continue
		}

		if resolver != nil {
			decision := resolver.Resolve(
				*item.Spec,
				runContext.TargetPath,
				runContext.TargetIsDir,
			)
			if !decision.Enabled {
				continue
			}

			effectiveSeverity = decision.Severity
			effectiveOptions = decision.Options
		}

		if effectiveOptions != nil {
			lint.SetCurrentRuleOptions(
				&runContext,
				cloneDynamicValue(effectiveOptions),
			)
			ruleOptionsSet = true
		} else if ruleOptionsSet {
			lint.ClearCurrentRuleOptions(&runContext)
			ruleOptionsSet = false
		}

		emitter.RuleID = item.Spec.ID
		emitter.DefaultSeverity = effectiveSeverity
		runErr := runRuleCheckSafely(ctx, item.Runner, &runContext, emit)
		if runErr == nil {
			continue
		}

		result.RuleErrors = append(result.RuleErrors, RuleError{
			RuleID: item.Spec.ID,
			Cause:  runErr,
		})

		if options != nil && options.StopOnRuleError {
			break
		}
	}

	if ruleOptionsSet {
		lint.ClearCurrentRuleOptions(&runContext)
	}

	return result, nil
}

// runRuleCheckSafely executes runner check and converts panic to runtime error.
func runRuleCheckSafely(
	ctx context.Context,
	runner lint.RuleRunner,
	runContext *lint.RunContext,
	emit lint.DiagnosticEmit,
) (runErr error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		recoveredErr, ok := recovered.(error)
		if ok {
			runErr = fmt.Errorf("%w: %w", ErrRuleRunnerPanic, recoveredErr)
			return
		}

		runErr = fmt.Errorf("%w: %v", ErrRuleRunnerPanic, recovered)
	}()

	return runner.Check(ctx, runContext, emit)
}

// selectRuntimeRules resolves deterministic runtime rule execution list.
func (engine *Engine) selectRuntimeRules(options *RunOptions) ([]runtimeRule, error) {
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	if options == nil || len(options.RuleIDs) == 0 {
		return engine.orderedRuntimeRules, nil
	}

	out := make([]runtimeRule, 0, len(options.RuleIDs))
	seen := make(map[string]struct{}, len(options.RuleIDs))
	for index := range options.RuleIDs {
		ruleID := strings.TrimSpace(options.RuleIDs[index])
		if ruleID == "" {
			continue
		}

		if _, exists := seen[ruleID]; exists {
			continue
		}

		runner, runnerExists := engine.runners[ruleID]
		spec, specExists := engine.specs[ruleID]
		if !runnerExists || !specExists {
			return nil, fmt.Errorf("%w: %q", ErrUnknownRuleID, ruleID)
		}

		seen[ruleID] = struct{}{}
		out = append(out, runtimeRule{
			Runner: runner,
			Spec:   spec,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("%w: empty rule selection", ErrUnknownRuleID)
	}

	return out, nil
}

// RunDefault executes rules with background context.
func (engine *Engine) RunDefault(
	runContext lint.RunContext,
	options *RunOptions,
) (RunResult, error) {
	return engine.Run(context.Background(), runContext, options)
}

// RunWithProfile executes one run using canonical run profile settings.
func (engine *Engine) RunWithProfile(
	ctx context.Context,
	runContext lint.RunContext,
	profile RunProfile,
) (RunResult, error) {
	normalizedProfile, err := profile.Normalize()
	if err != nil {
		return RunResult{}, err
	}

	if !normalizedProfile.Enabled {
		return RunResult{
			Diagnostics: make([]lint.Diagnostic, 0),
			RuleErrors:  make([]RuleError, 0),
		}, nil
	}

	options := normalizedProfile.optionsFromNormalized()

	return engine.Run(ctx, runContext, &options)
}

// resolvePolicy returns runtime resolver from options policy fields.
func (engine *Engine) resolvePolicy(options *RunOptions) (policyResolver, error) {
	if options == nil {
		return nil, nil
	}

	if options.Policy != nil && options.CompiledPolicy != nil {
		return nil, fmt.Errorf(
			"%w: set only one of policy or compiled_policy",
			ErrInvalidRunPolicy,
		)
	}

	if options.CompiledPolicy != nil {
		return options.CompiledPolicy, nil
	}

	if options.Policy == nil {
		return nil, nil
	}

	if options.Policy.Strict {
		if err := options.Policy.Validate(engine.Rules()); err != nil {
			return nil, err
		}
	}

	return options.Policy, nil
}

// refreshOrderedRuleIDsLocked rebuilds cached default execution rule order.
//
// Caller must hold engine.mu write lock.
func (engine *Engine) refreshOrderedRuleIDsLocked() {
	rules := engine.registry.Rules()
	orderedRuntime := make([]runtimeRule, 0, len(rules))
	for index := range rules {
		ruleID := rules[index].ID
		orderedRuntime = append(orderedRuntime, runtimeRule{
			Runner: engine.runners[ruleID],
			Spec:   engine.specs[ruleID],
		})
	}

	engine.orderedRuntimeRules = orderedRuntime
}

// runEmitter stores mutable emission state for current running rule.
type runEmitter struct {
	// target stores target diagnostics sink.
	target *[]lint.Diagnostic

	// suppressed stores suppression audit sink.
	suppressed *[]lint.SuppressionHit

	// context stores current file/target run context.
	context *lint.RunContext

	// suppressions stores optional suppression set.
	suppressions lint.SuppressionSet

	// RuleID stores currently running rule ID.
	RuleID string

	// DefaultSeverity stores currently resolved rule severity.
	DefaultSeverity lint.Severity
}

// newRunEmitter creates mutable emission state for one engine run.
func newRunEmitter(
	target *[]lint.Diagnostic,
	suppressed *[]lint.SuppressionHit,
	context *lint.RunContext,
	suppressions lint.SuppressionSet,
) *runEmitter {
	return &runEmitter{
		target:       target,
		suppressed:   suppressed,
		context:      context,
		suppressions: suppressions,
	}
}

// Emit normalizes diagnostic and appends it to run result.
func (emitter *runEmitter) Emit(diagnostic lint.Diagnostic) {
	normalized := normalizeDiagnostic(
		diagnostic,
		emitter.context,
		emitter.RuleID,
		emitter.DefaultSeverity,
	)

	decision := suppressionDecision(emitter.suppressions, normalized)
	if decision.Suppressed {
		if emitter.suppressed != nil {
			*emitter.suppressed = append(
				*emitter.suppressed,
				newSuppressionHit(normalized, decision),
			)
		}

		return
	}

	*emitter.target = append(*emitter.target, normalized)
}

// EmitNoSuppressions appends normalized diagnostic without suppression checks.
func (emitter *runEmitter) EmitNoSuppressions(diagnostic lint.Diagnostic) {
	*emitter.target = append(
		*emitter.target,
		normalizeDiagnostic(
			diagnostic,
			emitter.context,
			emitter.RuleID,
			emitter.DefaultSeverity,
		),
	)
}

// normalizeDiagnostic applies metadata and runtime defaults to one diagnostic.
func normalizeDiagnostic(
	diagnostic lint.Diagnostic,
	context *lint.RunContext,
	ruleID string,
	defaultSeverity lint.Severity,
) lint.Diagnostic {
	out := diagnostic
	if out.RuleID == "" {
		out.RuleID = ruleID
	}

	if out.Severity == "" {
		out.Severity = defaultSeverity
		if out.Severity == "" {
			out.Severity = lint.SeverityWarning
		}
	}

	if out.Path == "" {
		out.Path = context.TargetPath
	}

	if out.Start.File == "" && context.TargetPath != "" {
		out.Start.File = context.TargetPath
	}

	if isZeroPosition(out.End) {
		out.End = out.Start
	}

	return out
}

// isZeroPosition reports whether source position has zero coordinates.
func isZeroPosition(position lint.Position) bool {
	return position.File == "" &&
		position.Line == 0 &&
		position.Column == 0 &&
		position.Offset == 0
}

// suggestedDiagnosticCapacity returns initial diagnostics buffer capacity.
func suggestedDiagnosticCapacity(ruleCount int) int {
	if ruleCount > 16 {
		return ruleCount
	}

	return 16
}

// suggestedSuppressedCapacity returns initial suppressed audit buffer capacity.
func suggestedSuppressedCapacity(
	suppressions lint.SuppressionSet,
	ruleCount int,
) int {
	if suppressions == nil {
		return 0
	}

	if ruleCount > 4 {
		return ruleCount
	}

	return 4
}

// lookupSuppressions returns configured suppression set from run options.
func lookupSuppressions(options *RunOptions) lint.SuppressionSet {
	if options == nil {
		return nil
	}

	return options.Suppressions
}

// suppressionDecision returns suppression decision for one diagnostic.
func suppressionDecision(
	suppressions lint.SuppressionSet,
	diagnostic lint.Diagnostic,
) lint.SuppressionDecision {
	if suppressions == nil {
		return lint.SuppressionDecision{}
	}

	decision := suppressions.DecideSuppression(
		diagnostic.RuleID,
		diagnostic.Path,
		diagnostic.Start,
		diagnostic.End,
	)
	if !decision.Suppressed {
		return decision
	}

	decision.Scope = lint.NormalizeSuppressionScope(decision.Scope)
	return decision
}

// newSuppressionHit builds one suppression audit entry from diagnostic.
func newSuppressionHit(
	diagnostic lint.Diagnostic,
	decision lint.SuppressionDecision,
) lint.SuppressionHit {
	return lint.SuppressionHit{
		RuleID:    diagnostic.RuleID,
		Path:      diagnostic.Path,
		Start:     diagnostic.Start,
		End:       diagnostic.End,
		Scope:     lint.NormalizeSuppressionScope(decision.Scope),
		Reason:    strings.TrimSpace(decision.Reason),
		Source:    strings.TrimSpace(decision.Source),
		ExpiresAt: strings.TrimSpace(decision.ExpiresAt),
	}
}

// JoinErrors joins rule runtime errors into one aggregated error.
func (result RunResult) JoinErrors() error {
	if len(result.RuleErrors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(result.RuleErrors))
	for index := range result.RuleErrors {
		errs = append(errs, result.RuleErrors[index])
	}

	return errors.Join(errs...)
}

// DiagnosticCount returns diagnostic count by severity.
func (result RunResult) DiagnosticCount(severity lint.Severity) int {
	count := 0
	for index := range result.Diagnostics {
		if result.Diagnostics[index].Severity == severity {
			count++
		}
	}

	return count
}

// SortDiagnostics applies deterministic diagnostics order.
func (result *RunResult) SortDiagnostics() {
	if result == nil || len(result.Diagnostics) == 0 {
		return
	}

	slices.SortStableFunc(result.Diagnostics, func(left lint.Diagnostic, right lint.Diagnostic) int {
		if left.Path != right.Path {
			if left.Path < right.Path {
				return -1
			}

			return 1
		}

		if left.Start.Line != right.Start.Line {
			if left.Start.Line < right.Start.Line {
				return -1
			}

			return 1
		}

		if left.Start.Column != right.Start.Column {
			if left.Start.Column < right.Start.Column {
				return -1
			}

			return 1
		}

		if left.RuleID != right.RuleID {
			if left.RuleID < right.RuleID {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Message, right.Message)
	})
}

# linting package

`linting` is the upstream runtime package of `lintkit`.

If your application imports multiple lint-enabled modules
and you want one combined lint run with one policy,
this is the package you use.

## What you do in an application

Minimal flow:

1. create engine;
1. register providers from modules;
1. build policy from config;
1. run engine for target file/object;
1. read diagnostics and runtime rule errors.

## Core runtime API

* `Engine` executes registered rules.
* `Registry` stores validated `RuleSpec` and `ModuleSpec`.
* `RunProfile` is canonical normalized execution profile.
* `RunOptions` controls policy, filtering, and suppression behavior.
* `RunResult` contains diagnostics, suppressed items, and rule errors.

Helper constructor:

* `NewEngineWithProviders(...)` creates engine and registers providers.

## Policy model (config-facing)

`RunPolicyConfig` is the config-friendly model:

* `exclude`: global path exclusions.
* `rules[]`: ordered rule entries (`RunPolicyRuleConfig`).
* `soft_unknown_selectors`: enable soft mode for unknown selectors.
* `fail_on`: minimum severity that fails run result (`error` by default).

Each `rules[]` entry has:

* `rule`: selector token;
* optional `enabled`, `severity`, `options`;
* optional `exclude` for this specific rule entry.

Evaluation order is top-to-bottom.
If multiple entries match, later entries override earlier ones.
`config.ShouldFail(result)` always treats runtime `RuleErrors` as critical.

## Canonical profile flow

For upstream tools, preferred flow is:

* build one profile from config and overlays via `BuildRunProfile(...)`;
* convert profile to low-level runtime wiring via `profile.Options()`;
* execute with `Engine.RunWithProfile(...)`;
* evaluate fail condition via `profile.ShouldFail(result)`.

This keeps merge/compile/fail-threshold logic in one runtime API path.

## Selector forms

Supported selector forms:

* `*`
* `<module>.*`
* `<module>.<scope>.*`
* `<rule-id>`
* `<CODE>`

## Path matching

`linting` is matcher-agnostic by design.

Use `PathMatcher` implementations directly,
or build them from glob patterns via `PathRulesCompiler(...)`.

## Performance and repeated runs

For repeated runs with the same policy, compile once:

* `compiled, _ := policy.Compile(engine.Rules())`
* pass `RunOptions{CompiledPolicy: compiled}`
  to avoid repeated selector compilation.

`BuildRunProfile(...)` already compiles policy once and returns ready profile
for repeated runs.

## Related docs

* [../README.md](../README.md) for project overview.
* [../lint/README.md](../lint/README.md) for downstream contracts.
* [../cmd/lintkit/README.md](../cmd/lintkit/README.md) for snapshot/docs CLI.

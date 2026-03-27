# lint package

`lint` is the downstream API of `lintkit`.

If you write a module that has its own diagnostics,
this is the package you import.
It defines a common contract so any upstream application
can register your rules and run them in one shared runtime.

This package does not execute lint checks.
Runtime execution is in `lintkit/linting`.

## What you do in a module

Minimal flow:

1. describe rules (`RuleSpec`);
1. expose provider (`RuleProvider`) that registers those rules;
1. map your module diagnostics to shared `lint.Diagnostic`.

That is enough for upstream integration.

## Core contracts

* `RuleSpec` is rule metadata.
* `Diagnostic` is one normalized finding.
* `RunContext` carries target path/kind and module runtime values.
* `RuleRunner` is one executable rule callback.
* `RuleProvider`/`RuleRegistrar` are registration interfaces.
* `ComposeProviders(...)` combines nested providers in one declared order.

## Message vs description

`RuleSpec.Message` is the runtime diagnostic text.

`RuleSpec.Description` is documentation text.
It can be multiline and may include markdown-style formatting.

Catalog flow follows the same split:

* `CodeSpec.Message` is required and used for rule message generation.
* `CodeSpec.Description` is optional detailed docs text.
* `CodeSpec.Rule` accepts only `CodeRuleOverride` (narrow metadata override),
  not full `RuleSpec`.

## Optional helper: catalog model

Use `CodeSpec` + `CodeCatalog` if your module already has stable numeric codes
and you want less boilerplate.

What catalog gives you:

* public short code generation (`PREFIX2001`);
* default rule id generation from module/scope/description;
* one place to keep code metadata.
* explicit unknown-code handling (`RuleID(code)` returns error).

`CodeCatalogConfig.ScopeDescriptions`
is required for scopes used by catalog rows,
so scope docs are always available in generated reports.

## Runtime values and suppressions

Use `SetRunValue`/`GetRunValue` to pass precomputed module data
(AST, indexes, etc.) through run context without repeated parsing.
Use namespaced keys in `module.key` form (for example `module_alpha.ast`)
to avoid cross-module collisions.

Inline suppression parsing stays module-specific.
Module code parses comments and passes compiled suppression data
via shared suppression contracts.

## Related docs

* [../README.md](../README.md) for project overview.
* [../linting/README.md](../linting/README.md) for runtime behavior.
* [../cmd/lintkit/README.md](../cmd/lintkit/README.md) for snapshot/docs CLI.

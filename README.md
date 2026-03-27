# lintkit

`lintkit` is a shared integration toolkit for modular lint systems.

It lets downstream libraries define diagnostics once
and expose rules through stable contracts,
while upstream tools run all rules in one policy-driven runtime.

This repository covers three concerns:

* downstream contracts and provider registration;
* upstream runtime, registry, policy, and execution;
* deterministic snapshot and documentation generation for CI.

## Terms

`downstream module` means a library that owns domain parsing and diagnostics.

`upstream application` means a tool that imports many modules,
applies one policy, and produces one combined lint result.

## Repository layout

* [lint](./lint/README.md) contains downstream contracts and catalog helpers.
* [linting](./linting/README.md) contains upstream runtime and policy logic.
* [cmd/lintkit](./cmd/lintkit/README.md) is snapshot/doc/schema CLI tooling.
* [registry](./registry/README.md) is in-process snapshot assembly API.
* [linttest](./linttest/README.md) contains downstream contract test helpers.

## End-to-end example

Two downstream modules define local code catalogs:

```go
// module_a
var catalogA, _ = lint.NewCodeCatalog(lint.CodeCatalogConfig{
  Module:            "modulea",
  CodePrefix:        "MODA",
  ModuleName:        "Module A",
  ModuleDescription: "Rules for module_a parser.",
  ScopeDescriptions: map[lint.Stage]string{
    "parse": "Parser diagnostics.",
  },
}, []lint.CodeSpec{
  lint.WarningCodeSpec(1001, "parse", "empty block"),
})

// module_b
var catalogB, _ = lint.NewCodeCatalog(lint.CodeCatalogConfig{
  Module:            "moduleb",
  CodePrefix:        "MODB",
  ModuleName:        "Module B",
  ModuleDescription: "Rules for module_b binary decoder.",
  ScopeDescriptions: map[lint.Stage]string{
    "decode": "Binary decode diagnostics.",
  },
}, []lint.CodeSpec{
  lint.ErrorCodeSpec(2001, "decode", "invalid header signature"),
})
```

An upstream app registers providers once:

```go
engine, err := linting.NewEngineWithProviders(
  modulea.LintRulesProvider{},
  moduleb.LintRulesProvider{},
)
if err != nil {
  return err
}
```

Then it applies one policy to both modules:

```go
cfg := linting.RunPolicyConfig{
  Exclude: []string{"**/vendor/**"},
  FailOn:  lint.SeverityError,
  Rules: []linting.RunPolicyRuleConfig{
    {
      Rule: "*",
      Severity: lint.SeverityWarning
    },
    {
      Rule: "MODB2001", 
      Enabled: linting.BoolPtr(false)
    },
    {
      Rule:    "modulea.parse.*",
      Exclude: []string{"**/generated/**"},
      Severity: lint.SeverityNotice,
    },
  },
}
```

Canonical runtime path is:

* build effective profile via `linting.BuildRunProfile(...)`;
* execute rules via `engine.RunWithProfile(...)`;
* evaluate exit condition via `profile.ShouldFail(result)`.

`RunPolicyConfig.Rules` is ordered.
Later matching entries override earlier entries.
`cfg.ShouldFail(result)` evaluates final exit condition for this config
and always treats runtime rule errors as critical.

## Path matching

`linting` is matcher-agnostic.
Runtime uses `PathMatcher`,
while `PathRulesCompiler(...)` compiles declarative glob patterns
(`*`, `**`, `?`) into runtime matchers.

## Where to continue

Start with [lint](./lint/README.md),
then move to [linting](./linting/README.md).
For CI snapshot and docs generation use [cmd/lintkit](./cmd/lintkit/README.md).

# registry package

`registry` is the programmatic API for building lint rule snapshots in code.

Use this package when your application already imports provider packages
and you want to collect one deterministic `lint.RegistrySnapshot`
without going through CLI commands.

The package is intentionally small.
It focuses on provider registration and snapshot assembly only.

## Core API

* `SnapshotFromProviders(...)` builds a snapshot from providers.
* `RegisterProviders(...)` registers providers into an existing engine.
* `SnapshotFromEngine(...)` exports snapshot from an existing engine.

## Typical usage

```go
snapshot, err := registry.SnapshotFromProviders(
    modulea.LintRulesProvider{},
    moduleb.LintRulesProvider{},
)
if err != nil {
    return err
}
```

If you need source-root discovery, toolchain-based collection,
or markdown generation, use `cmd/lintkit`.

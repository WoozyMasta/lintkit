# linttest package

`linttest` provides small contract-check helpers for downstream modules.

A module that exposes `DiagnosticCatalog()`, `LintRuleSpecs()`, and
`LintRuleID(code)` can use `linttest.AssertCatalogContract(...)` to verify
that these pieces stay aligned.

The goal is to catch catalog drift early in module tests, before integration
with an upstream runtime.

## Related docs

See [../README.md](../README.md) for the full architecture and
[../lint/README.md](../lint/README.md) for downstream contract details.

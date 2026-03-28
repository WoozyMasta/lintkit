# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.2.1][] - 2026-03-28

### Added

* `lintkit snapshot --temp-dir`
  flagto set directory for generated collector source.
* `lintkit snapshot --keep-collector`
  flag to keep generated collector source after run for diagnostics.
* `lintkit snapshot --include-lintkit-rules`
  flag to include built-in lintkit rules in snapshot output.

### Changed

* `lintkit snapshot` now excludes built-in `lintkit` service rules by default.
* Generated collector source now uses system temporary directory
  by default instead of provider workdir.

[0.2.1]: https://github.com/WoozyMasta/lintkit/compare/v0.2.0...v0.2.1

## [0.2.0][] - 2026-03-28

### Added

* Lazy and reusable catalog helpers
  `lint.CodeCatalogHandle` and `lint.CodeCatalogBinding`.
* Scope/stage-based provider registration and snapshot collection filters,
  API helpers and `lintkit snapshot` flags `--scope` and `--stage`.
* Test helpers in `linttest` for deterministic diagnostics comparison
  and code-catalog contract checks.
* `lint.ErrorFromDiagnostics(...)` helper for threshold-based fail mode
  in utility and CI flows.

### Changed

* `CodeCatalogHandle.RuleSpec(...)` now returns `(RuleSpec, error)` to expose
  catalog init failures explicitly.

[0.2.0]: https://github.com/WoozyMasta/lintkit/compare/v0.1.0...v0.2.0

## [0.1.0][] - 2026-03-28

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/lintkit/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html

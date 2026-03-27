# lintkit CLI

`lintkit` CLI is a registry and documentation tool.

It does not lint source files directly.
It works with rule providers and registry snapshots.

## What it does

* `snapshot` collects rule providers and writes deterministic registry data.
  Built-in `lintkit` service diagnostics rules (`LINTKIT100*`) are included.
* `doc` renders documentation from a snapshot.
  Built-in templates: `list`, `table`, `html`.
  Rule blocks render `message` first, then optional detailed `description`.
* `schema` renders JSON Schema for `linting.RunPolicyConfig`.
  Selector enum values are derived from the snapshot rules.
* `template` prints built-in documentation templates.
* `version` prints build metadata.

> [!NOTE]
> Important: `snapshot` executes discovered provider code via `go run`.
> Use it only on trusted code in CI or local development.

## Quick start

Generate registry snapshot:

```bash
go run ./cmd/lintkit snapshot \
  --workdir . \
  docs/lint-rules.snapshot.yaml
```

Render markdown documentation:

```bash
go run ./cmd/lintkit doc \
  -t list \
  -e yaml \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.md
```

Render policy schema:

```bash
go run ./cmd/lintkit schema \
  docs/lint-rules.snapshot.yaml \
  docs/lint-policy.schema.json
```

## Common workflows

Use explicit provider modules instead of dependency discovery:

```bash
go run ./cmd/lintkit snapshot \
  --module github.com/your/moda \
  --module github.com/your/modb \
  docs/lint-rules.snapshot.yaml
```

Provider conflicts are strict by default.
Unknown duplicate rule IDs/codes from multiple providers fail snapshot build.
Use `--soft-providers` to keep first registered rule and continue:

```bash
go run ./cmd/lintkit snapshot \
  --soft-providers \
  docs/lint-rules.snapshot.yaml
```

Render HTML report:

```bash
go run ./cmd/lintkit doc \
  -t html \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.html
```

Check generated docs are up to date:

```bash
go run ./cmd/lintkit doc \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.md \
  --check
```

## Policy schema selector enum

`schema` command can inject allowed selector values
into generated policy schema for field `rules[].rule`.

This is useful when you want strict config validation in editors and CI.
The selector enum can be generated fully, partially, or disabled.

Default mode is `--selector all`.

Disable selector enum generation:

```bash
go run ./cmd/lintkit schema \
  --selector none \
  docs/lint-rules.snapshot.yaml \
  docs/lint-policy.schema.json
```

Explicit selector kinds:

```bash
go run ./cmd/lintkit schema \
  --selector module \
  --selector id \
  --selector code \
  docs/lint-rules.snapshot.yaml \
  docs/lint-policy.schema.json
```

`module + id + code` is equivalent to `all`.

## Related docs

* [../../README.md](../../README.md)
* [../../registry/README.md](../../registry/README.md)
* [../../lint/README.md](../../lint/README.md)
* [../../linting/README.md](../../linting/README.md)

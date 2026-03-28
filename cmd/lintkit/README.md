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
./lintkit snapshot \
  --workdir . \
  docs/lint-rules.snapshot.yaml
```

Render markdown documentation:

```bash
./lintkit doc \
  -t list \
  -e yaml \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.md
```

Render policy schema:

```bash
./lintkit schema \
  docs/lint-rules.snapshot.yaml \
  docs/lint-policy.schema.json
```

## Common workflows

### Snapshot

Use explicit provider modules instead of dependency discovery:

```bash
./lintkit snapshot \
  --module github.com/your/moda \
  --module github.com/your/modb \
  docs/lint-rules.snapshot.yaml
```

Snapshot excludes built-in `lintkit` rules by default.
Use `--include-lintkit-rules` when you want to include them.

Filter collected rules by scope:

```bash
./lintkit snapshot \
  --scope parse \
  --scope validate \
  docs/lint-rules.snapshot.yaml
```

Filter collected rules by stage:

```bash
./lintkit snapshot \
  --stage parse \
  docs/lint-rules.snapshot.yaml
```

Do not combine `--scope` and `--stage` in one run.

Provider conflicts are strict by default.
Unknown duplicate rule IDs/codes from multiple providers fail snapshot build.
Use `--soft-providers` to keep first registered rule and continue:

```bash
./lintkit snapshot \
  --soft-providers \
  docs/lint-rules.snapshot.yaml
```

Keep generated collector source for debugging:

```bash
./lintkit snapshot \
  --temp-dir .tmp \
  --keep-collector \
  docs/lint-rules.snapshot.yaml
```

By default collector source is created in system temp and removed after run.

### Doc

Render HTML report:

```bash
./lintkit doc \
  -t html \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.html
```

Check generated docs are up to date:

```bash
./lintkit doc \
  docs/lint-rules.snapshot.yaml \
  docs/lint-rules.md \
  --check
```

### Schema

`schema` command can inject allowed selector values
into generated policy schema for field `rules[].rule`.

This is useful when you want strict config validation in editors and CI.
The selector enum can be generated fully, partially, or disabled.

Default mode is `--selector all`.

Disable selector enum generation:

```bash
./lintkit schema \
  --selector none \
  docs/lint-rules.snapshot.yaml \
  docs/lint-policy.schema.json
```

Explicit selector kinds:

```bash
./lintkit schema \
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

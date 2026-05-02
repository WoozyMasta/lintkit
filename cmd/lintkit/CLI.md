<!-- markdownlint-disable MD013 MD024 MD036 -->
# lintkit

## NAME

**lintkit**

## SYNOPSIS

`lintkit [OPTIONS]`

## Table of Contents

- [OPTIONS](#options)
- [COMMANDS](#commands)
  - [help](#help)
  - [version](#version)
  - [completion](#completion)
  - [doc](#doc)
  - [docs](#docs)
  - [docs html](#docs-html)
  - [docs man](#docs-man)
  - [docs md](#docs-md)
  - [schema](#schema)
  - [snapshot](#snapshot)
  - [template](#template)

## OPTIONS

### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

## COMMANDS

**Help Commands**

### help

Show help

Print help for CLI and subcommands.

Examples:

- lintkit help
- lintkit help doc

**Usage:** `lintkit [OPTIONS] help`

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

### version

Show version information

Print CLI build and version metadata.

Examples:

- lintkit version

**Usage:** `lintkit [OPTIONS] version`

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

### completion

Generate shell completion

**Usage:** `lintkit [OPTIONS] completion [completion-OPTIONS]`

#### Generate shell completion

|Option|Description|Required|
|---|---|---|
|`--shell SHELL`|Shell completion format; choices: `bash, zsh, pwsh`|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output file path|no|

### docs

Generate documentation

Print generated CLI documentation in selected format.
Supported formats are command subcommands: `man`, `md`, `html`.

Examples:

- lintkit docs md
- lintkit docs html > lintkit.cli.html

**Usage:** `lintkit [OPTIONS] docs`

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

### docs html

Generate HTML documentation

**Usage:** `lintkit [OPTIONS] docs html [html-OPTIONS]`

#### Generate HTML documentation

|Option|Description|Default|Required|
|---|---|---|---|
|`--template TEMPLATE`|HTML documentation template; choices: `default, styled`|default|no|
|`--program-name NAME`|Override program name used in generated documentation templates||no|
|`--toc`|Include table of contents in output||no|
|`--trim-descriptions`|Trim description whitespace in generated output||no|
|`--include-hidden`|Include hidden options, groups and commands||no|
|`--mark-hidden`|Mark hidden entities in documentation output||no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output file path|no|

### docs man

Generate man page documentation

**Usage:** `lintkit [OPTIONS] docs man [man-OPTIONS]`

#### Generate man page documentation

|Option|Description|Required|
|---|---|---|
|`--program-name NAME`|Override program name used in generated documentation templates|no|
|`--trim-descriptions`|Trim description whitespace in generated output|no|
|`--include-hidden`|Include hidden options, groups and commands|no|
|`--mark-hidden`|Mark hidden entities in documentation output|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output file path|no|

### docs md

Generate Markdown documentation

**Usage:** `lintkit [OPTIONS] docs md [md-OPTIONS]`

#### Generate Markdown documentation

|Option|Description|Default|Required|
|---|---|---|---|
|`--template TEMPLATE`|Markdown documentation template; choices: `list, table, code`|list|no|
|`--program-name NAME`|Override program name used in generated documentation templates||no|
|`--toc`|Include table of contents in output||no|
|`--trim-descriptions`|Trim description whitespace in generated output||no|
|`--include-hidden`|Include hidden options, groups and commands||no|
|`--mark-hidden`|Mark hidden entities in documentation output||no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output file path|no|

### doc

Render documentation from registry snapshot

Render documentation from registry snapshot.
Reads snapshot from file argument or stdin; writes docs to file argument or stdout.

Examples:

- lintkit doc rules.snapshot.json docs/lint-rules.md
- lintkit doc -t html rules.snapshot.json docs/lint-rules.html
- lintkit doc rules.snapshot.json docs/lint-rules.md --check

**Usage:** `lintkit [OPTIONS] doc [doc-OPTIONS]`

#### Markdown Render

|Option|Description|Default|Required|
|---|---|---|---|
|`-t`, `--template`|Built-in documentation template style; choices: `list, table, html`|list|no|
|`-p`, `--template-file`|Path to external template file (.gotmpl), overrides built-in template||no|
|`-T`, `--title`|Custom document title in rendered output||no|
|`-D`, `--description`|Custom document description in rendered output||no|
|`-e`, `--example-format`|Append policy example block to markdown output; choices: `json, yaml`||no|
|`-o`, `--toc`|TOC mode for markdown output; choices: `auto, always, off`|auto|no|
|`-w`, `--wrap`|Wrap width for plain text fields in markdown output|80|no|

#### Output Check

|Option|Description|Required|
|---|---|---|
|`-c`, `--check`|Check rendered output against output file and exit non-zero on diff|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`input`|Input registry snapshot path (optional; stdin when omitted)|no|
|`output`|Output documentation path (optional; stdout when omitted)|no|

### schema

Render JSON Schema from registry snapshot

Render RunPolicyConfig JSON Schema from registry snapshot.
Reads snapshot from file argument or stdin; writes schema to file argument or stdout.

Selector enum modes (`--selector`, repeatable):

- `none` disables selector enum injection;
- `all` includes all selector kinds (default);
- `module`, `id`, `code` include selected kinds.

If `none` is present it has priority; if `all` is present it has priority over explicit kinds.

Examples:

- lintkit schema rules.snapshot.yaml docs/lint-policy.schema.json
- lintkit schema -s none rules.snapshot.yaml docs/lint-policy.schema.json
- lintkit schema -s module -s code rules.snapshot.yaml docs/lint-policy.schema.json

**Usage:** `lintkit [OPTIONS] schema [schema-OPTIONS]`

#### Schema Render

|Option|Description|Required|
|---|---|---|
|`-f`, `--format`|Schema output format (inferred from output extension when omitted); choices: `json, yaml`|no|
|`-s`, `--selector`|Selector enum kinds (repeatable); choices: `all, none, module, id, code`|no|

#### Output Check

|Option|Description|Required|
|---|---|---|
|`-c`, `--check`|Check rendered output against output file and exit non-zero on diff|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`input`|Input registry snapshot path (optional; stdin when omitted)|no|
|`output`|Output schema path (optional; stdout when omitted)|no|

### snapshot

Collect rules from modules or dependency graph and render registry snapshot

Collect lint rules from provider packages and write registry snapshot as json/yaml.
By default it auto-discovers providers from current module dependency graph (`go list -deps`).
You can also pass explicit provider packages with `--module`.
Built-in `lintkit` rules are excluded by default;
use `--include-lintkit-rules` to include them.

Examples:

- lintkit snapshot --module github.com/woozymasta/rvcfg rules.snapshot.json
- lintkit snapshot --scope parse --scope validate rules.snapshot.json
- lintkit snapshot --temp-dir .tmp --keep-collector rules.snapshot.json
- lintkit snapshot --include-lintkit-rules rules.snapshot.json

**Usage:** `lintkit [OPTIONS] snapshot [snapshot-OPTIONS]`

#### Snapshot

|Option|Description|Default|Required|
|---|---|---|---|
|`-f`, `--format`|Snapshot output format (inferred from snapshot extension when omitted); choices: `json, yaml`||no|
|`-r`, `--workdir`|Working directory for provider collection commands|.|no|
|`-t`, `--temp-dir`|Directory for generated collector source (default: system temp)||no|
|`-m`, `--module`|Go package import path with LintRulesProvider (repeatable)||no|
|`-p`, `--scope`|Filter providers by rule scope tokens (repeatable)||no|
|`-g`, `--stage`|Filter providers by stage tokens (repeatable); cannot be combined with --scope||no|
|`-i`, `--include-lintkit-rules`|Include built-in lintkit rules in auto-discovery and snapshot output||no|
|`--keep-collector`|Keep generated collector source file for diagnostics||no|
|`-s`, `--soft-providers`|Allow duplicate provider rule conflicts and keep first registered rule||no|

#### Output Check

|Option|Description|Required|
|---|---|---|
|`-c`, `--check`|Check rendered output against output file and exit non-zero on diff|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output snapshot path (optional; stdout when omitted)|no|

### template

Print built-in documentation template

Print built-in documentation template text (`list`, `table`, or `html`).
Use it as a starting point for a custom external template file.

Examples:

- lintkit template > list.gotmpl
- lintkit template -t html docs.html.gotmpl

**Usage:** `lintkit [OPTIONS] template [template-OPTIONS]`

#### Print built-in documentation template

|Option|Description|Default|Required|
|---|---|---|---|
|`-t`, `--template`|Built-in documentation template style; choices: `list, table, html`|list|no|

#### Help Options

|Option|Description|Required|
|---|---|---|
|`-h`, `--help`|Show this help message|no|
|`-v`, `--version`|Show version information|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`output`|Output template file path (optional; stdout when omitted)|no|

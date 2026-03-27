<!-- Automatically generated file, do not modify! -->

# linting config reference

* Source file: [`./cmd/lintkit/internal/app/schema.json`](https://github.com/woozymasta/lintkit/blob/HEAD/cmd/lintkit/internal/app/schema.json)
* Source URL: [Raw schema URL](https://raw.githubusercontent.com/woozymasta/lintkit/HEAD/cmd/lintkit/internal/app/schema.json)
* Schema identifier: `https://github.com/woozymasta/lintkit/linting/run-policy-config`
* JSON Schema version: `https://json-schema.org/draft/2020-12/schema`
* Version support: `supported (2020-12)`
* Root reference: `#/$defs/RunPolicyConfig`

## Contents

* [RunPolicyConfig](#runpolicyconfig)
  * [RunPolicyRuleConfig](#runpolicyruleconfig)
* [Example yaml document](#example-yaml-document)

## RunPolicyConfig

RunPolicyConfig is schema-friendly policy model for config files.

| Attribute | Value |
| --- | --- |
| Type | `object` |
| Properties | 4 |
| Additional properties | boolean schema=false |

### RunPolicyConfig.exclude

Key: `exclude`

Exclude stores optional global exclude path patterns.

| Attribute | Value |
| --- | --- |
| Type | `array` |
| Required | no |
| Items type | `string` |
| Items examples | `**/vendor/**`, `**/*.generated.*` |

### RunPolicyConfig.fail_on

Key: `fail_on`

FailOn defines minimum diagnostic severity that fails run result. Empty value
defaults to "error".

| Attribute | Value |
| --- | --- |
| Type | `string` |
| Required | no |
| Default | `error` |
| Enum | `error`, `warning`, `info`, `notice` |
| Examples | `error` |

### RunPolicyConfig.rules

Key: `rules`

Rules stores ordered selector-based rule settings. Later entries override
earlier entries.

| Attribute | Value |
| --- | --- |
| Type | `array` |
| Required | no |
| Items reference | [`RunPolicyRuleConfig`](#runpolicyruleconfig) (`#/$defs/RunPolicyRuleConfig`) |

### RunPolicyConfig.soft_unknown_selectors

Key: `soft_unknown_selectors`

SoftUnknownSelectors enables soft mode for unknown selectors. Default false
means unknown selectors fail build/compile.

| Attribute | Value |
| --- | --- |
| Type | `boolean` |
| Required | no |
| Default | `false` |

## RunPolicyRuleConfig

RunPolicyRuleConfig is one ordered rule policy entry.

| Attribute | Value |
| --- | --- |
| Type | `object` |
| Properties | 5 |
| Additional properties | boolean schema=false |

### RunPolicyRuleConfig.rule

Key: `rule`

Path: [`rules`](#runpolicyconfigrules).`[]`.`rule`

Rule is rule selector token. Supported forms: `*`, `<module>.*`,
`<module>.<scope>.*`, `<rule-id>`, `<CODE>`.

| Attribute | Value |
| --- | --- |
| Type | `string` |
| Required | yes |
| Examples | `*`, `module_alpha.*`, `module_alpha.parse.*`, `MODULE2001`, `module_alpha.parse.rule-name` |

### RunPolicyRuleConfig.enabled

Key: `enabled`

Path: [`rules`](#runpolicyconfigrules).`[]`.`enabled`

Enabled overrides rule enable state when set.

| Attribute | Value |
| --- | --- |
| Type | `boolean` |
| Required | no |

### RunPolicyRuleConfig.exclude

Key: `exclude`

Path: [`rules`](#runpolicyconfigrules).`[]`.`exclude`

Exclude stores optional path patterns where this entry is not applied.

| Attribute | Value |
| --- | --- |
| Type | `array` |
| Required | no |
| Items type | `string` |
| Items examples | `**/generated/**`, `**/vendor/**` |

### RunPolicyRuleConfig.options

Key: `options`

Path: [`rules`](#runpolicyconfigrules).`[]`.`options`

Arbitrary rule options payload for runner-specific behavior.

| Attribute | Value |
| --- | --- |
| Required | no |

### RunPolicyRuleConfig.severity

Key: `severity`

Path: [`rules`](#runpolicyconfigrules).`[]`.`severity`

Severity overrides effective diagnostic severity when set.

| Attribute | Value |
| --- | --- |
| Type | `string` |
| Required | no |
| Enum | `error`, `warning`, `info`, `notice` |
| Examples | `error` |

## Example yaml document

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/woozymasta/lintkit/HEAD/cmd/lintkit/internal/app/schema.json

# Exclude stores optional global exclude path patterns.
exclude:
  - '**/vendor/**'
  - '**/*.generated.*'
# FailOn defines minimum diagnostic severity that fails run result.
# Empty value defaults to "error".
# Default: error
# Allowed values: error, warning, info, notice
fail_on: error
# Rules stores ordered selector-based rule settings.
# Later entries override earlier entries.
rules:
  - # Enabled overrides rule enable state when set.
    enabled: false
    # Exclude stores optional path patterns where this entry is not applied.
    exclude:
      - '**/generated/**'
      - '**/vendor/**'
    # Arbitrary rule options payload for runner-specific behavior.
    options: null
    # Rule is rule selector token.
    # Supported forms: `*`, `<module>.*`, `<module>.<scope>.*`, `<rule-id>`, `<CODE>`.
    # Example: *
    rule: '*'
    # Severity overrides effective diagnostic severity when set.
    # Example: error
    # Allowed values: error, warning, info, notice
    severity: error
# SoftUnknownSelectors enables soft mode for unknown selectors.
# Default false means unknown selectors fail build/compile.
# Default: false
soft_unknown_selectors: false
```

---

> Generated with
> [schemadoc](https://github.com/woozymasta/schemadoc)
> version `v0.5.1`
> commit `dcbc2d93686e73dd32d4e52452dae3d140f8df46`

<!-- Automatically generated file, do not modify! -->

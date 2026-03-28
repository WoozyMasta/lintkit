// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"fmt"
	"strings"
)

// commandLongDescriptions returns detailed per-command help payload.
func commandLongDescriptions(programName string) map[string]string {
	return map[string]string{
		"template": strings.TrimSpace(fmt.Sprintf(`
Print built-in documentation template text (`+"`list`"+`, `+"`table`"+`, or `+"`html`"+`).
Use it as a starting point for a custom external template file.

Examples:
> $ %[1]s template > list.gotmpl
> $ %[1]s template -t html docs.html.gotmpl
`, programName)),

		"snapshot": strings.TrimSpace(fmt.Sprintf(`
Collect lint rules from provider packages and write registry snapshot as json/yaml.
By default it auto-discovers providers from current module dependency graph (`+"`go list -deps`"+`).
You can also pass explicit provider packages with `+"`--module`"+`.
Built-in `+"`lintkit`"+` rules are excluded by default;
use `+"`--include-lintkit-rules`"+` to include them.

Examples:
> $ %[1]s snapshot --module github.com/woozymasta/rvcfg rules.snapshot.json
> $ %[1]s snapshot --scope parse --scope validate rules.snapshot.json
> $ %[1]s snapshot --temp-dir .tmp --keep-collector rules.snapshot.json
> $ %[1]s snapshot --include-lintkit-rules rules.snapshot.json
`, programName)),

		"doc": strings.TrimSpace(fmt.Sprintf(`
Render documentation from registry snapshot.
Reads snapshot from file argument or stdin; writes docs to file argument or stdout.

Examples:
> $ %[1]s doc rules.snapshot.json docs/lint-rules.md
> $ %[1]s doc -t html rules.snapshot.json docs/lint-rules.html
> $ %[1]s doc rules.snapshot.json docs/lint-rules.md --check
`, programName)),

		"schema": strings.TrimSpace(fmt.Sprintf(`
Render RunPolicyConfig JSON Schema from registry snapshot.
Reads snapshot from file argument or stdin; writes schema to file argument or stdout.

Selector enum modes (`+"`--selector`"+`, repeatable):
* `+"`none`"+` disables selector enum injection;
* `+"`all`"+` includes all selector kinds (default);
* `+"`module`"+`, `+"`id`"+`, `+"`code`"+` include selected kinds.
If `+"`none`"+` is present it has priority; if `+"`all`"+` is present it has priority over explicit kinds.

Examples:
> $ %[1]s schema rules.snapshot.yaml docs/lint-policy.schema.json
> $ %[1]s schema -s none rules.snapshot.yaml docs/lint-policy.schema.json
> $ %[1]s schema -s module -s code rules.snapshot.yaml docs/lint-policy.schema.json
`, programName)),
	}
}

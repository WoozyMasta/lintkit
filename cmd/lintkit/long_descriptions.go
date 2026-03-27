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
> $ %s template > list.gotmpl
> $ %s template -t table table.gotmpl
> $ %s template -t html docs.html.gotmpl
`, programName, programName, programName)),

		"snapshot": strings.TrimSpace(fmt.Sprintf(`
Collect lint rules from provider packages and write registry snapshot as json/yaml.
By default it auto-discovers providers from current module dependency graph (`+"`go list -deps`"+`).
You can also pass explicit provider packages with `+"`--module`"+`.

Examples:
> $ %s snapshot --module github.com/woozymasta/rvcfg rules.snapshot.json
> $ %s snapshot --workdir .. rules.snapshot.json
> $ %s snapshot --workdir .. -f yaml rules.snapshot.yaml
> $ %s snapshot --module github.com/woozymasta/rvcfg --check rules.snapshot.json
`, programName, programName, programName, programName)),

		"doc": strings.TrimSpace(fmt.Sprintf(`
Render documentation from registry snapshot.
Reads snapshot from file argument or stdin; writes docs to file argument or stdout.

Examples:
> $ %s doc rules.snapshot.json docs/lint-rules.md
> $ %s doc rules.snapshot.yaml docs/lint-rules.md
> $ cat rules.snapshot.json | %s doc - docs/lint-rules.md
> $ %s doc -t table -e yaml rules.snapshot.json docs/lint-rules.md
> $ %s doc -t html rules.snapshot.json docs/lint-rules.html
> $ %s doc rules.snapshot.json docs/lint-rules.md --check
`, programName, programName, programName, programName, programName, programName)),

		"schema": strings.TrimSpace(fmt.Sprintf(`
Render RunPolicyConfig JSON Schema from registry snapshot.
Reads snapshot from file argument or stdin; writes schema to file argument or stdout.

Selector enum modes (`+"`--selector`"+`, repeatable):
* `+"`none`"+` disables selector enum injection;
* `+"`all`"+` includes all selector kinds (default);
* `+"`module`"+`, `+"`id`"+`, `+"`code`"+` include selected kinds.
If `+"`none`"+` is present it has priority; if `+"`all`"+` is present it has priority over explicit kinds.

Examples:
> $ %s schema rules.snapshot.yaml docs/lint-policy.schema.json
> $ %s schema -f yaml rules.snapshot.yaml docs/lint-policy.schema.yaml
> $ %s schema -s none rules.snapshot.yaml docs/lint-policy.schema.json
> $ %s schema -s module -s code rules.snapshot.yaml docs/lint-policy.schema.json
`, programName, programName, programName, programName)),
	}
}

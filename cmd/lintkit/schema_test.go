// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRunSchemaDefaultSelectorAll verifies default selector enum injection.
func TestRunSchemaDefaultSelectorAll(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSONTwoModules(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"schema", "-", "-"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(schema default) exit=%d", exitCode)
	}

	enum := readSelectorEnumFromSchemaJSON(t, stdout.Bytes())
	assertContains(t, enum, "*")
	assertContains(t, enum, "module_alpha.*")
	assertContains(t, enum, "module_beta.*")
	assertContains(t, enum, "module_alpha.parse.*")
	assertContains(t, enum, "module_beta.validate.*")
	assertContains(t, enum, "module_alpha.parse.trailing-comma")
	assertContains(t, enum, "module_beta.validate.rule")
	assertContains(t, enum, "RVCFG2020")
	assertContains(t, enum, "RVMAT1001")

	text := stdout.String()
	if !strings.Contains(text, "SoftUnknownSelectors enables soft mode for unknown selectors.") {
		t.Fatalf("expected Go comment-based schema description, got: %q", text)
	}
}

// TestRunSchemaSelectorNonePriority verifies none mode priority over all.
func TestRunSchemaSelectorNonePriority(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSONTwoModules(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"schema", "-", "-", "--selector", "all", "--selector", "none", "--selector", "id"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(schema none priority) exit=%d", exitCode)
	}

	if hasSelectorEnumInSchemaJSON(t, stdout.Bytes()) {
		t.Fatal("expected selector enum to be disabled when --selector none is set")
	}
}

// TestRunSchemaSelectorExplicitAllEquivalence verifies module+id+code equals all.
func TestRunSchemaSelectorExplicitAllEquivalence(t *testing.T) {
	t.Parallel()

	input := testSnapshotJSONTwoModules(t)
	var stdout bytes.Buffer

	exitCode := runWithIO(
		[]string{"schema", "-", "-", "--selector", "module", "--selector", "id", "--selector", "code"},
		bytes.NewReader(input),
		&stdout,
		bytes.NewBuffer(nil),
	)
	if exitCode != 0 {
		t.Fatalf("runWithIO(schema explicit all) exit=%d", exitCode)
	}

	enum := readSelectorEnumFromSchemaJSON(t, stdout.Bytes())
	assertContains(t, enum, "*")
	assertContains(t, enum, "module_alpha.*")
	assertContains(t, enum, "module_beta.*")
	assertContains(t, enum, "module_alpha.parse.*")
	assertContains(t, enum, "module_beta.validate.*")
	assertContains(t, enum, "module_alpha.parse.trailing-comma")
	assertContains(t, enum, "module_beta.validate.rule")
	assertContains(t, enum, "RVCFG2020")
	assertContains(t, enum, "RVMAT1001")
}

// readSelectorEnumFromSchemaJSON extracts rules[].rule enum values.
func readSelectorEnumFromSchemaJSON(t *testing.T, payload []byte) []string {
	t.Helper()

	root := make(map[string]any)
	if err := json.Unmarshal(payload, &root); err != nil {
		t.Fatalf("unmarshal schema json: %v", err)
	}

	node := selectorSchemaNode(t, root)
	raw, ok := node["enum"].([]any)
	if !ok {
		t.Fatalf("selector enum missing: %#v", node)
	}

	result := make([]string, 0, len(raw))
	for index := range raw {
		value, ok := raw[index].(string)
		if !ok {
			t.Fatalf("selector enum contains non-string value: %#v", raw[index])
		}

		result = append(result, value)
	}

	return result
}

// hasSelectorEnumInSchemaJSON reports whether rules rule enum is present.
func hasSelectorEnumInSchemaJSON(t *testing.T, payload []byte) bool {
	t.Helper()

	root := make(map[string]any)
	if err := json.Unmarshal(payload, &root); err != nil {
		t.Fatalf("unmarshal schema json: %v", err)
	}

	node := selectorSchemaNode(t, root)
	_, ok := node["enum"]
	return ok
}

// selectorSchemaNode resolves rules[].rule node from reflected schema root.
func selectorSchemaNode(t *testing.T, root map[string]any) map[string]any {
	t.Helper()

	rootNode := root
	if ref, ok := root["$ref"].(string); ok && strings.HasPrefix(ref, "#/$defs/") {
		defs, ok := root["$defs"].(map[string]any)
		if !ok {
			t.Fatalf("schema $defs missing: %#v", root)
		}

		key := strings.TrimPrefix(ref, "#/$defs/")
		rootNode, ok = defs[key].(map[string]any)
		if !ok {
			t.Fatalf("schema root ref target missing for %q", ref)
		}
	}

	properties, ok := rootNode["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties missing: %#v", rootNode)
	}

	rules, ok := properties["rules"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties.rules missing: %#v", properties)
	}

	items, ok := rules["items"].(map[string]any)
	if !ok {
		t.Fatalf("schema rules.items missing: %#v", rules)
	}

	if ref, ok := items["$ref"].(string); ok && strings.HasPrefix(ref, "#/$defs/") {
		defs, ok := root["$defs"].(map[string]any)
		if !ok {
			t.Fatalf("schema $defs missing: %#v", root)
		}

		key := strings.TrimPrefix(ref, "#/$defs/")
		items, ok = defs[key].(map[string]any)
		if !ok {
			t.Fatalf("schema ref target missing for %q", ref)
		}
	}

	itemProperties, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema items.properties missing: %#v", items)
	}

	selector, ok := itemProperties["rule"].(map[string]any)
	if !ok {
		t.Fatalf("schema rule property missing: %#v", itemProperties)
	}

	return selector
}

// assertContains fails when value does not exist in list.
func assertContains(t *testing.T, list []string, value string) {
	t.Helper()

	for index := range list {
		if list[index] == value {
			return
		}
	}

	t.Fatalf("value %q not found in %v", value, list)
}

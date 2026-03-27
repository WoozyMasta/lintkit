// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package lint

import "testing"

func TestSetGetDeleteRunValue(t *testing.T) {
	t.Parallel()

	var run RunContext

	if !SetRunValue(&run, " test.key ", 42) {
		t.Fatal("SetRunValue()=false, want true")
	}

	got, ok := GetRunValue[int](&run, "test.key")
	if !ok {
		t.Fatal("GetRunValue() ok=false, want true")
	}

	if got != 42 {
		t.Fatalf("GetRunValue()=%d, want 42", got)
	}

	if !DeleteRunValue(&run, " test.key ") {
		t.Fatal("DeleteRunValue()=false, want true")
	}

	if _, ok := GetRunValue[int](&run, "test.key"); ok {
		t.Fatal("GetRunValue() ok=true after delete, want false")
	}
}

func TestRunValueHelpersInvalidInput(t *testing.T) {
	t.Parallel()

	if SetRunValue[int](nil, "test.key", 1) {
		t.Fatal("SetRunValue(nil, ...) returned true, want false")
	}

	if SetRunValue(&RunContext{}, " ", 1) {
		t.Fatal("SetRunValue(empty key) returned true, want false")
	}

	if SetRunValue(&RunContext{}, "key", 1) {
		t.Fatal("SetRunValue(invalid key) returned true, want false")
	}

	if _, ok := GetRunValue[int](&RunContext{}, "test.key"); ok {
		t.Fatal("GetRunValue() ok=true on missing key, want false")
	}

	run := RunContext{
		Values: map[string]any{"test.key": "string"},
	}
	if _, ok := GetRunValue[int](&run, "test.key"); ok {
		t.Fatal("GetRunValue() ok=true on wrong type, want false")
	}

	if DeleteRunValue(nil, "test.key") {
		t.Fatal("DeleteRunValue(nil, ...) returned true, want false")
	}
}

func TestIndexByCode(t *testing.T) {
	t.Parallel()

	type item struct {
		Code string
		ID   int
	}

	items := []item{
		{Code: "A", ID: 1},
		{Code: "B", ID: 2},
		{Code: "A", ID: 3},
	}

	index := IndexByCode(items, func(current item) string {
		return current.Code
	})

	if len(index["A"]) != 2 {
		t.Fatalf("len(index[A])=%d, want 2", len(index["A"]))
	}

	if len(index["B"]) != 1 {
		t.Fatalf("len(index[B])=%d, want 1", len(index["B"]))
	}
}

func TestSetGetIndexedByCode(t *testing.T) {
	t.Parallel()

	type item struct {
		Code string
	}

	run := RunContext{}
	items := []item{
		{Code: "x"},
		{Code: "x"},
		{Code: "y"},
	}

	ok := SetIndexedByCode(&run, "test.items", items, func(current item) string {
		return current.Code
	})
	if !ok {
		t.Fatal("SetIndexedByCode()=false, want true")
	}

	index, ok := GetIndexedByCode[item, string](&run, "test.items")
	if !ok {
		t.Fatal("GetIndexedByCode() ok=false, want true")
	}

	if len(index["x"]) != 2 {
		t.Fatalf("len(index[x])=%d, want 2", len(index["x"]))
	}
}

func TestCurrentRuleOptionsHelpers(t *testing.T) {
	t.Parallel()

	run := RunContext{}

	if !SetCurrentRuleOptions(&run, map[string]any{"max_len": 64}) {
		t.Fatal("SetCurrentRuleOptions()=false, want true")
	}

	options, ok := GetCurrentRuleOptions[map[string]any](&run)
	if !ok {
		t.Fatal("GetCurrentRuleOptions() ok=false, want true")
	}

	if options["max_len"] != 64 {
		t.Fatalf("GetCurrentRuleOptions().max_len=%v, want 64", options["max_len"])
	}

	if !ClearCurrentRuleOptions(&run) {
		t.Fatal("ClearCurrentRuleOptions()=false, want true")
	}

	if _, ok := GetCurrentRuleOptions[map[string]any](&run); ok {
		t.Fatal("GetCurrentRuleOptions() ok=true after clear, want false")
	}
}

func TestValidateRunValueKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key   string
		valid bool
	}{
		{key: "module.key", valid: true},
		{key: "module.key.value", valid: true},
		{key: " module.key ", valid: true},
		{key: "module", valid: false},
		{key: ".module.key", valid: false},
		{key: "module..key", valid: false},
		{key: "1module.key", valid: false},
	}

	for _, testCase := range tests {
		if got := ValidateRunValueKey(testCase.key); got != testCase.valid {
			t.Fatalf(
				"ValidateRunValueKey(%q)=%v, want %v",
				testCase.key,
				got,
				testCase.valid,
			)
		}
	}
}

func TestRunValueNamespacedKeysNoCollision(t *testing.T) {
	t.Parallel()

	run := RunContext{}
	if !SetRunValue(&run, "module_alpha.ast", "alpha") {
		t.Fatal("SetRunValue(module_alpha.ast)=false, want true")
	}

	if !SetRunValue(&run, "module_beta.ast", "beta") {
		t.Fatal("SetRunValue(module_beta.ast)=false, want true")
	}

	alpha, ok := GetRunValue[string](&run, "module_alpha.ast")
	if !ok || alpha != "alpha" {
		t.Fatalf("GetRunValue(module_alpha.ast)=(%q,%v), want (alpha,true)", alpha, ok)
	}

	beta, ok := GetRunValue[string](&run, "module_beta.ast")
	if !ok || beta != "beta" {
		t.Fatalf("GetRunValue(module_beta.ast)=(%q,%v), want (beta,true)", beta, ok)
	}
}

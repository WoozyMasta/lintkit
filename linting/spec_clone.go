// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"reflect"

	"github.com/woozymasta/lintkit/lint"
)

// cloneRuleSpec returns detached copy of one rule spec value.
func cloneRuleSpec(spec lint.RuleSpec) lint.RuleSpec {
	out := spec
	out.DefaultEnabled = cloneBoolPtr(spec.DefaultEnabled)
	out.DefaultOptions = cloneDynamicValue(spec.DefaultOptions)

	if len(spec.FileKinds) == 0 {
		out.FileKinds = nil
		return out
	}

	out.FileKinds = make([]lint.FileKind, len(spec.FileKinds))
	copy(out.FileKinds, spec.FileKinds)

	return out
}

// cloneDynamicValue copies map/slice/pointer-heavy dynamic option payloads.
func cloneDynamicValue(value any) any {
	if value == nil {
		return nil
	}

	cloned := cloneReflectValue(reflect.ValueOf(value))
	if !cloned.IsValid() {
		return nil
	}

	return cloned.Interface()
}

// cloneReflectValue recursively clones dynamic container values.
func cloneReflectValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return reflect.Value{}
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		cloned := cloneReflectValue(value.Elem())
		out := reflect.New(value.Type()).Elem()
		assignValue(out, cloned, value)
		return out
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		out := reflect.New(value.Type().Elem())
		assignValue(out.Elem(), cloneReflectValue(value.Elem()), value.Elem())
		return out
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		out := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			key := iterator.Key()
			keyClone := cloneReflectValue(key)
			if !isAssignableOrConvertibleTo(keyClone, value.Type().Key()) {
				keyClone = key
			}

			item := iterator.Value()
			itemClone := cloneReflectValue(item)
			if !isAssignableOrConvertibleTo(itemClone, value.Type().Elem()) {
				itemClone = item
			}

			out.SetMapIndex(
				convertValueForType(keyClone, value.Type().Key()),
				convertValueForType(itemClone, value.Type().Elem()),
			)
		}

		return out
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}

		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			assignValue(
				out.Index(index),
				cloneReflectValue(value.Index(index)),
				value.Index(index),
			)
		}

		return out
	case reflect.Array:
		out := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			assignValue(
				out.Index(index),
				cloneReflectValue(value.Index(index)),
				value.Index(index),
			)
		}

		return out
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		out.Set(value)
		for index := 0; index < value.NumField(); index++ {
			field := out.Field(index)
			if !field.CanSet() {
				continue
			}

			assignValue(
				field,
				cloneReflectValue(value.Field(index)),
				value.Field(index),
			)
		}

		return out
	default:
		return value
	}
}

// assignValue stores preferred cloned value and falls back to source when needed.
func assignValue(target reflect.Value, preferred reflect.Value, fallback reflect.Value) {
	if converted, ok := convertedValue(preferred, target.Type()); ok {
		target.Set(converted)
		return
	}

	if converted, ok := convertedValue(fallback, target.Type()); ok {
		target.Set(converted)
	}
}

// isAssignableOrConvertibleTo reports whether value can be assigned to target.
func isAssignableOrConvertibleTo(value reflect.Value, targetType reflect.Type) bool {
	if !value.IsValid() {
		return false
	}

	return value.Type().AssignableTo(targetType) ||
		value.Type().ConvertibleTo(targetType)
}

// convertedValue converts or returns value for assignment into target type.
func convertedValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	if !value.IsValid() {
		return reflect.Value{}, false
	}

	if value.Type().AssignableTo(targetType) {
		return value, true
	}

	if value.Type().ConvertibleTo(targetType) {
		return value.Convert(targetType), true
	}

	return reflect.Value{}, false
}

// convertValueForType returns value assignable to target map key/value type.
func convertValueForType(value reflect.Value, targetType reflect.Type) reflect.Value {
	converted, ok := convertedValue(value, targetType)
	if ok {
		return converted
	}

	return reflect.Zero(targetType)
}

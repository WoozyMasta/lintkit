// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"fmt"
	"strings"
)

// setSchemaSelectorEnum writes selector enum into rules[].rule schema node.
func setSchemaSelectorEnum(root map[string]any, selectorEnum []string) error {
	ruleNode, err := resolveRuleSchemaNode(root)
	if err != nil {
		return err
	}

	enum := make([]any, 0, len(selectorEnum))
	for index := range selectorEnum {
		enum = append(enum, selectorEnum[index])
	}

	ruleNode["enum"] = enum
	return nil
}

// clearSchemaSelectorEnum removes selector enum from rules[].rule schema node.
func clearSchemaSelectorEnum(root map[string]any) error {
	ruleNode, err := resolveRuleSchemaNode(root)
	if err != nil {
		return err
	}

	delete(ruleNode, "enum")
	return nil
}

// resolveRuleSchemaNode returns mutable schema node for rules rule.
func resolveRuleSchemaNode(root map[string]any) (map[string]any, error) {
	rootObjectNode, err := resolveSchemaNodeRef(root, root)
	if err != nil {
		return nil, err
	}

	rulesNode, err := resolveObjectProperty(rootObjectNode, "rules")
	if err != nil {
		return nil, err
	}

	itemsNode, ok := rulesNode["items"].(map[string]any)
	if !ok {
		return nil, ErrInvalidPolicySchemaRulesItems
	}

	ruleNode, err := resolveSchemaNodeRef(root, itemsNode)
	if err != nil {
		return nil, err
	}

	ruleNode, err = resolveObjectProperty(ruleNode, "rule")
	if err != nil {
		return nil, err
	}

	return ruleNode, nil
}

// resolveObjectProperty resolves one property schema object from object node.
func resolveObjectProperty(node map[string]any, name string) (map[string]any, error) {
	propertiesNode, ok := node["properties"].(map[string]any)
	if !ok {
		return nil, ErrInvalidPolicySchemaProperties
	}

	propertyNode, ok := propertiesNode[name].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPolicySchemaProperty, name)
	}

	return propertyNode, nil
}

// resolveSchemaNodeRef resolves optional local $ref and returns object node.
func resolveSchemaNodeRef(
	root map[string]any,
	node map[string]any,
) (map[string]any, error) {
	ref, ok := node["$ref"].(string)
	if !ok || strings.TrimSpace(ref) == "" {
		return node, nil
	}

	resolved, err := resolveSchemaRef(root, ref)
	if err != nil {
		return nil, err
	}

	return resolved, nil
}

// resolveSchemaRef resolves local JSON pointer ref ("#/<path>").
func resolveSchemaRef(root map[string]any, ref string) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedSchemaRef, ref)
	}

	current := any(root)
	segments := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	for index := range segments {
		segment := strings.ReplaceAll(segments[index], "~1", "/")
		segment = strings.ReplaceAll(segment, "~0", "~")

		nodeMap, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidSchemaRefPath, ref)
		}

		next, ok := nodeMap[segment]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidSchemaRefPath, ref)
		}

		current = next
	}

	resolved, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSchemaRefPath, ref)
	}

	return resolved, nil
}

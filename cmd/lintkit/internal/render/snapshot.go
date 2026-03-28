// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package render

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// defaultDocumentTitle is fallback documentation title.
	defaultDocumentTitle = "Lint Rules Registry"

	// defaultDocumentDescription is fallback documentation description.
	defaultDocumentDescription = "This document contains the current registry of lint rules."
)

// Snapshot renders markdown snapshot by built-in or file template.
func Snapshot(
	snapshot lint.RegistrySnapshot,
	options Options,
) ([]byte, error) {
	templateName := options.TemplateName
	templatePath := options.TemplatePath
	exampleFormat := options.ExampleFormat
	wrapWidth := options.WrapWidth
	tocMode := options.TOCMode

	if exampleFormat != "" && exampleFormat != "json" && exampleFormat != "yaml" {
		return nil, fmt.Errorf("unsupported example format %q", exampleFormat)
	}

	templateText, err := resolveMarkdownTemplate(templateName, templatePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("rules_markdown").
		Funcs(markdownTemplateFuncs()).
		Parse(templateText)
	if err != nil {
		return nil, fmt.Errorf("parse markdown template: %w", err)
	}

	view := buildMarkdownView(snapshot, wrapWidth, tocMode)
	view.DocumentTitle = strings.TrimSpace(options.DocumentTitle)
	if view.DocumentTitle == "" {
		view.DocumentTitle = defaultDocumentTitle
	}

	view.DocumentDescription = strings.TrimSpace(options.DocumentDescription)
	if view.DocumentDescription == "" {
		view.DocumentDescription = defaultDocumentDescription
	}

	if exampleFormat != "" {
		content, err := renderPolicyExampleContent(snapshot, exampleFormat)
		if err != nil {
			return nil, err
		}

		view.ExampleFormat = exampleFormat
		view.ExampleContent = content
	}

	view.FooterToolName = strings.TrimSpace(options.FooterToolName)
	view.FooterToolURL = strings.TrimSpace(options.FooterToolURL)
	view.FooterVersion = strings.TrimSpace(options.FooterVersion)
	view.FooterCommit = strings.TrimSpace(options.FooterCommit)

	var output bytes.Buffer
	if err := tmpl.Execute(&output, view); err != nil {
		return nil, fmt.Errorf("execute markdown template: %w", err)
	}

	rendered := normalizeLineEndings(output.String())
	if usesHTMLTemplate(templateName, templatePath) {
		rendered = strings.TrimSpace(rendered)
	} else {
		rendered = compactMarkdownBlankLines(rendered)
	}

	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}

	return []byte(rendered), nil
}

// resolveMarkdownTemplate loads markdown template from file or built-in style.
func resolveMarkdownTemplate(name string, filePath string) (string, error) {
	if strings.TrimSpace(filePath) != "" {
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read template file %q: %w", filePath, err)
		}

		return string(payload), nil
	}

	return BuiltinTemplate(name)
}

// BuiltinTemplate returns one built-in markdown template by name.
func BuiltinTemplate(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "list":
		return markdownListTemplate, nil
	case "table":
		return markdownTableTemplate, nil
	case "html":
		return markdownHTMLTemplate, nil
	default:
		return "", fmt.Errorf("unsupported markdown template %q", name)
	}
}

// buildMarkdownView converts snapshot rules to deterministic markdown view.
func buildMarkdownView(
	snapshot lint.RegistrySnapshot,
	wrapWidth int,
	tocMode string,
) View {
	ordered := OrderedSnapshot(snapshot)

	moduleSpecs := make([]lint.ModuleSpec, len(ordered.Modules))
	copy(moduleSpecs, ordered.Modules)
	moduleByID := make(map[string]lint.ModuleSpec, len(moduleSpecs))
	for index := range moduleSpecs {
		moduleID := strings.TrimSpace(moduleSpecs[index].ID)
		if moduleID == "" {
			continue
		}

		moduleByID[moduleID] = moduleSpecs[index]
	}

	rules := make([]lint.RuleSpec, len(ordered.Rules))
	copy(rules, ordered.Rules)

	modules := make([]Module, 0, 8)
	for index := range rules {
		moduleID := strings.TrimSpace(rules[index].Module)
		if moduleID == "" {
			moduleID = "module"
		}

		scopeID := normalizeScopeID(rules[index].Scope)

		if len(modules) == 0 || modules[len(modules)-1].ID != moduleID {
			module := Module{
				ID:   moduleID,
				Name: moduleID,
			}
			if meta, ok := moduleByID[moduleID]; ok {
				module.Name = strings.TrimSpace(meta.Name)
				if module.Name == "" {
					module.Name = moduleID
				}

				module.Description = strings.TrimSpace(
					normalizeLineEndings(meta.Description),
				)
			}

			modules = append(modules, module)
		}

		module := &modules[len(modules)-1]
		if len(module.Scopes) == 0 || module.Scopes[len(module.Scopes)-1].ID != scopeID {
			module.Scopes = append(module.Scopes, Scope{
				ID:     scopeID,
				Anchor: markdownHeadingAnchor(moduleID + "-" + scopeID),
			})
		}

		lastScope := len(module.Scopes) - 1
		if module.Scopes[lastScope].Description == "" {
			module.Scopes[lastScope].Description = strings.TrimSpace(
				normalizeLineEndings(rules[index].ScopeDescription),
			)
		}

		module.Scopes[lastScope].Rules = append(module.Scopes[lastScope].Rules, rules[index])
		module.Scopes[lastScope].RuleCount++
		module.RuleCount++
	}

	normalizedTOCMode := normalizeTOCMode(tocMode)
	showTOC := false
	switch normalizedTOCMode {
	case "always":
		showTOC = len(modules) > 0
	case "auto":
		showTOC = len(modules) > 1
	}

	return View{
		Rules:        rules,
		Modules:      modules,
		ShowTOC:      showTOC,
		ShowScopeTOC: normalizedTOCMode != "off",
		WrapWidth:    normalizeWrapWidth(wrapWidth),
	}
}

type snapshotAnchorRef struct {
	kind  string
	label string
}

// SnapshotAnchorWarnings returns deterministic anchor collision warnings.
func SnapshotAnchorWarnings(
	snapshot lint.RegistrySnapshot,
	options Options,
) []string {
	view := buildMarkdownView(snapshot, options.WrapWidth, options.TOCMode)
	htmlTemplate := usesHTMLTemplate(options.TemplateName, options.TemplatePath)

	anchors := make(map[string][]snapshotAnchorRef, len(view.Rules)*2)
	addAnchor := func(anchor string, kind string, label string) {
		anchor = strings.TrimSpace(anchor)
		if anchor == "" {
			return
		}

		anchors[anchor] = append(anchors[anchor], snapshotAnchorRef{
			kind:  kind,
			label: strings.TrimSpace(label),
		})
	}

	for moduleIndex := range view.Modules {
		module := view.Modules[moduleIndex]
		addAnchor(markdownHeadingAnchor(module.ID), "module", module.ID)

		for scopeIndex := range module.Scopes {
			scope := module.Scopes[scopeIndex]
			scopeAnchor := markdownHeadingAnchor(scope.ID)
			if htmlTemplate {
				scopeAnchor = scope.Anchor
			}

			addAnchor(scopeAnchor, "scope", module.ID+"/"+scope.ID)

			for ruleIndex := range scope.Rules {
				rule := scope.Rules[ruleIndex]
				addAnchor(
					markdownHeadingAnchor(markdownRuleHeading(rule)),
					"rule",
					rule.ID,
				)
			}
		}
	}

	duplicateAnchors := make([]string, 0, 8)
	for anchor := range anchors {
		if len(anchors[anchor]) < 2 {
			continue
		}

		duplicateAnchors = append(duplicateAnchors, anchor)
	}

	if len(duplicateAnchors) == 0 {
		return nil
	}

	slices.Sort(duplicateAnchors)
	warnings := make([]string, 0, len(duplicateAnchors))
	for anchorIndex := range duplicateAnchors {
		anchor := duplicateAnchors[anchorIndex]
		refs := anchors[anchor]

		parts := make([]string, 0, len(refs))
		for refIndex := range refs {
			ref := refs[refIndex]
			parts = append(parts, ref.kind+"="+ref.label)
		}

		warnings = append(
			warnings,
			fmt.Sprintf(
				"anchor fragment #%s is reused by %s",
				anchor,
				strings.Join(parts, ", "),
			),
		)
	}

	return warnings
}

// OrderedSnapshot applies deterministic module/rule ordering.
func OrderedSnapshot(snapshot lint.RegistrySnapshot) lint.RegistrySnapshot {
	moduleSpecs := make([]lint.ModuleSpec, len(snapshot.Modules))
	copy(moduleSpecs, snapshot.Modules)
	sortModuleSpecs(moduleSpecs)

	ruleSpecs := make([]lint.RuleSpec, len(snapshot.Rules))
	copy(ruleSpecs, snapshot.Rules)
	sortRuleSpecs(ruleSpecs)

	return lint.RegistrySnapshot{
		Modules: moduleSpecs,
		Rules:   ruleSpecs,
	}
}

// sortRuleSpecs applies deterministic rule ordering for render output.
func sortRuleSpecs(specs []lint.RuleSpec) {
	slices.SortStableFunc(specs, func(left lint.RuleSpec, right lint.RuleSpec) int {
		if left.Module != right.Module {
			if left.Module < right.Module {
				return -1
			}

			return 1
		}

		leftScope := normalizeScopeID(left.Scope)
		rightScope := normalizeScopeID(right.Scope)
		if leftScope != rightScope {
			if leftScope < rightScope {
				return -1
			}

			return 1
		}

		leftCode, leftCodeOK := lint.ParsePublicCode(left.Code)
		rightCode, rightCodeOK := lint.ParsePublicCode(right.Code)
		if leftCodeOK && rightCodeOK {
			if leftCode != rightCode {
				if leftCode < rightCode {
					return -1
				}

				return 1
			}
		} else if leftCodeOK != rightCodeOK {
			if leftCodeOK {
				return -1
			}

			return 1
		}

		if left.ID != right.ID {
			if left.ID < right.ID {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Message, right.Message)
	})
}

// normalizeScopeID returns canonical scope label for grouped render output.
func normalizeScopeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "general"
	}

	return value
}

// usesHTMLTemplate reports whether built-in or file template targets HTML.
func usesHTMLTemplate(name string, filePath string) bool {
	if strings.TrimSpace(filePath) != "" {
		extension := strings.ToLower(strings.TrimSpace(filepath.Ext(filePath)))
		return extension == ".html" || extension == ".htm"
	}

	return strings.EqualFold(strings.TrimSpace(name), "html")
}

// sortModuleSpecs applies deterministic module metadata ordering for render.
func sortModuleSpecs(specs []lint.ModuleSpec) {
	slices.SortStableFunc(specs, func(left lint.ModuleSpec, right lint.ModuleSpec) int {
		if left.ID != right.ID {
			if left.ID < right.ID {
				return -1
			}

			return 1
		}

		if left.Name != right.Name {
			if left.Name < right.Name {
				return -1
			}

			return 1
		}

		return strings.Compare(left.Description, right.Description)
	})
}

// markdownTOCList renders compact module TOC without empty rows.
func markdownTOCList(modules []Module) string {
	if len(modules) == 0 {
		return ""
	}

	lines := make([]string, 0, len(modules)*2)
	for index := range modules {
		moduleID := strings.TrimSpace(modules[index].ID)
		if moduleID == "" {
			continue
		}

		lines = append(
			lines,
			"* ["+moduleID+"](#"+markdownHeadingAnchor(moduleID)+") ("+
				strconv.Itoa(modules[index].RuleCount)+")",
		)

		for scopeIndex := range modules[index].Scopes {
			scopeID := strings.TrimSpace(modules[index].Scopes[scopeIndex].ID)
			scopeAnchor := markdownHeadingAnchor(scopeID)
			if scopeID == "" || scopeAnchor == "" {
				continue
			}

			lines = append(
				lines,
				"  * ["+scopeID+"](#"+scopeAnchor+") ("+
					strconv.Itoa(modules[index].Scopes[scopeIndex].RuleCount)+")",
			)
		}
	}

	return strings.Join(lines, "\n")
}

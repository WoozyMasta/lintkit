// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package app

import (
	"encoding/json"
	"fmt"

	"github.com/woozymasta/lintkit/cmd/lintkit/internal/render"
	"github.com/woozymasta/lintkit/cmd/lintkit/internal/yamlutil"
	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linting"
)

// renderSnapshot validates snapshot and renders selected output format.
func renderSnapshot(
	snapshot lint.RegistrySnapshot,
	options renderOptions,
) ([]byte, []string, error) {
	registry := linting.NewRegistry()
	for index := range snapshot.Modules {
		if err := registry.RegisterModule(snapshot.Modules[index]); err != nil {
			return nil, nil, fmt.Errorf("register snapshot modules[%d]: %w", index, err)
		}
	}

	if err := registry.RegisterMany(snapshot.Rules...); err != nil {
		return nil, nil, fmt.Errorf("register snapshot rules: %w", err)
	}

	orderedSnapshot := orderedSnapshotForRender(registry.Snapshot())

	switch options.Format {
	case "json":
		if options.Pretty {
			data, err := json.MarshalIndent(orderedSnapshot, "", "  ")
			return data, nil, err
		}

		data, err := json.Marshal(orderedSnapshot)
		return data, nil, err
	case "yaml":
		data, err := yamlutil.Marshal(orderedSnapshot)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal snapshot yaml: %w", err)
		}

		return data, nil, nil
	case "markdown":
		renderOptions := render.Options{
			TemplateName:        options.TemplateName,
			TemplatePath:        options.TemplatePath,
			DocumentTitle:       options.DocumentTitle,
			DocumentDescription: options.DocumentDescription,
			ExampleFormat:       options.ExampleFormat,
			WrapWidth:           options.WrapWidth,
			TOCMode:             options.TOCMode,
			FooterToolName:      options.FooterToolName,
			FooterToolURL:       options.FooterToolURL,
			FooterVersion:       options.FooterVersion,
			FooterCommit:        options.FooterCommit,
		}

		rendered, err := render.Snapshot(orderedSnapshot, renderOptions)
		if err != nil {
			return nil, nil, err
		}

		warnings := render.SnapshotAnchorWarnings(orderedSnapshot, renderOptions)

		return rendered, warnings, nil
	default:
		return nil, nil, fmt.Errorf("unsupported render format %q", options.Format)
	}
}

// orderedSnapshotForRender applies deterministic module/rule ordering.
func orderedSnapshotForRender(
	snapshot lint.RegistrySnapshot,
) lint.RegistrySnapshot {
	return render.OrderedSnapshot(snapshot)
}

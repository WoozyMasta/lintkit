// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import (
	"bytes"
	"fmt"
	"text/template"
)

// BuildCollectorProgram renders helper source for provider collection.
func BuildCollectorProgram(
	packages []string,
	strictProviders bool,
) ([]byte, error) {
	imports := make([]collectorImport, 0, len(packages))
	for index := range packages {
		imports = append(imports, collectorImport{
			Alias: fmt.Sprintf("provider_%d", index),
			Path:  packages[index],
		})
	}

	data := collectorTemplateData{
		Imports:         imports,
		StrictProviders: strictProviders,
	}

	var output bytes.Buffer
	parsedTemplate, err := template.New("provider_collector_program").Parse(
		collectorProgramTemplateText,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: parse template: %w", ErrProviderCollection, err)
	}

	if err := parsedTemplate.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("%w: execute template: %w", ErrProviderCollection, err)
	}

	return output.Bytes(), nil
}

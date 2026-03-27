// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package linting

import (
	"encoding/json"
)

// ExportJSON returns registry snapshot encoded as JSON.
func (registry *Registry) ExportJSON(pretty bool) ([]byte, error) {
	snapshot := registry.Snapshot()
	if pretty {
		return json.MarshalIndent(snapshot, "", "  ")
	}

	return json.Marshal(snapshot)
}

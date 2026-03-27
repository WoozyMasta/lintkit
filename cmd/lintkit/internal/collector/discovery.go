// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// listDependencyPackages lists all transitive packages for module in workDir.
func listDependencyPackages(workDir string) ([]listedPackage, error) {
	command := exec.Command("go", "list", "-deps", "-json", "./...")
	command.Dir = workDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}

		return nil, fmt.Errorf(
			"%w: go list packages from %q: %s",
			ErrProviderDiscovery,
			workDir,
			detail,
		)
	}

	decoder := json.NewDecoder(&stdout)
	packages := make([]listedPackage, 0, 64)
	for {
		var item listedPackage
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf(
				"%w: decode go list package stream: %w",
				ErrProviderDiscovery,
				err,
			)
		}

		if strings.TrimSpace(item.Dir) == "" || len(item.GoFiles) == 0 {
			continue
		}

		packages = append(packages, item)
	}

	return packages, nil
}

// detectProviderImportPaths finds packages exposing LintRulesProvider.
func detectProviderImportPaths(packages []listedPackage) []string {
	fileSet := token.NewFileSet()
	imports := make([]string, 0, len(packages))
	for packageIndex := range packages {
		item := packages[packageIndex]
		if item.Standard || strings.TrimSpace(item.ImportPath) == "" ||
			strings.TrimSpace(item.Dir) == "" || len(item.GoFiles) == 0 {
			continue
		}

		state := directoryState{}
		for fileIndex := range item.GoFiles {
			filePath := filepath.Join(item.Dir, item.GoFiles[fileIndex])
			parsed, _ := parser.ParseFile(
				fileSet,
				filePath,
				nil,
				parser.SkipObjectResolution,
			)
			if parsed == nil {
				// Ignore broken or temporary Go files during provider discovery.
				continue
			}

			applyProviderDeclState(&state, parsed)
		}

		if state.hasProviderType && state.hasRegisterMethod {
			imports = append(imports, item.ImportPath)
		}
	}

	return normalizeModuleImportPaths(imports)
}

// applyProviderDeclState updates provider detection state from one parsed file.
func applyProviderDeclState(state *directoryState, parsed *ast.File) {
	for index := range parsed.Decls {
		switch declaration := parsed.Decls[index].(type) {
		case *ast.GenDecl:
			if declaration.Tok != token.TYPE {
				continue
			}

			for specIndex := range declaration.Specs {
				typeSpec, ok := declaration.Specs[specIndex].(*ast.TypeSpec)
				if !ok {
					continue
				}

				if typeSpec.Name == nil || typeSpec.Name.Name != "LintRulesProvider" {
					continue
				}

				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					state.hasProviderType = true
				}
			}
		case *ast.FuncDecl:
			if declaration.Recv == nil || declaration.Name == nil ||
				declaration.Name.Name != "RegisterRules" {
				continue
			}

			if receiverTypeName(declaration.Recv) == "LintRulesProvider" &&
				isProviderRegisterMethod(declaration) {
				state.hasRegisterMethod = true
			}
		}
	}
}

// isProviderRegisterMethod reports whether RegisterRules has provider shape.
func isProviderRegisterMethod(function *ast.FuncDecl) bool {
	if function == nil || function.Type == nil {
		return false
	}

	if fieldListItemCount(function.Type.Params) != 1 {
		return false
	}

	if fieldListItemCount(function.Type.Results) != 1 {
		return false
	}

	if len(function.Type.Results.List) != 1 {
		return false
	}

	return isErrorType(function.Type.Results.List[0].Type)
}

// fieldListItemCount returns expanded item count for one ast field list.
func fieldListItemCount(list *ast.FieldList) int {
	if list == nil || len(list.List) == 0 {
		return 0
	}

	count := 0
	for index := range list.List {
		if len(list.List[index].Names) == 0 {
			count++
			continue
		}

		count += len(list.List[index].Names)
	}

	return count
}

// isErrorType reports whether type expression is builtin error interface.
func isErrorType(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok || identifier == nil {
		return false
	}

	return identifier.Name == "error"
}

// ensureGoModuleInWorkDir validates presence of go.mod in workdir.
func ensureGoModuleInWorkDir(workDir string) error {
	modulePath := filepath.Join(workDir, "go.mod")
	info, err := os.Stat(modulePath)
	if err != nil || info.IsDir() {
		return fmt.Errorf("%w: %q", ErrNoGoModuleInWorkDir, workDir)
	}

	return nil
}

// normalizeModuleImportPaths trims, deduplicates and sorts import paths.
func normalizeModuleImportPaths(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for index := range values {
		item := strings.TrimSpace(values[index])
		if item == "" {
			continue
		}

		if _, exists := seen[item]; exists {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	slices.Sort(out)
	return out
}

// receiverTypeName extracts receiver base type name from receiver declaration.
func receiverTypeName(receiver *ast.FieldList) string {
	if receiver == nil || len(receiver.List) == 0 || receiver.List[0].Type == nil {
		return ""
	}

	switch item := receiver.List[0].Type.(type) {
	case *ast.Ident:
		return item.Name
	case *ast.StarExpr:
		ident, ok := item.X.(*ast.Ident)
		if !ok {
			return ""
		}

		return ident.Name
	default:
		return ""
	}
}

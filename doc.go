// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/lintkit

/*
Package lintkit is a shared integration layer for modular lint systems.

Import subpackages directly:

  - lintkit/lint defines downstream contracts and catalog helpers.
  - lintkit/linting provides upstream runtime, registry, and policy.
  - lintkit/registry assembles deterministic snapshots in-process.
  - lintkit/linttest provides contract checks for downstream module tests.

The root package intentionally exposes no runtime API.
*/
package lintkit

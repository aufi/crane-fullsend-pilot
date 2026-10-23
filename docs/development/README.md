# Crane Development Guide

This directory contains detailed guides for developers contributing to Crane.

## Guides

- [Architecture](./architecture.md) — Pipeline design, package structure, and data flow
- [Development Setup](./setup.md) — Setting up your development environment
- [Testing](./testing.md) — Unit tests, table-driven patterns, and E2E testing
- [Plugin Development](./plugin-development.md) — Writing custom transformation plugins

## Quick Reference

```bash
# Build all packages and write ./crane
make build

# Cluster-independent unit tests
make test-unit

# Compile E2E tests without running them
make test-e2e-compile

# Check tracked Go formatting without rewriting files
make fmt-check
```

`go test ./...` is not the unit-test command here: it includes the
cluster-backed suites under `e2e-tests/tests/`. Use `make test-unit` for local
unit checks and `make test-e2e-compile` to verify that all E2E packages compile.
Neither command provisions a cluster or executes a live E2E spec.

For the shared project rules, see [AGENTS.md](../../AGENTS.md). Claude Code
loads the same rules through [CLAUDE.md](../../CLAUDE.md).

## Responsibility map

- `cmd/` contains CLI command wiring and user-facing flag validation.
- `internal/` contains Crane-local implementation helpers and pipeline logic.
- `crane-lib` owns reusable transformation functionality; add shared
  transformation behavior there rather than duplicating it in `internal/`.
- `pvc-transfer` owns persistent-volume transfer operations used by
  `transfer-pvc`.

The pipeline is intentionally staged: `export` reads the source cluster,
`transform` processes exported files on disk, and `crane apply` renders output
manifests. Deployment is a separate `kubectl apply` operation. `transfer-pvc`
creates temporary cluster resources and cleans them up as part of the transfer.

## Project Structure

```text
crane/
├── cmd/                    # CLI command implementations
│   ├── apply/              # crane apply
│   ├── export/             # crane export
│   ├── transform/          # crane transform
│   ├── validate/           # crane validate
│   ├── transfer-pvc/       # crane transfer-pvc
│   └── ...
├── internal/               # Internal packages
│   ├── apply/              # Kustomize apply logic
│   ├── kustomize/          # Embedded kustomize runner (krusty API)
│   ├── transform/          # Orchestrator, stages, writer
│   ├── validate/           # Scanner, matcher, report
│   ├── file/               # File helper utilities
│   ├── flags/              # Global CLI flags
│   └── plugin/             # Plugin loading and execution
├── e2e-tests/              # End-to-end tests
│   ├── framework/          # Test framework utilities
│   ├── tests/              # Test cases
│   ├── config/              # Test configuration and flags
│   ├── golden-manifests-ocp/ # Expected output fixtures
│   └── utils/              # Test helper utilities
└── docs/                   # Documentation
```

## Related Repositories

- [konveyor/crane-lib](https://github.com/konveyor/crane-lib) — Transformation logic library
- [konveyor/crane-plugins](https://github.com/konveyor/crane-plugins) — Community plugins
- [konveyor/crane-plugin-openshift](https://github.com/konveyor/crane-plugin-openshift) — OpenShift plugin
- [migtools/pvc-transfer](https://github.com/migtools/pvc-transfer) — PV migration library

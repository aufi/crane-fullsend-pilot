# Contributing to Crane

Thank you for your interest in contributing to Crane.

## Quick Start

1. Fork and clone the repository:

   ```bash
   git clone https://github.com/<your-username>/crane.git
   cd crane
   ```

2. Run the cluster-free checks:

   ```bash
   make build
   make test-unit
   make test-e2e-compile
   make fmt-check
   ```

3. Run the live E2E suite only when the Kubernetes and `k8sdeploy`
   prerequisites described below are available.

4. Submit a pull request.

## Local and live-test prerequisites

The local build and unit checks require only the Go version declared in
`go.mod`; Go may download module dependencies. They do not install plugins,
create clusters, invoke `k8sdeploy`, or require a kubeconfig.

Live E2E execution additionally requires `kubectl`, `k8sdeploy`, two Kubernetes
clusters, and the contexts and images described in [the E2E guide](e2e-tests/README.md).

The unqualified command below is not the unit-test command for this repository:
it includes the cluster-backed packages under `e2e-tests/tests/`.

```bash
go test ./...
```

Run a specific unit package directly when useful, for example:

```bash
go test ./internal/transform/...
```

## Development commands

```bash
# Build all packages and write ./crane
make build

# Unit packages, including e2e-tests/config, framework, and utils
make test-unit

# Compile E2E packages without running specs
make test-e2e-compile

# Check tracked Go files without rewriting them
make fmt-check
```

CI can request verbose unit output and atomic coverage with:

```bash
make test-unit GO_TEST_ARGS=-v COVERAGE_PROFILE=coverage.out COVERAGE_MODE=atomic
```

## Prerequisites

- **Go** 1.25+ (`go.mod` is the source of truth)
- **kubectl** (for deploying output to clusters and for live validation)
- **Docker** (optional, for container-based testing)
- **k8sdeploy** and access to two Kubernetes clusters (for live E2E testing)

## Development Setup

For detailed guides on architecture, testing, and advanced development workflows, see the [Development Documentation](docs/development/README.md).

## Code Standards

### Go Conventions

- Follow standard Go idioms and [Effective Go](https://go.dev/doc/effective_go) practices
- Use `gofmt` for formatting
- Prefer explicit error messages with context — include the actual type received, the API resource being processed, and enough context to debug without re-running

### Error Handling

Error messages must be actionable:

```go
// Good — shows actual object type
fmt.Errorf("expected *unstructured.Unstructured but got %T", object)

// Bad — shows nil pointer
fmt.Errorf("expected *unstructured.Unstructured but got %T", u)
```

### Working with Kubernetes API Objects

- Use `*unstructured.Unstructured` for dynamic resources
- Always check type assertions and provide informative error messages
- Include the actual resource type and API resource name in errors

## Commit conventions

Commit messages use Conventional Commits, such as `fix: handle empty namespace`
or `docs: clarify cluster-free checks`. This is separate from the PR title
format; a valid commit message does not automatically produce a valid PR title.

## PR title conventions

PR titles must begin with one of the aliases accepted by the
[Konveyor versioning documentation](https://github.com/konveyor/release-tools/blob/main/VERSIONING.md):

- `:bug:` for a bug fix — `:bug: Fix empty namespace handling`
- `:book:` for documentation — `:book: Document cluster-free verification`
- `:seedling:` for infrastructure or other changes — `:seedling: Standardize contributor instructions and verification commands`
- `:test_tube:` for integration or E2E tests — `:test_tube: Add offline E2E compilation`

The validator also accepts `:warning:` (breaking), `:sparkles:` (feature),
and `:ghost:` (no release note). Do not substitute a Conventional Commit
prefix such as `fix:` for these PR title aliases.

## Pull Request Guidelines

**Title:** Clear and concise (under 70 characters)

**Body must include:**
- **Summary** — what changed and why
- **Impact** — severity, affected components
- **Test plan** — exact commands run and whether live E2E execution was skipped

**Before submitting:**

- [ ] Build succeeds: `make build`
- [ ] Cluster-free unit tests pass: `make test-unit`
- [ ] E2E packages compile: `make test-e2e-compile`
- [ ] Formatting passes: `make fmt-check`
- [ ] Error messages are informative
- [ ] No unnecessary refactoring in the same PR
- [ ] Backward compatible (or documented breaking change)

## Testing

- All new features require tests
- Bug fixes should include regression tests when possible
- Use table-driven tests for multiple scenarios
- Test files: `*_test.go` in the same package
- E2E tests: `e2e-tests/` (compile with `make test-e2e-compile`; execute only with live prerequisites)

See [Testing Guide](docs/development/testing.md) for detailed testing documentation.

## Reporting Issues

- Check existing issues before filing a new one
- Include reproduction steps, expected behavior, and actual behavior
- Include Crane version, Go version, and Kubernetes version

## Code of Conduct

Be respectful and constructive. We follow the [Konveyor community guidelines](https://www.konveyor.io/).

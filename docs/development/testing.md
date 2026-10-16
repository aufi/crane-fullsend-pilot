# Testing Guide

## Cluster-free checks

### Running Tests

```bash
# Build all packages and write ./crane
make build

# All cluster-independent unit packages
make test-unit

# Compile all E2E test packages, without running specs
make test-e2e-compile

# Check tracked Go files without rewriting them
make fmt-check

# Specific package
go test ./internal/transform/...

# Verbose output
go test -v ./internal/validate/...

# Verbose output and atomic coverage, as used by CI
make test-unit GO_TEST_ARGS=-v COVERAGE_PROFILE=coverage.out COVERAGE_MODE=atomic
go tool cover -func=coverage.out
```

The package list for `make test-unit` comes from `go list ./...` and excludes
only `e2e-tests/tests/`. It therefore includes `e2e-tests/config`,
`e2e-tests/framework`, and `e2e-tests/utils`. A bare `go test ./...` also runs
the Ginkgo suites under `e2e-tests/tests/`, which can require clusters and
external test tooling, so it is not the local unit-test command.

### Test Conventions

- Test files are `*_test.go` in the same package
- Use table-driven tests for multiple scenarios
- Test helpers go in `test_helpers.go` files (e.g., `internal/transform/test_helpers.go`)

### Table-Driven Test Example

```go
func TestValidateStageName(t *testing.T) {
    tests := []struct {
        name      string
        stageName string
        wantErr   bool
    }{
        {"valid plugin stage", "10_KubernetesPlugin", false},
        {"valid custom stage", "30_CustomEdits", false},
        {"missing number", "KubernetesPlugin", true},
        {"missing name", "10_", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateStageName(tt.stageName)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateStageName(%q) error = %v, wantErr %v",
                    tt.stageName, err, tt.wantErr)
            }
        })
    }
}
```

### Testing with Temporary Directories

Many tests create temporary directories for export/transform/output:

```go
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    exportDir := filepath.Join(tmpDir, "export")
    os.MkdirAll(exportDir, 0700)
    // ... test logic
}
```

## E2E Tests

### Location

E2E tests live in `e2e-tests/`:

```text
e2e-tests/
├── framework/              # Test framework (app deployment, kubectl wrappers, pipeline)
├── tests/                  # Test cases (mtc_XXX_*.go)
├── golden-manifests-ocp/   # Expected output fixtures
└── utils/                  # Helper utilities
```

### Compiling E2E tests

Compilation does not run test functions or connect to clusters:

```bash
make test-e2e-compile
```

### Running E2E tests against clusters

E2E tests require running Kubernetes clusters:

```bash
ginkgo run -v --recurse e2e-tests/tests -- \
  --k8sdeploy-bin=/path/to/k8sdeploy \
  --crane-bin=./crane \
  --source-context=src \
  --target-context=tgt
```

This requires the live E2E prerequisites; compilation alone does not prove a
cluster scenario works. The pilot checkout has no active live E2E workflow.
Retained composite actions do not provide live E2E coverage by themselves.

### Golden Manifests

The `golden-manifests-ocp/` directory contains expected export and output for
known applications (e.g., `redis`, `simple-nginx-nopv`). E2E tests compare
actual output against these fixtures.

### Writing E2E Tests

Follow existing patterns in `e2e-tests/tests/`:

1. Use `framework.DeployApp()` to set up the test application
2. Run the Crane pipeline using `framework.RunPipeline()`
3. Compare output against golden manifests or verify specific resource properties

## CI

The current build workflow runs the build and its existing package-filtered
unit check. It does not yet invoke `make test-e2e-compile` or `make fmt-check`,
so those remain local verification commands until the protected workflow is
updated. This pilot checkout has no active live E2E workflow; retained
composite actions are reusable building blocks, not scheduled coverage.

## Test Coverage

To check coverage for a specific package:

```bash
go test -coverprofile=cover.out ./internal/transform/
go tool cover -func=cover.out
```

Guidelines:
- All new features require tests
- Bug fixes should include regression tests
- Focus on testing behavior, not implementation details

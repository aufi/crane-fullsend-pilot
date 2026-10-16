# Development Setup

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ (`go.mod` is authoritative) | Building and cluster-free checks |
| kubectl | Latest stable | Deploying rendered output and live E2E tests |
| Docker | Latest stable | Container-based testing (optional) |
| k8sdeploy and two clusters | Current project test prerequisites | Live E2E execution only |

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/konveyor/crane.git
cd crane
```

### Build

```bash
make build
```

### Install Locally

```bash
go install .
```

Or move the binary to your PATH:

```bash
sudo mv crane /usr/local/bin/
```

### Run Tests

```bash
# Cluster-independent unit tests
make test-unit

# Specific package
go test ./internal/transform/...

# With verbose output
make test-unit GO_TEST_ARGS=-v

# Compile E2E packages without executing specs
make test-e2e-compile

# Check formatting without rewriting files
make fmt-check

# Unit coverage in the same mode used by CI
make test-unit GO_TEST_ARGS=-v COVERAGE_PROFILE=coverage.out COVERAGE_MODE=atomic
go tool cover -func=coverage.out
```

### Format Code

```bash
gofmt -w path/to/changed.go
```

`go test ./...` includes the live suite packages under `e2e-tests/tests/` and
is therefore not a cluster-free unit check. `make test-unit` derives its
package list from the module and excludes only that live-suite subtree; it
still includes `e2e-tests/config`, `e2e-tests/framework`, and `e2e-tests/utils`.

## Project Layout

Commands live in `cmd/<command>/` — each is a self-contained package with a cobra command constructor (`NewXxxCommand`), an `Options` struct holding flags, and `Complete/Validate/Run` methods.

Internal packages in `internal/` contain the business logic. Commands are thin wrappers that parse flags and call into internal packages.

## Setting Up Test Clusters

Live E2E testing requires `kubectl`, `k8sdeploy`, source and target Kubernetes
clusters, valid kubeconfig contexts, and the test images used by the selected
scenarios. The cluster-free build and unit commands above require none of
these tools.

For local clusters:

### Using kind

```bash
# Create source cluster
kind create cluster --name source

# Create target cluster
kind create cluster --name target

# Verify contexts
kubectl config get-contexts
```

### Using minikube

```bash
# Create source cluster
minikube start -p source

# Create target cluster
minikube start -p target
```

## IDE Configuration

### VS Code

Recommended extensions:
- Go (golang.go)
- YAML (redhat.vscode-yaml)

### GoLand / IntelliJ

The project should be auto-detected as a Go module. Ensure the Go SDK is configured to match the version in `go.mod`.

## Common Development Tasks

### Adding a New Command

1. Create `cmd/<command>/<command>.go`
2. Define an `Options` struct with `Complete`, `Validate`, and `Run` methods
3. Create a `NewXxxCommand` function returning `*cobra.Command`
4. Register in `main.go`

### Adding a New Flag

Add the flag in the `addFlagsForOptions` function of the relevant command. Use `mapstructure` tags for viper integration.

### Working with Kubernetes API Objects

Use `*unstructured.Unstructured` for dynamic resources. Always check type assertions:

```go
obj, ok := item.(*unstructured.Unstructured)
if !ok {
    return fmt.Errorf("expected *unstructured.Unstructured but got %T", item)
}
```

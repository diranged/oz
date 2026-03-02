# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Oz RBAC Controller is a Kubernetes operator that provides short-term elevated RBAC privileges to end-users through the native Kubernetes RBAC system. It manages four main Custom Resource Definitions (CRDs):

- **ExecAccessTemplate** / **ExecAccessRequest**: Grant temporary `kubectl exec` access to specific pods
- **PodAccessTemplate** / **PodAccessRequest**: Create temporary dedicated pods for shell access (not in traffic path)

The operator creates short-lived `Role`, `RoleBinding`, and `Pod` resources on-demand based on these templates and requests.

## Development Commands

### Building and Running

```bash
# Build the manager binary
make build

# Build with goreleaser (creates binaries in dist/)
make release

# Build Docker image
make docker-build IMG=<image-name>

# Load Docker image into KIND cluster
make docker-load

# Run the manager locally (without Docker)
make run
```

### Testing

```bash
# Run unit tests (all tests except e2e)
make test

# View test coverage report in terminal
make cover

# View test coverage report in browser
make coverhtml

# Run end-to-end tests (requires KIND cluster with cert-manager)
make test-e2e
```

The test suite uses Ginkgo/Gomega. Individual test files can be run with:
```bash
go test ./internal/path/to/package -v -ginkgo.v
```

### Linting and Formatting

```bash
# Run all linters (revive + golangci-lint)
make lint

# Format code (uses gofumpt + golines)
make fmt
```

### Kubernetes Development

```bash
# Create a KIND cluster for development
kind create cluster

# Install cert-manager (required for webhooks)
make cert-manager

# Generate CRD manifests and RBAC
make manifests

# Generate DeepCopy implementations
make generate

# Install CRDs into cluster
make install

# Deploy controller to cluster
make deploy

# Uninstall CRDs
make uninstall

# Remove controller from cluster
make undeploy
```

### Documentation

```bash
# Generate API documentation (updates API.md)
make godocs

# Generate Helm chart documentation
make helm-docs
```

## Architecture

### Code Organization

- **`internal/api/v1alpha1/`**: CRD definitions and webhook implementations
  - Each CRD has `*_types.go` (API schema) and `*_webhook.go` (admission webhooks)
  - Webhooks use mutating webhooks to inject user context and validating webhooks for logging

- **`internal/controllers/`**: Reconciliation controllers
  - Template controllers validate that templates are properly configured
  - Request controllers orchestrate resource creation via builders
  - Controllers delegate heavy lifting to builders to keep reconcile logic clean

- **`internal/builders/`**: Resource construction logic
  - `execaccessbuilder/`: Creates Role/RoleBinding for exec access
  - `podaccessbuilder/`: Creates Pod/Role/RoleBinding for pod access
  - `utils/`: Shared utilities for resource generation

- **`internal/cmd/ozctl/`**: End-user CLI tool for creating and managing access requests

- **`internal/testing/e2e/`**: End-to-end integration tests

### Key Patterns

1. **Reconciler → Builder separation**: Controllers call builder structs to create/verify resources. This keeps reconcile functions short and testable.

2. **Status conditions**: All CRDs use `.status.conditions[]` to track state (e.g., `TargetRefExists`, `AccessDurationsValid`). Update these when validation fails.

3. **Owner references**: Created resources (Roles, RoleBindings, Pods) have owner references set to the AccessRequest, enabling automatic cleanup via Kubernetes garbage collection.

4. **Admission webhooks**: Mutating webhooks populate user context (from admission.Request), validating webhooks log actions and enforce policies.

5. **Test structure**: Uses Ginkgo/Gomega with `suite_test.go` files for setup. Tests use envtest for realistic K8s API interactions.

## Build System

- Uses **Makefile** for standard targets and **Custom.mk** for project-specific extensions
- Uses **goreleaser** for multi-platform builds and releases
- Docker images are built for linux/arm64, linux/amd64, linux/s390x, linux/ppc64le
- The `main.go` files in `cmd/` are thin wrappers that call `internal/cmd/*/Main()`

## PR Requirements

This repository uses **conventional commits** for PR titles. Valid types:
- `build`: Build system or dependency changes
- `chore`: Maintenance tasks
- `docs`: Documentation changes
- `feat`: New features
- `fix`: Bug fixes

Scope is optional. PR titles are validated by the `pull-request-lint.yaml` workflow.

## Helm Chart

The Helm chart is maintained in `charts/oz/` and deployed separately via the `oz-charts` repository. Local chart testing uses `ct` (chart-testing) with the config in `ct.yaml`.

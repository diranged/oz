# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Oz RBAC Controller is a Kubernetes operator (controller-runtime based) that provides short-term elevated RBAC privileges to end-users through the native Kubernetes RBAC system. It manages four Custom Resource Definitions (CRDs) under the API group `crds.wizardofoz.co/v1alpha1`:

- **ExecAccessTemplate** / **ExecAccessRequest**: Grant temporary `kubectl exec` access to pods managed by an existing controller (Deployment, DaemonSet, StatefulSet, or Argo Rollout)
- **PodAccessTemplate** / **PodAccessRequest**: Create temporary dedicated pods (not in traffic path) for shell access, cloned from a controller's PodSpec with optional mutations

The operator creates short-lived `Role`, `RoleBinding`, and `Pod` resources on-demand. When access expires, the AccessRequest is deleted, cascading via owner references to clean up all created resources.

**Module path**: `github.com/diranged/oz`
**Container image**: `ghcr.io/diranged/oz`

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

- **`internal/api/v1alpha1/`**: CRD type definitions (`*_types.go`), webhook implementations (`*_webhook.go`), shared interfaces, status structs, and condition types
- **`internal/controllers/`**: Generic reconcilers
  - `templatecontroller/`: `TemplateReconciler` - validates templates (target ref exists, durations valid)
  - `requestcontroller/`: `RequestReconciler` - orchestrates access resource creation via builders, handles expiry
  - `podwatcher/`: ValidatingWebhook that audits pod exec/attach events
  - `internal/status/`: Condition setters and status update helpers
  - `internal/ctrlrequeue/`: Requeue helper functions
- **`internal/builders/`**: Resource construction logic implementing `IBuilder`
  - `execaccessbuilder/`: Creates Role/RoleBinding for exec access to existing pods
  - `podaccessbuilder/`: Creates Pod/Role/RoleBinding for dedicated pod access
  - `utils/`: Shared utilities (role creation, pod creation, duration logic, owner refs, access command templating)
- **`internal/webhook/`**: Custom webhook framework extending controller-runtime with contextual admission request support
- **`internal/cmd/manager/`**: Manager entrypoint (registers schemes, reconcilers, webhooks, field indexers)
- **`internal/cmd/ozctl/`**: Cobra-based CLI for end-users to create access requests
- **`internal/testing/e2e/`**: End-to-end integration tests (require KIND cluster)

### Interface Hierarchy

The codebase is built on interfaces that enable generic reconciler logic:

```
ICoreResource (metav1.Object + runtime.Object + GetStatus)
├── ITemplateResource  (+ GetTargetRef, GetAccessConfig)
└── IRequestResource   (+ GetTemplate, GetDuration, GetUptime, GetTemplateName)
    └── IPodRequestResource (+ SetPodName, GetPodName)

ICoreStatus (IsReady, SetReady, GetConditions)
├── IRequestStatus  (+ SetAccessMessage, GetAccessMessage)
└── ITemplateStatus (placeholder)
```

Both `ExecAccessRequest` and `PodAccessRequest` implement `IPodRequestResource`. Both template types implement `ITemplateResource`. When adding new access types, implement these interfaces and the `IBuilder` interface.

### IBuilder Interface

Builders abstract resource-creation logic away from the reconciler:

```go
type IBuilder interface {
    GetTemplate(ctx, client, req) (ITemplateResource, error)
    GetAccessDuration(req, tmpl) (duration, decision, error)
    SetRequestOwnerReference(ctx, client, req, tmpl) error
    CreateAccessResources(ctx, client, req, tmpl) (string, error)
    AccessResourcesAreReady(ctx, client, req, tmpl) (bool, error)
}
```

- **ExecAccessBuilder**: Selects a pod (random or specific via label selector), creates Role + RoleBinding scoped to that pod. `AccessResourcesAreReady()` always returns true.
- **PodAccessBuilder**: Clones PodSpec from controller, applies `PodTemplateSpecMutationConfig` mutations, creates Pod + Role + RoleBinding. `AccessResourcesAreReady()` polls pod for Running/Ready (30s timeout, 1s interval).

### Reconciliation Flows

**TemplateReconciler** (both template types):
1. `fetchRequestObject` - verify resource exists
2. `verifyTargetRef` - check controller target exists (unstructured client)
3. `verifyDuration` - validate defaultDuration <= maxDuration
4. `SetReadyStatus` - set ready=true if all conditions are True
5. Requeue after interval (default: 5min)

**RequestReconciler** (both request types):
1. `fetchRequestObject` - verify resource exists
2. `verifyTemplate` - find template via builder, set owner reference (template owns request)
3. `verifyDuration` - compute effective duration, check expiry
4. `isAccessExpired` - if `ConditionAccessStillValid=False`, **delete** the request (cascading cleanup)
5. `verifyAccessResources` - delegate to builder's CreateAccessResources + AccessResourcesAreReady
6. `SetReadyStatus` - final ready state
7. Requeue after interval

### Owner Reference Chain

```
Template → owns → AccessRequest → owns → Role, RoleBinding, [Pod]
```

Deleting a template cascades to requests, which cascade to all access resources.

### Status Conditions

**Template conditions**: `TemplateDurationsValid`, `TargetRefExists`

**Request conditions**: `TargetTemplateExists`, `AccessDurationsValid`, `AccessStillValid` (False triggers deletion), `AccessResourcesCreated`, `AccessResourcesReady`, `AccessMessage`

`status.ready` is derived: true only when ALL conditions are `ConditionTrue`.

### Webhook System

Custom framework in `internal/webhook/` extends controller-runtime to pass `admission.Request` into defaulter/validator methods (enables access to `req.UserInfo`):
- `IContextuallyDefaultableObject` - `Default(req admission.Request) error`
- `IContextuallyValidatableObject` - `ValidateCreate/Update/Delete(req admission.Request)`
- `PodWatcher` at `/watch-v1-pod` - audit-only webhook for pods/exec and pods/attach events

### Key Code Patterns

1. **Reconciler -> Builder separation**: Controllers delegate resource creation to `IBuilder` implementations, keeping reconcile logic generic and testable.

2. **`shouldReturn` pattern**: Request reconciler steps return `(shouldReturn bool, result, err)`. Caller checks `shouldReturn` to exit the reconcile loop early.

3. **`CreateOrUpdate` for idempotency**: Roles and RoleBindings use controller-runtime's `CreateOrUpdate`. Pods use get-then-create (not CreateOrUpdate) because updating running pods is problematic.

4. **Cached vs non-cached client**: Reconcilers use `client.Client` (cached) for normal operations and `APIReader` (non-cached) for existence checks to avoid stale data.

5. **Field indexers**: Manager registers custom field indexers for `metadata.name` and `status.phase` on Pods, enabling efficient filtered list operations in pod selection.

6. **Resource naming**: Generated resources use `{request-name}-{short-uid}` where short-uid is first 8 chars of the request's UID.

7. **Label purging**: PodAccessBuilder always strips original labels from cloned pods to prevent traffic routing. Custom labels can be added via `controllerTargetMutationConfig.podLabels`.

8. **Event filter**: Both reconcilers use `IgnoreStatusUpdatesAndDeletion()` to avoid infinite reconcile loops from status updates.

9. **Supported controller targets** (enum-validated in CRD): `apps/v1` Deployment, DaemonSet, StatefulSet; `argoproj.io/v1alpha1` Rollout.

## Build System

- Uses **Makefile** for standard targets and **Custom.mk** for project-specific extensions
- Uses **goreleaser** for multi-platform builds and releases
- Docker images are built for linux/arm64, linux/amd64, linux/s390x, linux/ppc64le
- The `main.go` files in `cmd/` are thin wrappers that call `internal/cmd/*/Main()`
- CRD generation: `controller-gen` produces manifests, RBAC, webhooks, and DeepCopy implementations
- Linting: `revive` (config in `revive.toml`) + `golangci-lint` (config in `.golangci.yml`)
- Formatting: `gofumpt` + `golines`

## Testing

- **Unit tests**: Ginkgo v2 + Gomega with envtest (real API server, in-memory). Each package has a `suite_test.go` for setup.
- **E2E tests**: Require KIND cluster with cert-manager. Tests deploy operator and verify full template/request lifecycle including actual `kubectl exec`.
- Run `make test` for unit tests, `make test-e2e` for integration tests.

## PR Requirements

This repository uses **conventional commits** for PR titles. Valid types:
- `build`: Build system or dependency changes
- `chore`: Maintenance tasks
- `docs`: Documentation changes
- `feat`: New features
- `fix`: Bug fixes

Scope is optional. PR titles are validated by the `pull-request-lint.yaml` workflow.

## Helm Chart

The Helm chart is maintained in `charts/oz/` and deployed separately via the `oz-charts` repository. Chart deploys: manager Deployment, RBAC, webhook Service + configurations, cert-manager Certificate, metrics Service. Local chart testing uses `ct` (chart-testing) with the config in `ct.yaml`.

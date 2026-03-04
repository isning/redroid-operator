# Contributing

Thank you for your interest in contributing to `redroid-operator`!

## Development Environment

### Option 1: Nix devshell (recommended)

The repository ships a `flake.nix` that provides a fully reproducible development environment with all required tools pre-installed.

```bash
# Enable flakes if not already done:
# echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

nix develop     # enter devshell
# pre-commit hooks (gofmt + goimports) are automatically installed
```

### Option 2: Manual

Required tools:

| Tool | Version | Purpose |
|---|---|---|
| Go | ≥ 1.22 | Build and test |
| `controller-gen` | v0.17.0 | CRD / RBAC generation |
| `kustomize` | any | YAML generation |
| `helm` | ≥ 3.8 | Chart linting / packaging |
| `golangci-lint` | v2.x | Linting |
| `actionlint` | any | CI workflow validation |

## Project Layout

```
api/v1alpha1/       CRD type definitions (edit → run make generate manifests)
cmd/
  main.go           Controller entrypoint
  kubectl-redroid/  CLI plugin
config/
  crd/bases/        Generated CRD YAML (do not edit by hand)
  rbac/             Generated RBAC manifests
  manager/          Deployment kustomization
charts/
  redroid-operator/ Helm chart (CRDs in crds/ synced from config/crd/bases/)
internal/
  controller/       Reconciler implementations
hack/               Boilerplate headers
```

## Workflow

### Adding / changing API types

1. Edit files under `api/v1alpha1/`
2. Run `make generate` to regenerate DeepCopy methods
3. Run `make manifests` to regenerate CRD YAML and RBAC
4. Run `make helm-crds` to sync CRDs into the Helm chart
5. Run `make docs` to regenerate the CRD reference (`docs/generated/crd-reference.md` is auto-updated on push to `main`)

### Adding a controller feature

1. Implement in `internal/controller/`
2. Add tests in `internal/controller/*_test.go` (uses `envtest` fake client — no cluster needed)
3. Run `make test lint`

### Changing CI workflows

1. Edit `.github/workflows/*.yml`
2. Run `actionlint .github/workflows/*.yml` to validate

## Make Targets

```bash
make help           # list all targets with descriptions

# Development
make generate       # regenerate DeepCopy methods
make manifests      # regenerate CRDs + RBAC
make fmt            # go fmt
make vet            # go vet
make test           # unit tests (verbose + race)
make test-short     # unit tests (fast)
make lint           # golangci-lint
make cover          # coverage report → coverage.html

# Build
make build          # build manager binary
make build-plugin   # build kubectl-redroid

# Helm
make helm-crds      # sync CRDs into chart
make helm-lint      # lint Helm chart
make helm-package   # package chart → dist/

# Docs
make docs           # generate API reference from CRD schema (requires crdoc)
```

## Testing

Tests use `controller-runtime`'s fake client — no real cluster or `envtest` binaries are needed.

```bash
make test               # all tests
go test ./internal/...  # controller tests only
go test -run TestRedroidInstanceReconciler ./internal/controller/
```

## Pull Request Checklist

- [ ] `make test` passes
- [ ] `make lint` reports 0 issues
- [ ] If API types changed: `make generate manifests helm-crds` was run and the generated files are committed
- [ ] CI checks pass (`actionlint` for workflow changes)

## Commit Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat: add TTL support for one-shot tasks
fix: prevent nil pointer when status.suspended.Until is unset
docs: update API reference for InstanceServiceSpec
chore: update controller-gen to v0.17.0
```

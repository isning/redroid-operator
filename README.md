# redroid-operator

> **⚠️ Disclaimer:** This is a toy project built for personal convenience. It comes with **no reliability or security guarantees** of any kind. Use at your own risk.

[![CI](https://github.com/isning/redroid-operator/actions/workflows/ci.yml/badge.svg)](https://github.com/isning/redroid-operator/actions/workflows/ci.yml)
[![Release](https://github.com/isning/redroid-operator/actions/workflows/release.yml/badge.svg)](https://github.com/isning/redroid-operator/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/isning/redroid-operator)](https://goreportcard.com/report/github.com/isning/redroid-operator)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

A Kubernetes operator for managing [Redroid](https://github.com/remote-android/redroid-doc) Android-in-Docker instances and automating integration workloads against them.

## Overview

**redroid-operator** manages two custom resources:

| Resource | Purpose |
|---|---|
| `RedroidInstance` | A persistent Android container backed by overlayfs storage |
| `RedroidTask` | A one-shot or recurring workload (Job / CronJob) that runs tool containers against one or more instances |

Key features:

- **Overlayfs storage model** — a shared read-only base layer (`/data-base`) plus per-instance writable layers (`/data-diff/<index>`), enabling cheap cloning of Android state
- **Service-based ADB access** — every instance gets a dedicated `ClusterIP` Service; optional `NodePort` / `LoadBalancer` exposure
- **Temporary suspend / on-demand wake** — `status.suspended` pauses an instance, `status.woken` forces it running, both without touching `spec`; GitOps tools (Flux, Argo CD) see no drift
- **kubectl plugin** — `kubectl redroid` for port-forward, ADB, logs, and more
- **Helm chart** — full parameterised installation with CRDs included

## Quick Start

### Prerequisites

- Kubernetes ≥ 1.27
- Helm ≥ 3.8 (for OCI chart support)
- A cluster node with KVM/Docker support for Redroid containers (typically a bare-metal or nested-virt node)

### Install via Helm

```bash
helm repo add redroid https://isning.github.io/redroid-operator
helm repo update
helm install redroid-operator redroid/redroid-operator \
  --namespace redroid-system --create-namespace
```

Or install the latest snapshot (tracks `main`):

```bash
helm install redroid-operator redroid/redroid-operator \
  --version 0.0.0-snapshot \
  --namespace redroid-system --create-namespace
```

### Create your first instance

```yaml
# instance.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-0
  namespace: redroid-system
spec:
  index: 0
  image: redroid/redroid:14.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  screen:
    width: 1080
    height: 1920
    dpi: 480
```

```bash
kubectl apply -f instance.yaml
kubectl -n redroid-system get redroidinstances
```

### Connect via ADB

```bash
kubectl redroid instance port-forward android-0 -n redroid-system
# In another terminal:
adb connect localhost:5555
adb shell
```

## Documentation

| Guide | Description |
|---|---|
| [Getting Started](docs/getting-started.md) | Full installation walkthrough |
| [Examples](docs/examples.md) | Real-world patterns: MAA automation, wakeInstance, suspendInstance, base-layer init |
| [API Reference](https://isning.github.io/redroid-operator/docs/generated/crd-reference.md) | `RedroidInstance` and `RedroidTask` field reference (auto-generated from CRD schema) |
| [kubectl Plugin](docs/kubectl-plugin.md) | `kubectl redroid` command reference |
| [Architecture](docs/architecture.md) | Design decisions, overlayfs model, controller flow |
| [Helm Chart Reference](charts/redroid-operator/README.md) | Chart values reference |

## kubectl Plugin

Install `kubectl-redroid`:

```bash
# Download from GitHub Releases (replace VERSION with the desired tag, e.g. v1.0.0)
VERSION=v1.0.0
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-linux-amd64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Latest snapshot (tracks main branch)
curl -L https://github.com/isning/redroid-operator/releases/download/snapshot/kubectl-redroid-0.0.0-snapshot-linux-amd64.tar.gz | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Build from source
make install-plugin
```

### Commands

```
kubectl redroid instance list                    # list instances
kubectl redroid instance port-forward <name>     # forward ADB port to localhost
kubectl redroid instance adb <name> -- <cmd>     # run arbitrary adb command
kubectl redroid instance shell <name>            # interactive adb shell
kubectl redroid instance logs <name>             # stream container logs
kubectl redroid instance suspend <name>          # temporarily stop pod
kubectl redroid instance resume  <name>          # resume stopped instance

kubectl redroid task list                        # list tasks
kubectl redroid task describe <name>             # show task + recent jobs
kubectl redroid task trigger  <name>             # manually run a CronJob now
kubectl redroid task logs     <name>             # stream latest job logs
```

## Development

```bash
# (Optional) Enter the Nix devshell — includes all tools pre-installed
nix develop

# Run tests
make test

# Lint
make lint

# Regenerate CRDs, RBAC manifests, and docs
make manifests
make docs
```

See [Contributing](docs/contributing.md) for more details.

## License

```text
Copyright 2026 ISNing

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

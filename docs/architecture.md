# Architecture

This document describes the design of `redroid-operator`: the overlayfs storage model, the controller reconciliation flow, Service-based ADB access, and the temporary-suspend mechanism.

## High-Level Components

```
┌─────────────────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                             │
│                                                                 │
│  ┌───────────────────────────────┐                             │
│  │  redroid-operator             │                             │
│  │  (controller-manager pod)     │                             │
│  │                               │                             │
│  │  RedroidInstanceReconciler ◄──┼── RedroidInstance CR        │
│  │  RedroidTaskReconciler     ◄──┼── RedroidTask CR            │
│  └───────────┬───────────────────┘                             │
│              │ creates/manages                                  │
│              ▼                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Per-instance Pod + Service                              │  │
│  │                                                          │  │
│  │  Pod: redroid container                                  │  │
│  │       ├── /data-base (RO)  ← sharedDataPVC              │  │
│  │       └── /data-diff/N (RW) ← diffDataPVC               │  │
│  │                                                          │  │
│  │  Service: ClusterIP → Pod:5555 (ADB)                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Per-task Job / CronJob                                  │  │
│  │                                                          │  │
│  │  Pod: sidecar + integration containers                   │  │
│  │       ADB_ADDRESS=<service>:5555                         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Overlayfs Storage Model

Redroid stores Android's `/data` partition on a PersistentVolume. The operator uses an **overlayfs** scheme that allows multiple instances to share a common base state while each maintaining an independent writable layer.

```
┌──────────────────────────────────────────────────────────────┐
│  sharedDataPVC (ReadWriteMany, large)                        │
│  mounted at: /data-base (read-only lower layer)              │
│  contents: base Android system — APKs, accounts, config      │
└──────────────────────────────────────────────────────────────┘
         ▲                             ▲
         ┊  lower layer                ┊  lower layer
         ┊                             ┊
┌────────────────────┐      ┌────────────────────┐
│  diffDataPVC       │      │  diffDataPVC       │
│  /data-diff/0 (RW) │      │  /data-diff/1 (RW) │
│  instance android-0│      │  instance android-1│
└────────────────────┘      └────────────────────┘
```

The operating system inside each container sees `/data` as the merged overlayfs view: reads hit the upper layer first, then fall through to the lower layer; writes go only to the upper layer.

### Implications

- **Storage-efficient** — the base state is stored once and shared; only diffs are duplicated
- **`index` field** — every `RedroidInstance` has a unique `spec.index` that determines the `/data-diff/<index>` subdirectory; two instances with the same index on the same `diffDataPVC` will corrupt each other
- **Base mode** — setting `spec.baseMode: true` mounts `sharedDataPVC` directly as `/data` (read-write), bypassing overlayfs; used for initial setup

## RedroidInstance Reconciler

The reconciler is triggered on every `RedroidInstance` change and runs the following loop:

```
Reconcile(instance)
  ├─ determine desired phase (Running / Stopped) — 4-level priority
  │    ├─ status.woken != nil (not expired) → Running  (overrides everything)
  │    ├─ spec.suspend == true              → Stopped
  │    ├─ status.suspended != nil           → Stopped  (temporary override)
  │    └─ otherwise                         → Running
  │
  ├─ auto-clear expired overrides
  │    ├─ status.woken.Until elapsed   → clear status.woken
  │    └─ status.suspended.Until elapsed → clear status.suspended
  │
  ├─ ensure Pod
  │    ├─ phase == Running → create Pod if not exists, adopt if orphaned
  │    └─ phase == Stopped → delete Pod if exists, wait for termination
  │
  ├─ ensure Service
  │    └─ always create/update ClusterIP Service exposing ADB port
  │
  ├─ update status
  │    ├─ phase, podName, adbAddress
  │    └─ conditions (Ready, Scheduled)
  │
  └─ requeue if pod not yet in Running phase
```

### Pod naming

Each reconciled Pod is named `<instance-name>-pod-<randomSuffix>`. The controller does **not** use `StatefulSet` or `Deployment` — it manages the single Pod directly to give precise control over the overlayfs mount options.

### Service naming

The Service is named identically to the `RedroidInstance` resource. `status.adbAddress` is set to `<service-fqdn>:<adbPort>`.

## RedroidTask Reconciler

```
Reconcile(task)
  ├─ one-shot task (spec.schedule == "")
  │    ├─ if spec.suspendInstance
  │    │    ├─ patch status.suspended on each referenced instance
  │    │    └─ wait until all instance pods are Stopped
  │    ├─ if spec.wakeInstance
  │    │    ├─ patch status.woken on each referenced instance
  │    │    └─ wait until all instance pods are Running
  │    ├─ create Job per instance (or use spec.parallelism to limit concurrency)
  │    ├─ watch Job completion/failure
  │    ├─ clear status.suspended on instances (auto-resume) if suspendInstance
  │    └─ clear status.woken on instances (returns to spec.suspend) if wakeInstance
  │
  └─ scheduled task (spec.schedule != "")
       ├─ create/update CronJob per instance
       └─ sync status from CronJob status
```

### Integration container injection

For each integration container the controller injects:

- `ADB_ADDRESS` — `<service-name>.<namespace>.svc.cluster.local:<adbPort>`
- `INSTANCE_INDEX` — the integer `spec.index` of the target instance

ConfigMap keys from `spec.integrations[].configs` are mounted as volumes at the specified `mountPath`.

### Volume merge semantics

The controller builds the Pod `volumes` list in four layers, with strict override rules:

| Origin | Source | Overrideable? |
|---|---|---|
| **Reserved** | `data-base`, `data-diff`, `dev-dri` (overlayfs + GPU) | Never |
| **Generated** | `cm-<hash>` volumes auto-created from `spec.integrations[].configs` | Never |
| **Task-level** | `spec.volumes` | Yes — by instance-level volumes with the same name |
| **Instance-level** | `spec.instances[].volumes` | Replaces task-level entry of the same name; ignored if reserved/generated |

The same precedence applies to `VolumeMounts` per integration container: `spec.integrations[].volumeMounts` are the base, then config-derived mounts, then `spec.instances[].volumeMounts` (instance wins on `mountPath` collision).

#### ConfigMap volume naming

ConfigMap volumes are named via `ConfigMapVolumeName(configMapName)`: the name is lowercased, dots and underscores replaced with hyphens, and a deterministic 8-character SHA-256 suffix (derived from the original name) is **always** appended.  This guarantees that two ConfigMap names that normalise to the same string (e.g. `foo.bar` and `foo-bar`) still produce distinct volume names, and that the final volume name never exceeds the 63-character DNS label (name) limit imposed on Kubernetes resource names.

## Temporary Suspend / Wake (`status.suspended` / `status.woken`)

A key design goal is compatibility with GitOps tools. If the controller modified `spec.suspend` when automatically pausing an instance for a task, Flux/Argo CD would continuously revert the change, causing reconciliation fights.

The solution: suspension and wake overrides live in **`status`** not **`spec`**. Status is not tracked by GitOps tools. The 4-level priority table governs the desired Pod phase:

```
status.woken   spec.suspend   status.suspended   │  Pod desired phase
───────────────────────────────────────────────────────────────────
non-nil        any            any                │  Running  (wake wins)
nil            false          nil                │  Running  (default)
nil            false          non-nil            │  Stopped  (temp stop)
nil            true           nil                │  Stopped  (GitOps intent)
nil            true           non-nil            │  Stopped
```

The `Until` field allows timed auto-release for both `status.suspended` and `status.woken`:

```yaml
# Temporary stop
status:
  suspended:
    reason: "task/maa-task is running"
    actor: "task/maa-task"
    until: "2025-01-12T04:30:00Z"
```

```yaml
# On-demand wake (set automatically by task controller)
status:
  woken:
    reason: "on-demand wake for one-shot task maa-task"
    actor: "task/maa-task"
```

After `Until` passes, the controller auto-clears the override. The task controller clears `status.woken` after Job completion so the instance returns to the `spec.suspend` state.

## Service-Based Port-Forward

`kubectl-redroid instance port-forward` connects to the **Service** ClusterIP (via the Kubernetes port-forward API), not directly to the Pod. This means:

- The forward still works if the Pod is recreated
- The address is stable even across Pod restarts (same Service name)

## RBAC

The controller requires cluster-level RBAC to manage `Pods`, `Services`, `Jobs`, and `CronJobs` across all namespaces, and `get/list/watch/patch/update` on all `RedroidInstance` and `RedroidTask` resources. The exact permissions are generated from `//+kubebuilder:rbac:...` markers in the controller source and live in `config/rbac/`.

## Webhook (Optional)

Admission webhooks are not currently implemented. Validation is handled by CRD OpenAPI schema rules (`+kubebuilder:validation:...` markers). Defaulting is handled by `+kubebuilder:default:...` markers.

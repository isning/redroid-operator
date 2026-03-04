# API Reference

Group: `redroid.io`  Version: `v1alpha1`

This document is generated from the CRD OpenAPI schema. For auto-generated HTML see the
[online reference](https://isning.github.io/redroid-operator/api/).

---

## RedroidInstance

`RedroidInstance` represents a single persistent Android container instance backed by overlayfs storage.

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-0
  namespace: redroid-system
spec: {}   # see fields below
```

### RedroidInstanceSpec

| Field | Type | Default | Required | Description |
|---|---|---|---|---|
| `index` | `integer` | — | **yes** | Overlayfs partition index. `/data-diff/<index>` is this instance's writable layer. Must be unique per `diffDataPVC`. |
| `image` | `string` | `redroid/redroid:16.0.0-latest` | no | Redroid container image. |
| `imagePullPolicy` | `string` | `IfNotPresent` | no | Image pull policy. One of `Always`, `IfNotPresent`, `Never`. |
| `imagePullSecrets` | `[]LocalObjectReference` | — | no | Secrets for pulling private images. |
| `suspend` | `boolean` | `false` | no | Set `true` to stop the instance Pod without deleting the resource (same semantics as `CronJob.spec.suspend`). |
| `sharedDataPVC` | `string` | `redroid-data-base-pvc` | **yes** | PVC name for the shared `/data-base` volume (read-only lower layer). |
| `diffDataPVC` | `string` | `redroid-data-diff-pvc` | **yes** | PVC name for the per-instance `/data-diff` volume (writable upper layer). |
| `baseMode` | `boolean` | `false` | no | Mount `sharedDataPVC` as `/data` (read-write) and skip overlayfs. Used to initialise the shared base image. See [Base Mode](#base-mode). |
| `gpuMode` | `string` | `host` | no | `androidboot.redroid_gpu_mode`. One of `host`, `guest`, `auto`, `none`. |
| `gpuNode` | `string` | — | no | `androidboot.redroid_gpu_node` DRM device path. Auto-detected when empty. |
| `adbPort` | `integer` | `5555` | no | ADB TCP port exposed by the container. Range: 1–65535. |
| `screen` | `ScreenSpec` | — | no | Virtual display configuration. |
| `network` | `NetworkSpec` | — | no | DNS and proxy configuration. |
| `extraArgs` | `[]string` | — | no | Additional `androidboot.*` arguments. Supports `$(VAR_NAME)` substitution from `extraEnv`. |
| `extraEnv` | `[]EnvVar` | — | no | Extra environment variables. Supports `valueFrom.secretKeyRef` / `configMapKeyRef`. |
| `nodeSelector` | `map[string]string` | — | no | Node selection constraints for the Pod. |
| `tolerations` | `[]Toleration` | — | no | Pod scheduling tolerations. |
| `affinity` | `Affinity` | — | no | Advanced Pod scheduling constraints. |
| `resources` | `ResourceRequirements` | — | no | CPU/memory limits and requests for the redroid container. |
| `service` | `InstanceServiceSpec` | — | no | Customise the Kubernetes Service exposing the ADB port. |

### ScreenSpec

| Field | Type | Default | Description |
|---|---|---|---|
| `width` | `integer` | `720` | Screen width in pixels. min: 1 |
| `height` | `integer` | `1280` | Screen height in pixels. min: 1 |
| `dpi` | `integer` | `320` | Screen density. min: 1 |
| `fps` | `integer` | `30` | Display frame rate. min: 1 |

### NetworkSpec

| Field | Type | Description |
|---|---|---|
| `dns` | `[]string` | DNS server addresses. Maps to `androidboot.redroid_net_ndns` and `androidboot.redroid_net_dns1..N`. |
| `proxy` | `ProxySpec` | HTTP/HTTPS proxy configuration. |

### ProxySpec

| Field | Type | Default | Description |
|---|---|---|---|
| `type` | `string` | — | Proxy mode. One of `static`, `pac`, `none`, `unassigned`. |
| `host` | `string` | — | Proxy server hostname or IP (used with `type: static`). |
| `port` | `integer` | `3128` | Proxy server port. Range: 1–65535. |
| `excludeList` | `string` | — | Comma-separated hosts that bypass the proxy. |
| `pac` | `string` | — | PAC file URL (used with `type: pac`). |

### InstanceServiceSpec

| Field | Type | Default | Description |
|---|---|---|---|
| `type` | `string` | `ClusterIP` | Service type. One of `ClusterIP`, `NodePort`, `LoadBalancer`. |
| `annotations` | `map[string]string` | — | Extra annotations merged onto the Service (e.g. cloud-provider-specific). |
| `nodePort` | `integer` | — | Pin the node port when `type: NodePort`. Auto-assigned if unset. Range: 1–65535. |

### RedroidInstanceStatus

| Field | Type | Description |
|---|---|---|
| `observedGeneration` | `integer` | Most recent generation observed by the controller. |
| `phase` | `string` | Lifecycle phase: `Pending`, `Running`, `Stopped`, or `Failed`. |
| `podName` | `string` | Name of the managed Pod. |
| `adbAddress` | `string` | In-cluster `host:port` to reach this instance's ADB. |
| `conditions` | `[]Condition` | Detailed conditions: `Ready`, `Scheduled`. |
| `suspended` | `SuspendedStatus` | Temporary suspend override (see [Temporary Suspend](#temporary-suspend)). |
| `woken` | `WokenStatus` | Wake override set by the task controller (`wakeInstance`). Forces the instance Running even when `spec.suspend=true`. |

### SuspendedStatus

Set via `kubectl patch` on the status subresource (not reconciled by GitOps tools).

| Field | Type | Description |
|---|---|---|
| `reason` | `string` | Human-readable explanation for the temporary suspend. |
| `until` | `Time` | Optional expiry timestamp. Controller auto-clears once elapsed. |
| `actor` | `string` | Who set the suspend (e.g. `manual`, `task/maa-task`). |

**Manual suspend:**

```bash
kubectl patch redroidinstance android-0 -n redroid-system \
  --subresource=status --type=merge \
  -p '{"status":{"suspended":{"reason":"maintenance","actor":"manual"}}}'
```

**Clear:**

```bash
kubectl patch redroidinstance android-0 -n redroid-system \
  --subresource=status --type=merge \
  -p '{"status":{"suspended":null}}'
```

### WokenStatus

Set automatically by the task controller when `spec.wakeInstance=true`.  Mirrors `SuspendedStatus` but forces the instance **Running** instead of stopping it.

| Field | Type | Description |
|---|---|---|
| `reason` | `string` | Human-readable explanation for the temporary wake. |
| `until` | `Time` | Optional expiry timestamp. Controller auto-clears once elapsed. |
| `actor` | `string` | Who set the wake (e.g. `task/maa-task`). |

**Priority rule:** `status.woken` (highest) → `spec.suspend` → `status.suspended` → default Running.

### Base Mode

When `baseMode: true`:

- `sharedDataPVC` is mounted at `/data` (read-write)
- `diffDataPVC` is **not** mounted
- `androidboot.use_redroid_overlayfs` is set to `0`

Typical workflow:

1. Create a `RedroidInstance` with `baseMode: true`
2. ADB-connect and perform initial setup
3. Suspend or delete the base instance
4. Normal instances (same `sharedDataPVC`) inherit the initialised state

> **Warning:** Do not run a base-mode instance concurrently with normal instances sharing the same `sharedDataPVC`.

### `kubectl get` columns

```
NAME        INDEX   SUSPEND   PHASE     POD                  ADB               AGE
android-0   0       false     Running   android-0-pod-abcd   10.96.0.10:5555   5m
```

---

## RedroidTask

`RedroidTask` describes a workload that runs integration tool containers against one or more `RedroidInstance` overlay partitions.  Each instance spawns its own Job/CronJob execution. The controller injects `ADB_ADDRESS` and `INSTANCE_INDEX` environment variables.

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: daily-run
  namespace: redroid-system
spec: {}   # see fields below
```

### RedroidTaskSpec

| Field | Type | Default | Required | Description |
|---|---|---|---|---|
| `instances` | `[]InstanceRef` | — | **yes** | List of `RedroidInstance` names to target. min: 1 |
| `integrations` | `[]IntegrationSpec` | — | **yes** | Ordered list of tool containers to run per instance. min: 1 |
| `schedule` | `string` | — | no | Cron expression for recurring execution (e.g. `"0 4 * * *"`). Leave empty for a one-shot task. |
| `suspend` | `boolean` | `false` | no | Pause CronJob execution (ignored for one-shot tasks). |
| `timezone` | `string` | — | no | IANA timezone for the CronJob schedule (e.g. `Asia/Shanghai`). Requires Kubernetes ≥ 1.27. |
| `startingDeadlineSeconds` | `integer` | — | no | Deadline (seconds) for starting a missed CronJob run. |
| `backoffLimit` | `integer` | `0` | no | Number of retries before marking the Job failed. min: 0 |
| `suspendInstance` | `boolean` | `false` | no | Temporarily stop the referenced instance Pod while the Job runs, then auto-resume. Only for one-shot tasks. Mutually exclusive with `wakeInstance`. |
| `wakeInstance` | `boolean` | `false` | no | Temporarily start the referenced instance Pod (overrides `spec.suspend`) while the Job runs, then clears the wake-override. Only for one-shot tasks. Mutually exclusive with `suspendInstance`. |
| `activeDeadlineSeconds` | `integer` | — | no | Max duration (seconds) for each Job. min: 1 |
| `ttlSecondsAfterFinished` | `integer` | — | no | Automatically remove completed one-shot Jobs after N seconds. Ignored for scheduled tasks. |
| `parallelism` | `integer` | len(instances) | no | Max concurrent instance Pods. Defaults to run all in parallel. min: 1 |
| `imagePullSecrets` | `[]LocalObjectReference` | — | no | Applied to all integration containers. |
| `successfulJobsHistoryLimit` | `integer` | `3` | no | How many successful CronJob Jobs to retain. |
| `failedJobsHistoryLimit` | `integer` | `3` | no | How many failed CronJob Jobs to retain. |

### InstanceRef

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | **yes** | `RedroidInstance` name in the same namespace. |

### IntegrationSpec

| Field | Type | Default | Required | Description |
|---|---|---|---|---|
| `name` | `string` | — | **yes** | Unique identifier within the task. |
| `image` | `string` | — | **yes** | Container image for this tool. |
| `imagePullPolicy` | `string` | `Always` | no | Image pull policy. |
| `command` | `[]string` | — | no | Override the container entrypoint. |
| `args` | `[]string` | — | no | Arguments passed to the command. |
| `workingDir` | `string` | — | no | Working directory inside the container. |
| `env` | `[]EnvVar` | — | no | Additional env vars, merged after `ADB_ADDRESS` / `INSTANCE_INDEX`. |
| `configs` | `[]ConfigFile` | — | no | ConfigMap keys mounted as files. |
| `volumeMounts` | `[]VolumeMount` | — | no | Extra volume mounts. |
| `serviceAccountName` | `string` | — | no | ServiceAccount for containers that call the Kubernetes API. |
| `securityContext` | `SecurityContext` | — | no | Per-container security options. |
| `resources` | `ResourceRequirements` | — | no | CPU/memory limits/requests. |

**Injected environment variables:**

| Variable | Description |
|---|---|
| `ADB_ADDRESS` | In-cluster `host:port` for the target instance's ADB socket |
| `INSTANCE_INDEX` | The `spec.index` value of the target instance |

### ConfigFile

| Field | Type | Required | Description |
|---|---|---|---|
| `configMapName` | `string` | **yes** | ConfigMap name in the same namespace. |
| `key` | `string` | **yes** | Key within the ConfigMap to mount. |
| `mountPath` | `string` | **yes** | Absolute path inside the container. |

### RedroidTaskStatus

| Field | Type | Description |
|---|---|---|
| `observedGeneration` | `integer` | Most recent generation observed by the controller. |
| `lastScheduleTime` | `Time` | Last time the CronJob was scheduled. |
| `lastSuccessfulTime` | `Time` | Last time a Job completed successfully. |
| `activeJobs` | `[]string` | Names of currently running Jobs. |
| `conditions` | `[]Condition` | Conditions: `Active`, `Complete`, `Failed`. |

### `kubectl get` columns

```
NAME        SCHEDULE    SUSPEND   ACTIVE   LASTSCHEDULE   LASTSUCCESS   AGE
daily-run   0 4 * * *   false     0        10m            10m           1d
```

---

## Examples

### Minimal instance

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-0
spec:
  index: 0
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
```

### Full instance with GPU and networking

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-gpu
spec:
  index: 1
  image: redroid/redroid:14.0.0-latest
  gpuMode: host
  screen:
    width: 1080
    height: 1920
    dpi: 480
    fps: 60
  network:
    dns: ["8.8.8.8", "8.8.4.4"]
  service:
    type: NodePort
    nodePort: 30555
  nodeSelector:
    feature.node.kubernetes.io/kvm: "true"
  resources:
    limits:
      cpu: "4"
      memory: 8Gi
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
```

### Scheduled task with ConfigMap config

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: maa-daily
spec:
  instances:
    - name: android-0
    - name: android-1
  schedule: "0 4 * * *"
  timezone: Asia/Shanghai
  successfulJobsHistoryLimit: 7
  failedJobsHistoryLimit: 3
  integrations:
    - name: maa
      image: ghcr.io/myorg/maa-cli:latest
      configs:
        - configMapName: maa-config
          key: config.json
          mountPath: /config/config.json
      env:
        - name: MAA_LOG_LEVEL
          value: INFO
      resources:
        requests:
          memory: 256Mi
```

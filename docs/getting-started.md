# Getting Started

This guide walks you through installing redroid-operator and running your first Android instance on Kubernetes.

## Prerequisites

| Requirement | Notes |
|---|---|
| Kubernetes ≥ 1.27 | Required for `spec.timeZone` in CronJobs |
| Helm ≥ 3.8 | OCI chart support |
| Node with `/dev/kvm` | Redroid requires hardware virtualisation |
| PersistentVolume provisioner | Any RWX-capable StorageClass |
| `adb` (Android Debug Bridge) | On your local machine for CLI access |

### Check node capabilities

```bash
# KVM must be available on the node running Redroid Pods
kubectl get nodes -o wide
ssh <node> ls /dev/kvm   # should return /dev/kvm
```

## Install the Operator

### Option 1: Helm (recommended)

```bash
helm repo add redroid https://isning.github.io/redroid-operator
helm repo update

helm install redroid-operator redroid/redroid-operator \
  --namespace redroid-system \
  --create-namespace \
  --set installCRDs=true
```

Verify the controller is running:

```bash
kubectl -n redroid-system get pods
# NAME                                        READY   STATUS    RESTARTS   AGE
# redroid-operator-controller-manager-...     1/1     Running   0          30s
```

### Option 2: Kustomize

```bash
# Install CRDs
kubectl apply -k github.com/isning/redroid-operator/config/crd

# Deploy controller
kubectl apply -k github.com/isning/redroid-operator/config
```

## Create Storage

Redroid uses an overlayfs model with two PVCs per node:

- **`redroid-data-base-pvc`** — shared read-only lower layer (installed APKs, initial Android state)
- **`redroid-data-diff-pvc`** — per-instance writable upper layer (diverging state per instance)

```yaml
# storage.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redroid-data-base-pvc
  namespace: redroid-system
spec:
  accessModes: [ReadWriteMany]
  resources:
    requests:
      storage: 20Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redroid-data-diff-pvc
  namespace: redroid-system
spec:
  accessModes: [ReadWriteMany]
  resources:
    requests:
      storage: 10Gi
```

```bash
kubectl apply -f storage.yaml
```

## Base-Mode Bootstrap (Optional)

If you want all instances to share a common pre-installed state (apps, accounts, config), use base mode to initialise `redroid-data-base-pvc` first.

```yaml
# base-instance.yaml
apiVersion: redroid.isning.moe/v1alpha1
kind: RedroidInstance
metadata:
  name: android-base
  namespace: redroid-system
spec:
  index: 0
  image: redroid/redroid:14.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  baseMode: true   # mounts sharedDataPVC as /data (read-write), skips overlayfs
```

```bash
kubectl apply -f base-instance.yaml

# Wait for it to be Running
kubectl -n redroid-system get redroidinstances -w

# Connect and set up (install apps, accounts, etc.)
kubectl redroid instance shell android-base -n redroid-system
# adb install myapp.apk, etc.

# Suspend when done — normal instances now inherit this state
kubectl redroid instance suspend android-base -n redroid-system

# Or delete — the shared PVC data persists
kubectl delete redroidinstance android-base -n redroid-system
```

> **Warning:** Never run base-mode and normal instances concurrently against the same `sharedDataPVC` — this will corrupt the overlayfs lower layer.

## Create an Instance

```yaml
# instance.yaml
apiVersion: redroid.isning.moe/v1alpha1
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
  gpuMode: host
```

```bash
kubectl apply -f instance.yaml
kubectl -n redroid-system get redroidinstances
# NAME        INDEX   SUSPEND   PHASE     POD                    ADB                       AGE
# android-0   0       false     Running   android-0-pod-xxxxx    10.96.123.45:5555         1m
```

## Connect via ADB

### Using kubectl-redroid

```bash
# Install the plugin (replace VERSION with the desired tag, e.g. v1.0.0)
VERSION=v1.0.0
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-linux-amd64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Or install the latest snapshot (tracks main)
curl -L https://github.com/isning/redroid-operator/releases/download/snapshot/kubectl-redroid-0.0.0-snapshot-linux-amd64.tar.gz | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Port-forward and connect
kubectl redroid instance port-forward android-0 -n redroid-system
# → Forwarding  localhost:5555 → android-0  (ADB)
# In a new terminal:
adb connect localhost:5555

# Or launch shell directly (handles port-forward automatically)
kubectl redroid instance shell android-0 -n redroid-system
```

### Manual port-forward

```bash
POD=$(kubectl -n redroid-system get redroidinstance android-0 -o jsonpath='{.status.podName}')
kubectl -n redroid-system port-forward pod/${POD} 5555:5555 &
adb connect localhost:5555
adb shell
```

## Create a Task

A `RedroidTask` runs a workload (sidecar container) against one or more instances. The controller injects `ADB_ADDRESS` automatically.

### One-shot task

```yaml
# task.yaml
apiVersion: redroid.isning.moe/v1alpha1
kind: RedroidTask
metadata:
  name: screenshot
  namespace: redroid-system
spec:
  instances:
    - name: android-0
  suspendInstance: true   # stop the instance pod while the task runs
  integrations:
    - name: screenshotter
      image: ghcr.io/myorg/adb-tools:latest
      command: [sh, -c]
      args:
        - |
          adb -s $ADB_ADDRESS wait-for-device
          adb -s $ADB_ADDRESS shell screencap -p > /output/screenshot.png
```

### Scheduled (CronJob) task

```yaml
apiVersion: redroid.isning.moe/v1alpha1
kind: RedroidTask
metadata:
  name: daily-run
  namespace: redroid-system
spec:
  instances:
    - name: android-0
    - name: android-1
  schedule: "0 4 * * *"   # 04:00 daily
  timezone: Asia/Shanghai
  integrations:
    - name: maa
      image: ghcr.io/myorg/maa-cli:latest
      configs:
        - configMapName: maa-config
          key: config.json
          mountPath: /config/config.json
```

## Expose ADB Outside the Cluster

By default each instance has a `ClusterIP` Service. To expose it externally:

```yaml
spec:
  service:
    type: NodePort
    nodePort: 30555   # optional — omit to auto-assign
```

Or via LoadBalancer with cloud annotations:

```yaml
spec:
  service:
    type: LoadBalancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

## Next Steps

- [Examples](examples.md) — real-world patterns: MAA automation, wakeInstance, suspendInstance, base-layer init
- [API Reference](https://isning.github.io/redroid-operator/docs/generated/crd-reference.md) — all spec fields explained (auto-generated)
- [kubectl Plugin](kubectl-plugin.md) — full CLI reference
- [Architecture](architecture.md) — deep-dive into the overlayfs model and controller design

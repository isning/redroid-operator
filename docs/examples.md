# Examples

Real-world usage patterns for redroid-operator.  Each example is self-contained and can be adapted to your environment.

---

## Example 1 — MAA Daily Automation (Arknights Assistant)

The canonical use case: run [MAA](https://github.com/MaaAssistantArknights/MaaAssistantArknights) automatically every morning against one or more Arknights accounts, each living in its own overlay partition.

### Storage

```yaml
# storage.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redroid-data-base-pvc
  namespace: default
spec:
  accessModes: [ReadWriteMany]
  resources:
    requests:
      storage: 15Gi   # shared read-only base layer (game APK + common data)
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redroid-data-diff-pvc
  namespace: default
spec:
  accessModes: [ReadWriteMany]
  resources:
    requests:
      storage: 15Gi   # per-instance writable overlay (account state diverges here)
```

### Instances

Each account gets its own `RedroidInstance` at a unique `spec.index`.  They share `redroid-data-base-pvc` as a read-only lower layer.

```yaml
# redroid-instances.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: maa-0
  namespace: default
spec:
  index: 0
  image: redroid/redroid:16.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  gpuMode: host
---
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: maa-1
  namespace: default
spec:
  index: 1
  image: redroid/redroid:16.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  gpuMode: host
  suspend: false   # set true to free resources when account is not in use
```

### MAA config

```yaml
# maa-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: maa-config
  namespace: default
data:
  maa-config.json: |
    {
      "adb": {
        "adb_path": "adb",
        "address": "127.0.0.1:5555",
        "config": {}
      },
      "tasks": [
        { "type": "StartUp" },
        { "type": "Infrast" },
        { "type": "Fight", "stage": "1-7" }
      ]
    }
```

### Daily task (CronJob)

```yaml
# maa-daily-task.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: maa-daily
  namespace: default
spec:
  schedule: "0 4 * * *"     # 04:00 every day
  timezone: "Asia/Shanghai"

  instances:
    - name: maa-0
    - name: maa-1            # comment out if maa-1's spec.suspend: true

  integrations:
    - name: maa-cli
      image: ghcr.io/isning/maa-cli-debian:latest
      imagePullPolicy: Always
      command: ["maa"]
      args: ["run", "--config", "/etc/maa/maa-config.json"]
      configs:
        - configMapName: maa-config
          key: maa-config.json
          mountPath: /etc/maa/maa-config.json

  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  activeDeadlineSeconds: 7200   # abort if a run takes more than 2 hours
```

```bash
kubectl apply -f storage.yaml
kubectl apply -f redroid-instances.yaml
kubectl apply -f maa-config.yaml
kubectl apply -f maa-daily-task.yaml

# Watch the daily run
kubectl get redroidtasks -w
kubectl logs -l redroid.io/task=maa-daily -c maa-cli -f
```

---

## Example 2 — On-Demand Wake (`wakeInstance`)

Run MAA against an instance that is normally kept suspended (`spec.suspend: true`) — e.g. a second account that you only want to run on manual trigger.

`wakeInstance: true` tells the task controller to:
1. Set `status.woken` on the instance → instance controller starts the Pod.
2. Wait for `phase == Running`.
3. Execute the Job.
4. Clear `status.woken` → instance controller stops the Pod again.

The `spec.suspend` field is never modified, so GitOps tools (Flux, Argo CD) see no drift.

### Instance (normally off)

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: maa-1
  namespace: default
spec:
  index: 1
  image: redroid/redroid:16.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  gpuMode: host
  suspend: true   # Pod is normally stopped; wakeInstance starts it temporarily
```

### One-shot wake task

```yaml
# maa-wakeinstance-task.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: maa-wake-run
  namespace: default
spec:
  # No schedule = one-shot Job (delete-and-recreate to re-run).
  wakeInstance: true   # powers on instances with spec.suspend: true while Job runs

  instances:
    - name: maa-1

  integrations:
    - name: maa-cli
      image: ghcr.io/isning/maa-cli-debian:latest
      imagePullPolicy: Always
      command: ["maa"]
      args: ["run", "--config", "/etc/maa/maa-config.json"]
      configs:
        - configMapName: maa-config
          key: maa-config.json
          mountPath: /etc/maa/maa-config.json

  ttlSecondsAfterFinished: 3600   # auto-clean Job 1 hour after completion
```

```bash
# Trigger an on-demand run:
kubectl apply -f maa-wakeinstance-task.yaml

# Watch execution:
kubectl get redroidinstances maa-1 -w   # observe Stopped → Running → Stopped

# Or re-trigger by deleting and re-applying:
kubectl delete redroidtask maa-wake-run
kubectl apply -f maa-wakeinstance-task.yaml
```

> **Note:** `wakeInstance` is only meaningful for one-shot tasks (no `spec.schedule`).  For scheduled tasks, instances should be running continuously or managed separately.

---

## Example 3 — Overlayfs-Safe Storage Access (`suspendInstance`)

Some tasks (base-layer update, device image backup) need exclusive write access to the overlayfs storage.  Running them while normal instances have the PVC mounted read-only risks corruption.

`suspendInstance: true` tells the controller to:
1. Set `status.suspended` on the instance → Pod is stopped.
2. Wait for `phase == Stopped`.
3. Execute the Job (now has exclusive storage access).
4. Clear `status.suspended` → Pod restarts automatically.

```yaml
# backup-task.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: diff-backup
  namespace: default
spec:
  suspendInstance: true   # stops instance Pod before Job runs, restarts after

  instances:
    - name: maa-0

  integrations:
    - name: backup
      image: busybox:latest
      command: [sh, -c]
      args:
        - |
          echo "Instance stopped. Safe to access /data-diff."
          tar czf /backup/maa-0-$(date +%F).tar.gz /data-diff/0
      volumeMounts:
        - name: backup-vol
          mountPath: /backup
        - name: diff-vol
          mountPath: /data-diff
```

```bash
kubectl apply -f backup-task.yaml
kubectl logs -l redroid.io/task=diff-backup -c backup -f
```

---

## Example 4 — Base Layer Initialisation

All normal instances share a common read-only base layer (`redroid-data-base-pvc`).  Use a base-mode instance to write the initial state (Android setup, APK installs, account login).

> **Warning:** Never run base-mode and normal instances against the same PVC concurrently.

### One-time setup (manual)

```yaml
# base-init.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-base
  namespace: default
spec:
  index: 255               # high index avoids conflicts with normal instances
  image: redroid/redroid:16.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc   # required by schema; unused in baseMode
  baseMode: true            # mounts sharedDataPVC directly as /data (read-write)
  gpuMode: host
  suspend: false
```

```bash
kubectl apply -f base-init.yaml

# Wait for Running
kubectl get redroidinstances android-base -w

# Connect via ADB (requires kubectl-redroid plugin)
kubectl redroid instance port-forward android-base
# or manually:
kubectl port-forward svc/android-base 5555:5555 &
adb connect localhost:5555

# Install the game APK, log in, complete first-boot, etc.
adb install MAA.apk
adb shell am start -n com.your.game/.MainActivity
# ... perform setup ...

# Suspend when done — all normal instances now inherit this state
kubectl patch redroidinstance android-base -p '{"spec":{"suspend":true}}'
```

### Automated init via RedroidTask

```yaml
# base-init-automated.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: base-init-script
  namespace: default
data:
  init.sh: |
    #!/usr/bin/env sh
    set -e
    until adb connect "$ADB_ADDRESS"; do sleep 5; done
    adb wait-for-device
    sleep 30
    adb install /apks/MAA.apk
    echo "[base-init] done"
---
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: base-init-task
  namespace: default
spec:
  instances:
    - name: android-base
  backoffLimit: 1
  ttlSecondsAfterFinished: 3600
  integrations:
    - name: init
      image: androidsdk/android-32:latest
      command: ["/bin/sh"]
      args: ["/scripts/init.sh"]
      configs:
        - configMapName: base-init-script
          key: init.sh
          mountPath: /scripts/init.sh
```

---

## Example 5 — Multi-Instance with Parallelism Limit

Run the same task against five instances but only two at a time (useful when GPU or network bandwidth is limited).

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: weekly-scan
  namespace: default
spec:
  schedule: "0 3 * * 0"   # every Sunday at 03:00
  timezone: Asia/Tokyo

  parallelism: 2           # at most 2 instance Jobs running simultaneously

  instances:
    - name: android-0
    - name: android-1
    - name: android-2
    - name: android-3
    - name: android-4

  integrations:
    - name: scanner
      image: ghcr.io/myorg/malware-scanner:latest
      args: ["--adb", "$(ADB_ADDRESS)"]
```

---

## Example 6 — Expose ADB Externally

By default each instance has a `ClusterIP` Service.  Override for external access:

### NodePort

```yaml
spec:
  service:
    type: NodePort
    nodePort: 30555   # omit to auto-assign in the NodePort range
```

```bash
# Connect from outside the cluster
adb connect <node-ip>:30555
```

### LoadBalancer (cloud)

```yaml
spec:
  service:
    type: LoadBalancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

---

## Example 7 — Secret-Based Proxy Configuration

Pass proxy credentials from a Kubernetes `Secret` into Android via `androidboot.*` args.

```yaml
# proxy-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: proxy-creds
  namespace: default
stringData:
  host: "proxy.corp.example.com"
  port: "8080"
  user: "android"
  pass: "s3cr3t"
```

```yaml
# proxy-instance.yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidInstance
metadata:
  name: android-proxy
  namespace: default
spec:
  index: 0
  image: redroid/redroid:16.0.0-latest
  sharedDataPVC: redroid-data-base-pvc
  diffDataPVC:   redroid-data-diff-pvc
  gpuMode: host

  extraEnv:
    - name: PROXY_HOST
      valueFrom:
        secretKeyRef:
          name: proxy-creds
          key: host
    - name: PROXY_PORT
      valueFrom:
        secretKeyRef:
          name: proxy-creds
          key: port

  extraArgs:
    - "androidboot.redroid_net_proxy_type=static"
    - "androidboot.redroid_net_proxy_host=$(PROXY_HOST)"
    - "androidboot.redroid_net_proxy_port=$(PROXY_PORT)"
```

---

## Example 8 — Per-Instance Volumes and Secrets

Supply instance-specific credentials (e.g. per-account API tokens stored in separate Secrets) using `spec.instances[].volumes` and `spec.instances[].volumeMounts`.

```yaml
apiVersion: redroid.io/v1alpha1
kind: RedroidTask
metadata:
  name: maa-daily
  namespace: default
spec:
  schedule: "0 4 * * *"

  # Task-level extra volume available to every instance.
  volumes:
    - name: shared-proxy-cert
      configMap:
        name: corp-proxy-ca

  instances:
    - name: maa-0
      # Per-instance Secret: only this instance's Job gets this volume.
      volumes:
        - name: account-token
          secret:
            secretName: maa-0-token   # Secret specific to account 0
      volumeMounts:
        - name: account-token
          mountPath: /run/secrets/token
          subPath: token
          readOnly: true
    - name: maa-1
      volumes:
        - name: account-token
          secret:
            secretName: maa-1-token   # Different Secret for account 1
      volumeMounts:
        - name: account-token
          mountPath: /run/secrets/token
          subPath: token
          readOnly: true

  integrations:
    - name: maa-cli
      image: ghcr.io/isning/maa-cli-debian:latest
      command: ["maa"]
      args: ["run", "--token", "/run/secrets/token"]
      # Integration-level mount present in every instance's container.
      volumeMounts:
        - name: shared-proxy-cert
          mountPath: /etc/ssl/certs/corp-ca.crt
          subPath: ca.crt
          readOnly: true
      configs:
        - configMapName: maa-config
          key: maa-config.json
          mountPath: /etc/maa/maa-config.json
```

**Override rule:** if an instance lists a volume with the same name as a task-level entry in `spec.volumes`, the instance definition wins. Reserved volumes (`data-base`, `data-diff`, `dev-dri`) and controller-generated ConfigMap volumes (`cm-*`) cannot be overridden by either task-level or instance-level volumes.

---

## maa-gitops Reference Layout

The [maa-gitops](https://github.com/isning/maa-gitops) repository puts all the patterns above together in a GitOps-friendly structure managed by Flux:

```
apps/maa/
├── kustomization.yaml        # Flux entry point (excludes manual-only files)
├── redroid-pvc.yaml          # base + diff PVCs
├── maa-config.yaml           # MAA task configuration (ConfigMap)
├── redroid-instances.yaml    # maa-0 (running) + maa-1 (suspend: true)
├── maa-task.yaml             # daily CronJob for running instances
├── maa-wakeinstance-task.yaml  # on-demand one-shot for maa-1 (wakeInstance)
├── base-init.yaml            # one-time base layer setup (NOT in kustomization)
└── base-update.yaml          # daily base APK/resource updater (NOT in kustomization)
```

Key design decisions:
- `spec.suspend` is owned by Git — never mutated at runtime.
- `status.suspended` / `status.woken` are runtime-only — invisible to Flux.
- `maa-1` keeps `spec.suspend: true` in Git; the `maa-wakeinstance-task` wakes it only for the duration of the Job.
- `base-init.yaml` and `base-update.yaml` are excluded from `kustomization.yaml` and applied manually to prevent accidental resets.

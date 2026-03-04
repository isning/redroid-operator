# kubectl redroid — Plugin Reference

`kubectl-redroid` is a `kubectl` plugin that provides convenience commands for managing `RedroidInstance` and `RedroidTask` resources, including port-forwarding, ADB access, and log streaming.

## Installation

### From GitHub Releases

Filenames follow the pattern `kubectl-redroid-<version>-<os>-<arch>.tar.gz` (`.zip` on Windows).

```bash
# Set your desired version
VERSION=v1.0.0

# Linux amd64
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-linux-amd64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Linux arm64
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-linux-arm64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# macOS amd64
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-darwin-amd64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -L "https://github.com/isning/redroid-operator/releases/download/${VERSION}/kubectl-redroid-${VERSION}-darwin-arm64.tar.gz" | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Windows amd64 (PowerShell)
$VERSION = 'v1.0.0'
Invoke-WebRequest "https://github.com/isning/redroid-operator/releases/download/$VERSION/kubectl-redroid-$VERSION-windows-amd64.zip" -OutFile kubectl-redroid.zip
Expand-Archive kubectl-redroid.zip
Move-Item kubectl-redroid\kubectl-redroid.exe 'C:\Windows\System32\'
```

### Latest snapshot (tracks `main`)

```bash
# Linux amd64
curl -L https://github.com/isning/redroid-operator/releases/download/snapshot/kubectl-redroid-0.0.0-snapshot-linux-amd64.tar.gz | tar xz
sudo install kubectl-redroid /usr/local/bin/

# macOS arm64
curl -L https://github.com/isning/redroid-operator/releases/download/snapshot/kubectl-redroid-0.0.0-snapshot-darwin-arm64.tar.gz | tar xz
sudo install kubectl-redroid /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest https://github.com/isning/redroid-operator/releases/download/snapshot/kubectl-redroid-0.0.0-snapshot-windows-amd64.zip -OutFile kubectl-redroid.zip
Expand-Archive kubectl-redroid.zip
Move-Item kubectl-redroid\kubectl-redroid.exe 'C:\Windows\System32\'
```

### From source

```bash
git clone https://github.com/isning/redroid-operator
cd redroid-operator
make install-plugin   # installs to ~/.local/bin/kubectl-redroid
```

### Verify

```bash
kubectl redroid --version
kubectl redroid --help
```

## Global Flags

These flags apply to all subcommands.

| Flag | Shorthand | Default | Description |
|---|---|---|---|
| `--kubeconfig` | — | `$KUBECONFIG` / `~/.kube/config` | Path to kubeconfig file |
| `--context` | — | current context | Kubernetes context to use |
| `--namespace` | `-n` | current context namespace | Namespace to operate in |

---

## instance commands

### `instance list`

List all `RedroidInstance` resources in the namespace.

```bash
kubectl redroid instance list [flags]
kubectl redroid instance ls   # alias
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `-A`, `--all-namespaces` | false | List instances across all namespaces |

**Example output:**

```
NAMESPACE        NAME        INDEX   PHASE     ADB                  AGE
redroid-system   android-0   0       Running   10.96.123.45:5555    5m
redroid-system   android-1   1       Stopped   <none>               5m
```

---

### `instance describe`

Show detailed information about a `RedroidInstance`.

```bash
kubectl redroid instance describe <name> [flags]
kubectl redroid instance get <name>  # alias
```

---

### `instance port-forward`

Forward the ADB port of an instance to `localhost`.

```bash
kubectl redroid instance port-forward <name> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--local-port` | `5555` | Local port to listen on |

The command blocks until interrupted (`Ctrl-C`). While running:

```bash
adb connect localhost:5555
adb devices
```

**Example:**

```bash
kubectl redroid instance port-forward android-0 -n redroid-system --local-port 5556
# Forwarding  localhost:5556 → android-0  (ADB)
```

---

### `instance adb`

Run an arbitrary `adb` command against an instance. Port-forwarding is handled automatically.

```bash
kubectl redroid instance adb <name> -- <adb-args...>
```

**Examples:**

```bash
# Install an APK
kubectl redroid instance adb android-0 -- install myapp.apk

# Push a file
kubectl redroid instance adb android-0 -- push myfile.txt /sdcard/

# Run a command
kubectl redroid instance adb android-0 -- shell pm list packages
```

---

### `instance shell`

Open an interactive `adb shell` session. Port-forwarding is handled automatically.

```bash
kubectl redroid instance shell <name>
```

---

### `instance logs`

Stream logs from the redroid container.

```bash
kubectl redroid instance logs <name> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `-f`, `--follow` | false | Stream logs in real time |
| `--tail` | `-1` (all) | Number of recent lines to show |
| `--since` | — | Show logs since relative time (e.g. `5m`, `1h`) |

---

### `instance suspend`

Temporarily stop the instance Pod by setting `status.suspended` (does not modify `spec.suspend`, so GitOps tools won't fight it).

```bash
kubectl redroid instance suspend <name> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--reason` | `manual` | Human-readable reason for the suspend |
| `--until` | — | Auto-resume at this time (RFC3339 or relative, e.g. `+30m`) |

---

### `instance resume`

Clear `status.suspended` to allow the instance Pod to restart.

```bash
kubectl redroid instance resume <name>
```

---

## task commands

### `task list`

List all `RedroidTask` resources in the namespace.

```bash
kubectl redroid task list [flags]
kubectl redroid task ls   # alias
```

**Example output:**

```
NAMESPACE        NAME        SCHEDULE    SUSPEND   ACTIVE   LAST SCHEDULE   AGE
redroid-system   maa-daily   0 4 * * *   false     0        2h              7d
```

---

### `task describe`

Show detailed status for a `RedroidTask` including recent Job history.

```bash
kubectl redroid task describe <name>
```

---

### `task trigger`

Manually trigger a CronJob-based task immediately (creates a Job from the CronJob template).

```bash
kubectl redroid task trigger <name>
```

For one-shot tasks this command is a no-op (the task runs on creation).

---

### `task logs`

Stream logs from the latest Job spawned by a task.

```bash
kubectl redroid task logs <name> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `-f`, `--follow` | false | Stream logs in real time |
| `--container` | — | Container name within the Job Pod |
| `--tail` | `-1` | Number of recent lines |

---

## Tips

### Shell completion

```bash
# Bash
kubectl redroid completion bash > /etc/bash_completion.d/kubectl-redroid

# Zsh
kubectl redroid completion zsh > "${fpath[1]}/_kubectl-redroid"

# Fish
kubectl redroid completion fish > ~/.config/fish/completions/kubectl-redroid.fish
```

### Using a different namespace by default

```bash
kubectl config set-context --current --namespace=redroid-system
```

### Targeting a specific kubeconfig context

```bash
kubectl redroid --context staging instance list
```

# redroid-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Kubernetes operator for managing Redroid Android container instances

**Homepage:** <https://github.com/isning/redroid-operator>

## Usage

```bash
helm repo add redroid https://isning.github.io/redroid-operator
helm repo update
helm install redroid-operator redroid/redroid-operator \
  --namespace redroid-system --create-namespace
```

Or via OCI:

```bash
helm install redroid-operator \
  oci://ghcr.io/isning/charts/redroid-operator \
  --namespace redroid-system --create-namespace
```

## Source Code

* <https://github.com/isning/redroid-operator>

## Maintainers

| Name | Url |
|------|-----|
| isning | <https://github.com/isning> |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `replicaCount` | int | `1` | Number of controller-manager replicas. |
| `image.repository` | string | `ghcr.io/isning/redroid-operator` | Image repository. |
| `image.pullPolicy` | string | `IfNotPresent` | Image pull policy. |
| `image.tag` | string | `""` | Image tag. Defaults to the chart appVersion. |
| `imagePullSecrets` | list | `[]` | Image pull secrets (list of `{name: ...}`). |
| `nameOverride` | string | `""` | Override the deployed resource names. |
| `fullnameOverride` | string | `""` | Override the full deployed resource name. |
| `serviceAccount.create` | bool | `true` | Whether to create the ServiceAccount. |
| `serviceAccount.annotations` | object | `{}` | Extra annotations to add to the ServiceAccount. |
| `serviceAccount.name` | string | `""` | Override service account name (defaults to chart fullname). |
| `podAnnotations` | object | `{}` | Extra annotations added to the controller-manager Pod. |
| `podSecurityContext` | object | `{runAsNonRoot: true}` | SecurityContext for the controller-manager Pod. |
| `securityContext` | object | `{allowPrivilegeEscalation: false, capabilities: {drop: [ALL]}}` | SecurityContext for the manager container. |
| `resources.limits.cpu` | string | `200m` | CPU limit for the manager container. |
| `resources.limits.memory` | string | `128Mi` | Memory limit for the manager container. |
| `resources.requests.cpu` | string | `10m` | CPU request for the manager container. |
| `resources.requests.memory` | string | `64Mi` | Memory request for the manager container. |
| `nodeSelector` | object | `{}` | Node selector for the controller-manager Pod. |
| `tolerations` | list | `[]` | Tolerations for the controller-manager Pod. |
| `affinity` | object | `{}` | Affinity for the controller-manager Pod. |
| `extraEnv` | list | `[]` | Extra environment variables to set in the manager container. |
| `leaderElection` | bool | `true` | Enable leader election (recommended when `replicaCount > 1`). |
| `metricsPort` | int | `8080` | Metrics port exposed by the manager. |
| `healthzPort` | int | `8081` | Health probe port exposed by the manager. |
| `rbac.create` | bool | `true` | Create ClusterRole and ClusterRoleBinding for the operator. |
| `installCRDs` | bool | `true` | Whether to install CRDs as part of this chart. Set to `false` if you manage CRDs separately. |

# redroid-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Kubernetes operator for managing Redroid Android container instances

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
  --version 0.1.0 \
  --namespace redroid-system --create-namespace
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for the controller-manager Pod. |
| extraEnv | list | `[]` | Extra environment variables to set in the manager container. |
| fullnameOverride | string | `""` |  |
| healthzPort | int | `8081` | Health probe port exposed by the manager. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `"ghcr.io/isning/redroid-operator"` | Image repository. |
| image.tag | string | `""` | Image tag. Defaults to the chart appVersion. |
| imagePullSecrets | list | `[]` | Image pull secrets (list of {name: ...}). |
| installCRDs | bool | `true` | Whether to install CRDs as part of this chart. Set to false if you manage CRDs separately. |
| kmsgToolsImage.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| kmsgToolsImage.repository | string | `"ghcr.io/isning/redroid-operator/kmsg-tools"` | Image repository. |
| kmsgToolsImage.tag | string | `""` | Overrides the kmsg-tools image tag (defaults to the operator's image tag or chart appVersion) |
| leaderElection | bool | `true` | Enable leader election (recommended when replicaCount > 1). |
| metricsPort | int | `8080` | Metrics port exposed by the manager. |
| nameOverride | string | `""` | Override the deployed resource names. |
| nodeSelector | object | `{}` | Node selector for the controller-manager Pod. |
| podAnnotations | object | `{}` | Extra annotations added to the controller-manager Pod. |
| podSecurityContext | object | `{"runAsNonRoot":true}` | SecurityContext for the controller-manager Pod. |
| rbac | object | `{"create":true}` | RBAC: create ClusterRole/Role and their bindings for the operator. |
| replicaCount | int | `1` | Number of controller-manager replicas. |
| resources.limits.cpu | string | `"200m"` |  |
| resources.limits.memory | string | `"128Mi"` |  |
| resources.requests.cpu | string | `"10m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | SecurityContext for the manager container. |
| serviceAccount.annotations | object | `{}` | Extra annotations to add to the ServiceAccount. |
| serviceAccount.create | bool | `true` | Whether to create the ServiceAccount. |
| serviceAccount.name | string | `""` | Override service account name (defaults to chart fullname). |
| tolerations | list | `[]` | Tolerations for the controller-manager Pod. |
| watchNamespaces | object | `{"enabled":false,"namespaces":[]}` | Namespaced-mode configuration. When enabled is false (default) the operator runs cluster-wide:   a ClusterRole + ClusterRoleBinding are created. When enabled is true the operator watches only the listed namespaces:   a Role + RoleBinding are created per namespace instead. |

## Source Code

* <https://github.com/isning/redroid-operator>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| isning |  | <https://github.com/isning> |

# Resource Reaper Helm Chart

This Helm chart deploys the Resource Reaper, which manages Kubernetes resource lifecycles using TTL annotations.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+

## Installing the Chart

```bash
# Add the repository (once available)
helm repo add resource-reaper https://[REPO_URL]
helm repo update

# Install the chart with the release name 'resource-reaper'
helm install resource-reaper resource-reaper/resource-reaper
```

Or install from local directory:

```bash
helm install resource-reaper ./helm/resource-reaper
```

## Configuration

The following table lists the configurable parameters of the Resource Reaper chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of Resource Reaper replicas | `1` |
| `image.repository` | Image repository | `resource-reaper` |
| `image.tag` | Image tag | `0.1.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `controller.leaderElection.enabled` | Enable leader election | `true` |
| `controller.metrics.enabled` | Enable metrics | `true` |
| `controller.metrics.port` | Metrics port | `8080` |
| `controller.logLevel` | Log level | `info` |
| `controller.logFormat` | Log format | `json` |

## Example Installation

```bash
helm install resource-reaper ./helm/resource-reaper \
  --set image.repository=your-registry/resource-reaper \
  --set image.tag=latest \
  --set controller.logLevel=debug
```

## Uninstalling the Chart

```bash
helm uninstall resource-reaper
```

## Development

To package the chart:

```bash
helm package ./helm/resource-reaper
```

To validate the chart:


```bash
helm lint ./helm/resource-reaper
helm template ./helm/resource-reaper

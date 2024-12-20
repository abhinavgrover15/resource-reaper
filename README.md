# Resource Reaper

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/abhinavgrover15/resource-reaper)](https://github.com/abhinavgrover15/resource-reaper/releases)
[![CI](https://github.com/abhinavgrover15/resource-reaper/actions/workflows/ci.yml/badge.svg)](https://github.com/abhinavgrover15/resource-reaper/actions/workflows/ci.yml)
[![Helm Release](https://github.com/abhinavgrover15/resource-reaper/actions/workflows/helm-release.yml/badge.svg)](https://github.com/abhinavgrover15/resource-reaper/actions/workflows/helm-release.yml)
[![codecov](https://codecov.io/gh/abhinavgrover15/resource-reaper/branch/main/graph/badge.svg)](https://codecov.io/gh/abhinavgrover15/resource-reaper)
[![Go Report Card](https://goreportcard.com/badge/github.com/abhinavgrover15/resource-reaper)](https://goreportcard.com/report/github.com/abhinavgrover15/resource-reaper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Image Size](https://ghcr-badge.egpl.dev/abhinavgrover15/resource-reaper/size)](https://github.com/abhinavgrover15/resource-reaper/pkgs/container/resource-reaper)


Resource Reaper is a Kubernetes controller that automatically manages resource lifecycles using TTL (Time-To-Live) annotations. It helps maintain cluster hygiene by automatically cleaning up resources that have exceeded their specified lifetime.

## 🌟 Features

- **TTL-based Cleanup**: Automatically delete resources when they exceed their time-to-live
- **Flexible Annotations**: Use simple annotations to specify resource lifetimes
- **Multiple Resource Support**: Works with various Kubernetes resources (Pods, Deployments, etc.)
- **Leader Election**: Supports high availability deployments
- **Metrics**: Prometheus metrics for monitoring cleanup operations
- **Helm Installation**: Easy deployment using Helm chart
- **Resource Efficient**: Minimal resource footprint
- **Kubernetes Native**: Built using controller-runtime framework

## 🚀 Quick Start

Install using Helm:

```bash
helm repo add resource-reaper https://abhinavgrover15.github.io/resource-reaper
helm repo update
helm install resource-reaper resource-reaper/resource-reaper
```

Add TTL annotation to your resources:

```yaml
annotations:
  reaper.resource/ttl: "24h"  # Resource will be deleted after 24 hours
```

## Overview

Resource Reaper watches Kubernetes resources and automatically deletes them when their TTL expires. This is useful for:
- Cleaning up temporary resources
- Managing development/test environments
- Automatic cleanup of completed jobs
- Time-limited demo resources

## Supported Resources

- Core Resources: Pods, Services, ConfigMaps, Secrets
- Apps Resources: Deployments, StatefulSets, DaemonSets
- Batch Resources: Jobs, CronJobs

## Usage

Add the `resource-reaper/ttl` annotation to any supported resource:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  annotations:
    resource-reaper/ttl: "2h"  # Pod will be deleted after 2 hours
spec:
  containers:
  - name: nginx
    image: nginx:latest
```

Supported TTL formats:
- Hours: "1h", "2h"
- Minutes: "30m", "45m"
- Days: "2d", "7d"
- Immediate deletion: "0s"

## Installation

### Using kubectl

1. Apply the RBAC configuration:
```bash
kubectl apply -f https://raw.githubusercontent.com/abhinavgrover15/resource-reaper/main/config/rbac.yaml
```

2. Deploy the controller:
```bash
kubectl apply -f https://raw.githubusercontent.com/abhinavgrover15/resource-reaper/main/config/deployment.yaml
```

### Using Helm

1. Add the Helm repository:
```bash
helm repo add resource-reaper https://abhinavgrover15.github.io/resource-reaper
helm repo update
```

2. Install the chart:
```bash
helm install resource-reaper resource-reaper/resource-reaper
```

## Development

### Building from source

1. Clone the repository:
```bash
git clone https://github.com/abhinavgrover15/resource-reaper.git
cd resource-reaper
```

2. Build the container:
```bash
docker build -t resource-reaper:latest .
```

3. Deploy to Kubernetes:
```bash
kubectl apply -f config/
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

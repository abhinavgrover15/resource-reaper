# Resource Reaper

A Kubernetes controller that automatically deletes resources based on TTL (Time-To-Live) annotations.

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

1. Clone the repository:
```bash
git clone https://github.com/abhinav/resource-reaper.git
cd resource-reaper
```

2. Deploy to Kubernetes:
```bash
kubectl apply -f config/rbac.yaml
kubectl apply -f config/deployment.yaml
```

## Development

Build the controller:
```bash
make build
```

Run tests:
```bash
make test
```

## License

MIT License

# TTL Controller Examples

This directory contains example Kubernetes manifests that demonstrate how to use the TTL Controller with various resource types.

## Usage

The TTL Controller uses the `ttl-controller/ttl` annotation to specify when a resource should be deleted. The TTL value should be a valid Go duration string (e.g., "1h", "24h", "7d").

### Available Examples

1. **Pod (pod.yaml)**
   - Simple nginx pod with 1-hour TTL
   - ```bash
     kubectl apply -f pod.yaml
     ```

2. **Deployment (deployment.yaml)**
   - Nginx deployment with 24-hour TTL
   - ```bash
     kubectl apply -f deployment.yaml
     ```

3. **Service (service.yaml)**
   - ClusterIP service with 1-week TTL
   - ```bash
     kubectl apply -f service.yaml
     ```

4. **Job (job.yaml)**
   - Simple cleanup job with 2-hour TTL
   - ```bash
     kubectl apply -f job.yaml
     ```

5. **ConfigMap (configmap.yaml)**
   - Application config with 30-day TTL
   - ```bash
     kubectl apply -f configmap.yaml
     ```

## TTL Duration Format

The TTL value supports the following units:
- `ns` - nanoseconds
- `us` - microseconds
- `ms` - milliseconds
- `s` - seconds
- `m` - minutes
- `h` - hours

Examples:
- `0s` - Delete immediately
- `30m` - Delete after 30 minutes
- `24h` - Delete after 24 hours
- `168h` - Delete after 1 week
- `720h` - Delete after 30 days

## Notes

1. The TTL countdown starts from the resource's creation timestamp
2. Setting TTL to "0s" will delete the resource immediately
3. Invalid TTL values will be logged as errors and the resource won't be deleted
4. Removing the TTL annotation will prevent the resource from being deleted

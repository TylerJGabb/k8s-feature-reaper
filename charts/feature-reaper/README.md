# Feature Reaper Helm Chart

This chart deploys a Kubernetes `CronJob` that runs the feature reaper every hour. The job relies on an image that contains the `k8s-feature-reaper` binary.

## Installing the Chart

```
helm install feature-reaper ./feature-reaper
```

## Values

| Key | Description | Default |
|-----|-------------|---------|
| `image.repository` | Image repository | `ghcr.io/example/feature-reaper` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Kubernetes image pull policy | `IfNotPresent` |
| `schedule` | Cron schedule for the job | `"0 * * * *"` |
| `maxAge` | Duration before a namespace is eligible for deletion | `72h` |

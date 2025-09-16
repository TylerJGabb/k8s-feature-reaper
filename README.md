# k8s-feature-reaper

## TL;DR
- Single-shot Go utility that prunes feature preview namespaces in a Kubernetes cluster.
- Targets namespaces with the label `isFeature=true` and reads the `updatedAt` annotation formatted as `20060102150405` (UTC `YYYYMMDDHHMMSS`).
- Deletes namespaces whose `updatedAt` timestamp is older than the `--max-age` duration supplied on the command line.

## How It Works
The entrypoint (`main.go`) builds an in-cluster Kubernetes client configuration, falling back to the local kubeconfig when run outside the cluster. It invokes `reaper.ReapNamespaces`, which:

1. Lists namespaces with the label selector `isFeature=true`.
2. Parses each namespace's `updatedAt` annotation using the `20060102150405` layout.
3. Deletes namespaces whose last updated timestamp is older than the configured `maxAge` window, logging invalid or failed deletions but continuing with the rest.

The repository also contains a Helm chart (`charts/feature-reaper`) that schedules this binary as a CronJob so the cleanup can run on an interval inside a cluster.

## Quickstart
### Run locally against a cluster
1. Ensure you have Go 1.21+ installed and your `KUBECONFIG` points at the cluster you want to clean.
2. Run the reaper (adjust the max age to fit your policy):
   ```sh
   go run ./main.go --max-age=72h
   ```
   The process queries the cluster once, deletes any stale feature namespaces it finds, then exits.

### Develop and test
- Run fast feedback tests: `make unit-test`
- Build the container image: `make build-docker`
- (Optional) Load the image into a local kind cluster: `make kind-create` and `make load`

## Deploy with Helm
The chart under `charts/feature-reaper` deploys the binary as a CronJob. To install or upgrade into a cluster namespace:
```sh
make upgrade-install
```
Key values:
- `image`: Container image reference that includes the `k8s-feature-reaper` binary (defaults to `placeholder`).
- `schedule`: Cron expression for how often to run the job (defaults to hourly).
- `maxAge`: Duration passed to the binary's `--max-age` flag.

Adjust these in `charts/feature-reaper/values.yaml` or via `--set/--values` when installing.

## Related Scripts
- `integration-test.sh`: Spins up a kind cluster, deploys the chart, and verifies only stale namespaces are removed. It requires Docker, kind, helm, kubectl, and make, so it is best run locally rather than in automated CI environments.

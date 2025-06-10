A golang program that runs once and then exits. When it runs it looks at all namespaces in the cluster that its currently in with `labels.isFeature=true` and `annotations.updatedAt` (formatted in `20060102150405`) is more than 72 hours ago - then deletes them without waiting for confirmation.

Run unit tests with `make unit-test`.
Do not run `make integration-test` -- it requires too many dependencies to be run by an Agent.
A golang program that runs once and then exits. When it runs it looks at all namespaces in the cluster that its currently in with `labels.isFeature=true` and `annotations.updatedAt` (formatted in `20060102150405`) is more than 72 hours ago - then deletes them without waiting for confirmation.

You won't be able to run the tests unless you have docker, kubectl, and kind installed.

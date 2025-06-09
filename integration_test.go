package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestIntegrationKind(t *testing.T) {
	t.Logf("Checking for kind and kubectl...")
	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kind not installed")
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed")
	}
	t.Logf("✅ kind and kubectl found, proceeding with test...")

	t.Logf("🚀 Creating a kind cluster for testing...")
	clusterName := "k8s-feature-reaper-test"
	if out, err := exec.Command("kind", "create", "cluster", "--name", clusterName, "--wait", "60s").CombinedOutput(); err != nil {
		t.Fatalf("failed to create kind cluster: %v\n%s", err, out)
	}
	t.Logf("✅ Kind cluster created successfully.")
	defer exec.Command("kind", "delete", "cluster", "--name", clusterName).Run()

	oldTS := time.Now().Add(-73 * time.Hour).Format(timeLayout)
	newTS := time.Now().Add(-1 * time.Hour).Format(timeLayout)

	nsYAML := func(name, ts string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    isFeature: "true"
  annotations:
    updatedAt: "%s"
`, name, ts)
	}

	t.Logf("📂 Creating temporary YAML files for namespaces.")
	tmpDir, err := os.MkdirTemp("", "ns")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	oldFile := tmpDir + "/old.yaml"
	newFile := tmpDir + "/new.yaml"
	os.WriteFile(oldFile, []byte(nsYAML("ns-old", oldTS)), 0644)
	os.WriteFile(newFile, []byte(nsYAML("ns-new", newTS)), 0644)

	t.Logf("📄 creating namespaces...")
	if out, err := exec.Command("kubectl", "apply", "-f", oldFile).CombinedOutput(); err != nil {
		t.Fatalf("failed to create old namespace: %v\n%s", err, out)
	}
	if out, err := exec.Command("kubectl", "apply", "-f", newFile).CombinedOutput(); err != nil {
		t.Fatalf("failed to create new namespace: %v\n%s", err, out)
	}

	t.Logf("✅ Namespaces created successfully.")
	t.Logf("⏳ Running the reaper to clean up old namespaces...")
	if out, err := exec.Command("go", "run", ".", "--max-age=72h").CombinedOutput(); err != nil {
		t.Fatalf("failed to run reaper: %v\n%s", err, out)
	} else {
		t.Logf("reaper output:\n%s", out)
	}

	t.Logf("✅ Reaper executed successfully, checking namespaces...")
	out, err := exec.Command("kubectl", "get", "ns", "--field-selector=status.phase=Active", "-o", "jsonpath={.items[*].metadata.name}").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list namespaces: %v\n%s", err, out)
	}
	names := string(out)
	if !strings.Contains(names, "ns-new") {
		t.Fatalf("expected ns-new to exist; got %s", names)
	}
	if strings.Contains(names, "ns-old") {
		t.Fatalf("expected ns-old to be deleted; got %s", names)
	}
	t.Logf("🎉 Integration test completed successfully!")
}

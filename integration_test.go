package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestIntegrationKind(t *testing.T) {
	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kind not installed")
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed")
	}

	clusterName := "k8s-feature-reaper-test"
	if out, err := exec.Command("kind", "create", "cluster", "--name", clusterName, "--wait", "60s").CombinedOutput(); err != nil {
		t.Fatalf("failed to create kind cluster: %v\n%s", err, out)
	}
	defer exec.Command("kind", "delete", "cluster", "--name", clusterName).Run()

	oldTS := time.Now().Add(-73 * time.Hour).Format(timeLayout)
	newTS := time.Now().Add(-1 * time.Hour).Format(timeLayout)

	nsYAML := func(name, ts string) string {
		return "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: " + name + "\n  labels:\n    isFeature: \"true\"\n  annotations:\n    updatedAt: \"" + ts + "\"\n"
	}

	tmpDir, err := os.MkdirTemp("", "ns")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	oldFile := tmpDir + "/old.yaml"
	newFile := tmpDir + "/new.yaml"
	os.WriteFile(oldFile, []byte(nsYAML("ns-old", oldTS)), 0644)
	os.WriteFile(newFile, []byte(nsYAML("ns-new", newTS)), 0644)

	if out, err := exec.Command("kubectl", "apply", "-f", oldFile).CombinedOutput(); err != nil {
		t.Fatalf("failed to create old namespace: %v\n%s", err, out)
	}
	if out, err := exec.Command("kubectl", "apply", "-f", newFile).CombinedOutput(); err != nil {
		t.Fatalf("failed to create new namespace: %v\n%s", err, out)
	}

	if out, err := exec.Command("go", "run", ".").CombinedOutput(); err != nil {
		t.Fatalf("failed to run reaper: %v\n%s", err, out)
	} else {
		t.Logf("reaper output:\n%s", out)
	}

	out, err := exec.Command("kubectl", "get", "ns", "-o", "jsonpath={.items[*].metadata.name}").CombinedOutput()
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
}

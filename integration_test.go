package main

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReaperWithFakeClient(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()

	oldTS := time.Now().Add(-73 * time.Hour).Format(timeLayout)
	newTS := time.Now().Add(-1 * time.Hour).Format(timeLayout)

	oldNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-old",
			Labels: map[string]string{
				"isFeature": "true",
			},
			Annotations: map[string]string{
				updatedAtKey: oldTS,
			},
		},
	}
	newNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-new",
			Labels: map[string]string{
				"isFeature": "true",
			},
			Annotations: map[string]string{
				updatedAtKey: newTS,
			},
		},
	}

	if _, err := client.CoreV1().Namespaces().Create(ctx, oldNs, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to create old ns: %v", err)
	}
	if _, err := client.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to create new ns: %v", err)
	}

	if err := reapNamespaces(ctx, client, 72*time.Hour, time.Now()); err != nil {
		t.Fatalf("reapNamespaces returned error: %v", err)
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}
	foundOld, foundNew := false, false
	for _, ns := range nsList.Items {
		switch ns.Name {
		case "ns-old":
			foundOld = true
		case "ns-new":
			foundNew = true
		}
	}
	if foundOld {
		t.Fatalf("expected ns-old to be deleted")
	}
	if !foundNew {
		t.Fatalf("expected ns-new to exist")
	}
}

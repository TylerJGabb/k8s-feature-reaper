package reaper_test

import (
	"context"
	"testing"
	"time"

	"k8s-feature-reaper/reaper"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReapNamespaces(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()

	oldTS := time.Now().Add(-73 * time.Hour).Format(reaper.TIME_LAYOUT)
	newTS := time.Now().Add(-1 * time.Hour).Format(reaper.TIME_LAYOUT)

	oldNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-old",
			Labels: map[string]string{
				reaper.IS_FEATURE_KEY: "true",
			},
			Annotations: map[string]string{
				reaper.UPDATED_AT_KEY: oldTS,
			},
		},
	}
	newNs := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-new",
			Labels: map[string]string{
				reaper.IS_FEATURE_KEY: "true",
			},
			Annotations: map[string]string{
				reaper.UPDATED_AT_KEY: newTS,
			},
		},
	}

	if _, err := client.CoreV1().Namespaces().Create(ctx, oldNs, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to create old ns: %v", err)
	}
	if _, err := client.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to create new ns: %v", err)
	}

	if err := reaper.ReapNamespaces(ctx, client, 72*time.Hour, time.Now()); err != nil {
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

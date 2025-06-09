package main

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	labelSelector = "isFeature=true"
	updatedAtKey  = "updatedAt"
	timeLayout    = "20060102150405"
	maxAge        = 72 * time.Hour
)

func buildConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	// Fallback to kubeconfig from env
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	return kubeconfig.ClientConfig()
}

func main() {
	ctx := context.Background()
	config, err := buildConfig()
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Fatalf("failed to list namespaces: %v", err)
	}

	now := time.Now()
	for _, ns := range namespaces.Items {
		ann := ns.Annotations
		if ann == nil {
			continue
		}
		tsStr, ok := ann[updatedAtKey]
		if !ok {
			continue
		}
		ts, err := time.Parse(timeLayout, tsStr)
		if err != nil {
			log.Printf("namespace %s has invalid updatedAt: %v", ns.Name, err)
			continue
		}
		if now.Sub(ts) > maxAge {
			if err := clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); err != nil {
				log.Printf("failed to delete namespace %s: %v", ns.Name, err)
			} else {
				fmt.Printf("deleted namespace %s\n", ns.Name)
			}
		}
	}
}

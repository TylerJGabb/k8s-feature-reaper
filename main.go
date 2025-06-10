package main

import (
	"context"
	"flag"
	"k8s-feature-reaper/reaper"
	"log"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var maxAge = flag.Duration("max-age", 72*time.Hour, "maximum age of a namespace before deletion")

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
	flag.Parse()
	ctx := context.Background()
	config, err := buildConfig()
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	if err := reaper.ReapNamespaces(ctx, clientset, *maxAge, time.Now()); err != nil {
		log.Fatalf("failed to reap namespaces: %v", err)
	}
}

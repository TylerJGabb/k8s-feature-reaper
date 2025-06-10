package reaper

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	IS_FEATURE_KEY = "isFeature"
	LABEL_SELECTOR = IS_FEATURE_KEY + "=true"
	UPDATED_AT_KEY = "updatedAt"
	TIME_LAYOUT    = "20060102150405"
)

func ReapNamespaces(ctx context.Context, client kubernetes.Interface, maxAge time.Duration, now time.Time) error {
	listOpts := metav1.ListOptions{
		LabelSelector: LABEL_SELECTOR,
	}
	namespaces, err := client.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		return err
	}

	for _, ns := range namespaces.Items {
		ann := ns.Annotations
		if ann == nil {
			continue
		}
		tsStr, ok := ann[UPDATED_AT_KEY]
		if !ok {
			continue
		}
		ts, err := time.Parse(TIME_LAYOUT, tsStr)
		if err != nil {
			log.Printf("namespace %s has invalid updatedAt: %v", ns.Name, err)
			continue
		}
		if now.Sub(ts) > maxAge {
			if err := client.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); err != nil {
				log.Printf("failed to delete namespace %s: %v", ns.Name, err)
			} else {
				fmt.Printf("deleted namespace %s\n", ns.Name)
			}
		}
	}
	return nil
}

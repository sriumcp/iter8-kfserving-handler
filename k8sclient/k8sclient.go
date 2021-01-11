// Package k8sclient enables in-cluster interaction with Kubernetes API server.
package k8sclient

import (
	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// K8s interface enables getting a k8s client.
type K8s interface {
	GetClient() (client.Client, error)
}

// Iter8K8s is an implementation of the K8s interface.
type Iter8K8s struct{}

// GetClient constructs and returns an controller-runtime client.
func (k *Iter8K8s) GetClient() (client.Client, error) {
	crScheme := runtime.NewScheme()
	err := etc3.AddToScheme(crScheme)
	if err != nil {
		return nil, err
	}
	// This will use in-cluster configuration
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	rc, err := client.New(config, client.Options{
		Scheme: crScheme,
	})
	if err != nil {
		return nil, err
	}
	return rc, nil
}

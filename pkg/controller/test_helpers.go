package controller

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// getTestKubeconfig returns a kubeconfig for testing
func getTestKubeconfig() (*rest.Config, error) {
	// First try in-cluster config
	cfg, err := config.GetConfig()
	if err == nil {
		return cfg, nil
	}

	// Then try KUBECONFIG env var
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Finally try default kubeconfig path
	home, err := os.UserHomeDir()
	if err == nil {
		kubeconfig := filepath.Join(home, ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err == nil {
			return cfg, nil
		}
	}

	// If no kubeconfig is available, return a fake config for testing
	return &rest.Config{
		Host: "http://localhost:8080",
	}, nil
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// GetKubeConfig loads or creates a Kubernetes configuration and returns its path.
// It uses default loading rules and creates a new config if none exists.
func GetKubeConfig() (*api.Config, string, error) {
	const op = "kubeconfig.GetKubeConfig"

	// Initializing default loading rules
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfigPath := loadingRules.GetDefaultFilename()
	if kubeconfigPath == "" {
		return nil, "", fmt.Errorf("%s: could not determine default kubeconfig path", op)
	}

	// Checking if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		fmt.Printf("\033[33m[%s] Kubeconfig file not found at %s, creating new config\033[0m\n", op, kubeconfigPath)
		return api.NewConfig(), kubeconfigPath, nil
	} else if err != nil {
		return nil, "", fmt.Errorf("%s: failed to check kubeconfig file %s: %w", op, kubeconfigPath, err)
	}

	// Reading kubeconfig file
	configBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("%s: failed to read kubeconfig file %s: %w", op, kubeconfigPath, err)
	}

	// Parsing kubeconfig
	config, err := clientcmd.Load(configBytes)
	if err != nil {
		return nil, "", fmt.Errorf("%s: failed to parse kubeconfig file %s: %w", op, kubeconfigPath, err)
	}

	if config == nil {
		return nil, "", fmt.Errorf("%s: parsed kubeconfig is nil", op)
	}

	return config, kubeconfigPath, nil
}

// ValidateClusterAccess verifies connectivity to a Kubernetes cluster using the provided server and token.
// It attempts to list namespaces to confirm access.
func ValidateClusterAccess(server, token string) error {
	const op = "kubeconfig.ValidateClusterAccess"

	// Validating input parameters
	if server == "" {
		return fmt.Errorf("%s: server address cannot be empty", op)
	}
	if token == "" {
		return fmt.Errorf("%s: token cannot be empty", op)
	}

	// Configuring REST client
	restConfig := &rest.Config{
		Host:            server,
		BearerToken:     token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true}, // Note: Consider making TLS verification configurable
		Timeout:         5 * time.Second,
	}

	// Creating Kubernetes client
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("%s: failed to create Kubernetes client for %s: %w", op, server, err)
	}

	// Testing API access by listing namespaces
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("%s: failed to access API at %s: %w", op, server, err)
	}

	return nil
}

// CheckNameConflicts verifies that the provided name does not conflict with existing clusters, users, or contexts.
func CheckNameConflicts(config *api.Config, name string) error {
	const op = "kubeconfig.CheckNameConflicts"

	// Validating input parameters
	if config == nil {
		return fmt.Errorf("%s: config cannot be nil", op)
	}
	if name == "" {
		return fmt.Errorf("%s: name cannot be empty", op)
	}

	// Checking for conflicts
	if _, exists := config.Clusters[name]; exists {
		return fmt.Errorf("%s: cluster named %q already exists", op, name)
	}
	if _, exists := config.AuthInfos[name]; exists {
		return fmt.Errorf("%s: user named %q already exists", op, name)
	}
	if _, exists := config.Contexts[name]; exists {
		return fmt.Errorf("%s: context named %q already exists", op, name)
	}

	return nil
}

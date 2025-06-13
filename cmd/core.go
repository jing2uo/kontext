package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NewContext adds a new Kubernetes context with the specified name, server, and token.
// It performs name conflict checks and saves the configuration, returning any errors.
func NewContext(name, server, token string) error {
	const op = "kubeconfig.NewContext"

	// Validating input parameters
	if name == "" {
		return fmt.Errorf("%s: context name cannot be empty", op)
	}
	if server == "" {
		return fmt.Errorf("%s: server address cannot be empty", op)
	}
	if token == "" {
		return fmt.Errorf("%s: token cannot be empty", op)
	}

	// Loading kubeconfig
	config, kubeconfigPath, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("%s: failed to load kubeconfig: %w", op, err)
	}

	// Checking for name conflicts
	if err := CheckNameConflicts(config, name); err != nil {
		return fmt.Errorf("%s: name conflict check failed: %w", op, err)
	}

	// Building new configuration
	// Adding cluster
	cluster := api.NewCluster()
	cluster.Server = server
	cluster.InsecureSkipTLSVerify = true // Note: Consider making TLS verification configurable
	config.Clusters[name] = cluster

	// Adding user credentials
	authInfo := api.NewAuthInfo()
	authInfo.Token = token
	config.AuthInfos[name] = authInfo

	// Adding context
	ctx := api.NewContext()
	ctx.Cluster = name
	ctx.AuthInfo = name
	config.Contexts[name] = ctx

	// Setting current context
	config.CurrentContext = name

	// Saving updated configuration
	if err := SafeWriteConfig(config, kubeconfigPath); err != nil {
		return fmt.Errorf("%s: failed to save kubeconfig to %s: %w", op, kubeconfigPath, err)
	}

	return nil
}

// CleanContext cleans orphaned clusters and users from the kubeconfig.
// It returns the lists of removed clusters and users, along with any errors.
func CleanContext(config *api.Config) ([]string, []string, error) {

	// Analyzing resource references
	usedResources := struct {
		Clusters map[string]struct{}
		Users    map[string]struct{}
	}{make(map[string]struct{}), make(map[string]struct{})}

	// Collecting references from contexts
	for _, ctx := range config.Contexts {
		usedResources.Clusters[ctx.Cluster] = struct{}{}
		usedResources.Users[ctx.AuthInfo] = struct{}{}
	}

	// Identifying orphaned resources
	var resourcesToRemove struct {
		Clusters []string
		Users    []string
	}

	for name := range config.Clusters {
		if _, used := usedResources.Clusters[name]; !used {
			resourcesToRemove.Clusters = append(resourcesToRemove.Clusters, name)
		}
	}
	for name := range config.AuthInfos {
		if _, used := usedResources.Users[name]; !used {
			resourcesToRemove.Users = append(resourcesToRemove.Users, name)
		}
	}

	// Removing orphaned resources
	for _, name := range resourcesToRemove.Clusters {
		delete(config.Clusters, name)
	}
	for _, name := range resourcesToRemove.Users {
		delete(config.AuthInfos, name)
	}

	return resourcesToRemove.Clusters, resourcesToRemove.Users, nil
}

// SafeWriteConfig writes the kubeconfig to a file atomically using a temporary file.
func SafeWriteConfig(config *api.Config, path string) error {
	// Creating temporary file for atomic write
	tempFile, err := os.CreateTemp(filepath.Dir(path), "kubeconfig-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Serializing config to temporary file
	if err := clientcmd.WriteToFile(*config, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Atomically replacing original file
	if err := os.Rename(tempFile.Name(), path); err != nil {
		return fmt.Errorf("failed to replace file %s: %w", path, err)
	}
	return nil
}

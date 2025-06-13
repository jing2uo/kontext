package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Scan scans for sub-clusters based on the specified cluster type and returns a list of context configurations.
// It delegates to type-specific scan functions (e.g., ScanAlauda).
func Scan(name, server, token, clusterType string) ([]ContextConfig, error) {
	const op = "kubeconfig.Scan"

	// Validating input parameters
	if name == "" {
		return nil, fmt.Errorf("%s: context name cannot be empty", op)
	}
	if server == "" {
		return nil, fmt.Errorf("%s: server address cannot be empty", op)
	}
	if token == "" {
		return nil, fmt.Errorf("%s: token cannot be empty", op)
	}

	// Dispatching to type-specific scan function
	switch clusterType {
	case "alauda":
		return ScanAlauda(name, server, token)
	default:
		fmt.Printf("\033[33m[%s] Skipped: unsupported clusterType %q\033[0m\n", op, clusterType)
		return nil, nil
	}
}

// ScanAlauda scans for clusters.platform.tkestack.io resources, constructs new context names,
// and modifies the server URL by replacing the last path segment with the cluster name.
func ScanAlauda(name, server, token string) ([]ContextConfig, error) {
	const op = "kubeconfig.ScanAlauda"

	// Configuring REST client for API access
	restConfig := &rest.Config{
		Host:            server,
		BearerToken:     token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true}, // Note: Consider making TLS verification configurable
		Timeout:         5 * time.Second,
	}

	// Creating Kubernetes client
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create Kubernetes client for %s: %w", op, server, err)
	}

	// Checking for clusters.platform.tkestack.io resources
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	apiResources, err := clientset.Discovery().ServerResourcesForGroupVersion("platform.tkestack.io/v1")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to discover platform.tkestack.io/v1 resources: %w", op, err)
	}

	hasClusterResource := false
	for _, resource := range apiResources.APIResources {
		if resource.Name == "clusters" {
			hasClusterResource = true
			break
		}
	}

	if !hasClusterResource {
		fmt.Printf("\033[33m[%s] No clusters.platform.tkestack.io resources found\033[0m\n", op)
		return nil, nil
	}

	// Retrieving cluster resources
	resp, err := clientset.RESTClient().Get().
		AbsPath("/apis/platform.tkestack.io/v1/clusters").
		DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list clusters.platform.tkestack.io: %w", op, err)
	}

	// Parsing cluster list response
	var clusterList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp, &clusterList); err != nil {
		return nil, fmt.Errorf("%s: failed to parse clusters.platform.tkestack.io response: %w", op, err)
	}

	// Constructing context configurations
	var configs []ContextConfig
	for _, item := range clusterList.Items {
		clusterName := item.Metadata.Name
		if strings.ToLower(clusterName) == "global" {
			continue
		}

		// Generating new context name
		newContextName := fmt.Sprintf("%s-%s", name, clusterName)

		// Extracting protocol and base path
		protocol := "https://"
		serverPath := server
		if strings.HasPrefix(server, "https://") {
			serverPath = strings.TrimPrefix(server, "https://")
		} else if strings.HasPrefix(server, "http://") {
			protocol = "http://"
			serverPath = strings.TrimPrefix(server, "http://")
		}

		// Constructing new server URL by replacing the last path segment
		newServerPath := path.Join(path.Dir(serverPath), clusterName)
		newServer := protocol + strings.TrimLeft(newServerPath, "/")

		configs = append(configs, ContextConfig{
			Name:   newContextName,
			Server: newServer,
			Token:  token,
		})
	}

	if len(configs) == 0 {
		return nil, nil
	}

	return configs, nil

}

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

// ContextConfig represents a context configuration for merging or scanning
type ContextConfig struct {
	Name   string
	Server string
	Token  string
}

// MergeContext handles the merge command, merging contexts from an external kubeconfig file.
// It validates inputs, scans for sub-clusters if requested, and manages program output.
func MergeContext(filePath, namePrefix string, scan *string) error {
	const op = "kubeconfig.MergeContext"

	// Validating input
	if filePath == "" {
		return fmt.Errorf("%s: kubeconfig file path cannot be empty", op)
	}

	// Loading current kubeconfig
	currentConfig, _, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("%s: failed to load current kubeconfig: %w", op, err)
	}

	// Loading external kubeconfig
	externalConfig, err := clientcmd.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("%s: failed to parse external kubeconfig %s: %w", op, filePath, err)
	}

	// Collecting valid contexts
	var configs []ContextConfig
	var certificateContexts []string

	for ctxName, ctx := range externalConfig.Contexts {
		// Validating associated cluster and user
		cluster, cExists := externalConfig.Clusters[ctx.Cluster]
		authInfo, aExists := externalConfig.AuthInfos[ctx.AuthInfo]
		if !cExists || !aExists {
			fmt.Printf("\033[33m[%s] Skipped context %s: missing resources (cluster: %t, user: %t)\033[0m\n",
				op, ctxName, cExists, aExists)
			continue
		}

		// Skipping certificate-based authentication
		if authInfo.ClientCertificateData != nil || authInfo.ClientCertificate != "" {
			certificateContexts = append(certificateContexts, ctxName)
			continue
		}

		// Generating final context name
		finalCtxName := ctxName
		if namePrefix != "" {
			finalCtxName = fmt.Sprintf("%s-%s", namePrefix, ctxName)
		} else {
			filename := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
			finalCtxName = fmt.Sprintf("%s-%s", filename, ctxName)
		}

		configs = append(configs, ContextConfig{
			Name:   finalCtxName,
			Server: cluster.Server,
			Token:  authInfo.Token,
		})
	}

	// Reporting skipped certificate-based contexts
	if len(certificateContexts) > 0 {
		fmt.Printf("\033[33m[%s] Skipped %d certificate-based contexts: %v\033[0m\n",
			op, len(certificateContexts), certificateContexts)
	}

	// Checking for name conflicts
	for _, cfg := range configs {
		if _, exists := currentConfig.Contexts[cfg.Name]; exists {
			return fmt.Errorf("%s: name conflict detected for context %q; use --name to specify a prefix (e.g., --name=prod)", op, cfg.Name)
		}
	}

	// Processing contexts with optional scanning
	fmt.Printf("\033[36m[%s] Merging contexts...\033[0m\n", op)
	var allConfigs []ContextConfig
	successCount := 0
	for _, cfg := range configs {
		// Adding the primary context
		var contexts []ContextConfig
		contexts = append(contexts, cfg)

		// Scanning for sub-clusters if requested
		if scan != nil {
			scannedContexts, err := Scan(cfg.Name, cfg.Server, cfg.Token, *scan)
			if err != nil {
				fmt.Printf("\033[31m  ✗ Failed to scan sub-clusters for %s: %v\033[0m\n", cfg.Name, err)
				continue
			}
			if len(scannedContexts) == 0 {
				fmt.Printf("\033[33m[%s] No sub-clusters found for %s with type %q\033[0m\n", op, cfg.Name, *scan)
			} else {
				contexts = append(contexts, scannedContexts...)
			}
		}

		// Adding all contexts
		for _, ctx := range contexts {
			if err := NewContext(ctx.Name, ctx.Server, ctx.Token); err != nil {
				fmt.Printf("\033[31m  ✗ Failed to add context %s: %v\033[0m\n", ctx.Name, err)
				continue
			}
			fmt.Printf("\033[32m  ✓ Added context: %s (%s)\033[0m\n", ctx.Name, ctx.Server)
			successCount++
			allConfigs = append(allConfigs, ctx)
		}
	}

	// Displaying summary
	fmt.Printf("\033[36m\n%s Summary:\n", op)
	fmt.Printf("  ✓ Added contexts: %d\n", successCount)
	fmt.Printf("  ✗ Skipped certificate-based contexts: %d\n", len(certificateContexts))
	fmt.Printf("  ✗ Failed contexts: %d\n", len(allConfigs)-successCount)
	fmt.Println(strings.Repeat("=", 50) + "\033[0m")

	return nil
}
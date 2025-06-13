package cmd

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

// CleanContextCmd handles the clean command, validating and removing invalid contexts and orphaned resources.
// It manages program output for the operation.
func CleanContextCmd() error {
	const op = "kubeconfig.CleanContextCmd"

	// Loading kubeconfig file
	config, kubeconfigPath, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("%s: failed to load kubeconfig: %w", op, err)
	}

	// Phase 1: Identifying invalid contexts
	var contextsToRemove []string
	for ctxName, ctx := range config.Contexts {
		// Validating cluster and user references
		if _, ok := config.Clusters[ctx.Cluster]; !ok {
			contextsToRemove = append(contextsToRemove, ctxName)
			continue
		}
		if _, ok := config.AuthInfos[ctx.AuthInfo]; !ok {
			contextsToRemove = append(contextsToRemove, ctxName)
			continue
		}

		// Validating cluster connectivity
		if err := ValidateClusterAccess(config.Clusters[ctx.Cluster].Server, config.AuthInfos[ctx.AuthInfo].Token); err != nil {
			contextsToRemove = append(contextsToRemove, ctxName)
		}
	}

	// Phase 2: Checking for changes
	currentModified := false
	if config.CurrentContext != "" {
		if _, exists := config.Contexts[config.CurrentContext]; !exists || contains(contextsToRemove, config.CurrentContext) {
			currentModified = true
		}
	}

	// Phase 3: Creating backup if changes will occur
	var backupPath string
	if len(contextsToRemove) > 0 || currentModified {
		backupPath, err = BackupKubeConfig(config, kubeconfigPath)
		if err != nil {
			return fmt.Errorf("%s: failed to create backup: %w", op, err)
		}
	}

	// Phase 4: Removing invalid contexts
	fmt.Printf("\033[36m[%s] Cleaning contexts...\033[0m\n", op)
	for _, ctxName := range contextsToRemove {
		delete(config.Contexts, ctxName)
		fmt.Printf("\033[31m  ✓ Removed invalid context: %s\033[0m\n", ctxName)
	}

	// Updating current context if necessary
	if currentModified {
		config.CurrentContext = ""
		fmt.Printf("\033[31m  ✓ Cleared current context setting\033[0m\n")
	}

	// Phase 5: Cleaning orphaned resources
	removedClusters, removedUsers, err := CleanContext(config)
	if err != nil {
		return fmt.Errorf("%s: failed to clean orphaned resources: %w", op, err)
	}

	// Reporting cleaned resources
	for _, cluster := range removedClusters {
		fmt.Printf("\033[33m  ✓ Removed orphaned cluster: %s\033[0m\n", cluster)
	}
	for _, user := range removedUsers {
		fmt.Printf("\033[33m  ✓ Removed orphaned user: %s\033[0m\n", user)
	}

	// Phase 6: Saving configuration
	if len(contextsToRemove) > 0 || len(removedClusters) > 0 || len(removedUsers) > 0 || currentModified {
		if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
			return fmt.Errorf("%s: failed to write updated kubeconfig to %s: %w", op, kubeconfigPath, err)
		}
	} else {
		fmt.Printf("\033[32m[%s] No invalid or orphaned resources found. Kubeconfig is healthy.\033[0m\n", op)
	}

	// Phase 7: Displaying summary
	fmt.Printf("\033[36m\n%s Summary:\n", op)
	fmt.Printf("  ✓ Removed contexts: %d\n", len(contextsToRemove))
	if currentModified {
		fmt.Printf("  ✓ Current context reset\n")
	}
	fmt.Printf("  ✓ Removed clusters: %d\n", len(removedClusters))
	fmt.Printf("  ✓ Removed users: %d\n", len(removedUsers))
	if backupPath != "" {
		fmt.Printf("  ✓ Backup saved at: %s\n", backupPath)
	}
	fmt.Println(strings.Repeat("=", 50) + "\033[0m")

	return nil
}

// contains checks if a string slice contains a specific value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

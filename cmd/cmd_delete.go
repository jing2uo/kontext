package cmd

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

// DeleteContext removes Kubernetes contexts matching the given name or wildcard pattern.
// It cleans up orphaned resources using CleanContext and manages program output.
func DeleteContext(namePattern string) error {
	const op = "kubeconfig.DeleteContext"

	// Validating input pattern
	if namePattern == "" {
		return fmt.Errorf("%s: context name or pattern cannot be empty", op)
	}

	// Loading kubeconfig file
	config, kubeconfigPath, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("%s: failed to load kubeconfig: %w", op, err)
	}

	// Determining if the pattern is a wildcard
	isWildcard := strings.Contains(namePattern, "*")
	var matchedContexts []string

	// Matching contexts based on pattern
	if isWildcard {
		prefix := strings.ReplaceAll(namePattern, "*", "")
		if prefix == "" {
			return fmt.Errorf("%s: wildcard pattern must include a prefix", op)
		}
		for ctxName := range config.Contexts {
			if strings.HasPrefix(ctxName, prefix) {
				matchedContexts = append(matchedContexts, ctxName)
			}
		}
		if len(matchedContexts) == 0 {
			return fmt.Errorf("%s: no contexts found matching pattern %q", op, namePattern)
		}
	} else {
		if _, exists := config.Contexts[namePattern]; !exists {
			return fmt.Errorf("%s: context %q does not exist", op, namePattern)
		}
		matchedContexts = []string{namePattern}
	}

	// Checking if changes are needed
	currentModified := false
	for _, ctxName := range matchedContexts {
		if config.CurrentContext == ctxName {
			currentModified = true
		}
	}

	// Creating backup if changes will occur
	var backupPath string
	if len(matchedContexts) > 0 || currentModified {
		backupPath, err = BackupKubeConfig(config, kubeconfigPath)
		if err != nil {
			return fmt.Errorf("%s: failed to create backup: %w", op, err)
		}
	}

	// Deleting matched contexts
	fmt.Printf("\033[36m[%s] Deleting contexts...\033[0m\n", op)
	for _, ctxName := range matchedContexts {
		delete(config.Contexts, ctxName)
		fmt.Printf("\033[31m  ✓ Removed context: %s\033[0m\n", ctxName)

		// Updating current context if necessary
		if config.CurrentContext == ctxName {
			config.CurrentContext = ""
			fmt.Printf("\033[31m  ✓ Cleared current context setting\033[0m\n")
		}
	}

	// Cleaning up orphaned resources
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

	// Saving updated configuration
	if len(matchedContexts) > 0 || len(removedClusters) > 0 || len(removedUsers) > 0 || currentModified {
		if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
			return fmt.Errorf("%s: failed to save kubeconfig to %s: %w", op, kubeconfigPath, err)
		}
	}

	// Displaying summary
	fmt.Printf("\033[36m\n%s Summary:\n", op)
	fmt.Printf("  ✓ Removed contexts: %d\n", len(matchedContexts))
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
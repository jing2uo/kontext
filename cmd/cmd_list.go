package cmd

import (
	"fmt"
	"sort"
)

// ListContexts displays all Kubernetes contexts and identifies orphaned resources.
// It lists active contexts with their associated clusters and users, and reports any unused clusters or users.
func ListContexts() error {
	const op = "kubeconfig.ListContexts"

	// Loading kubeconfig file
	config, _, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("%s: failed to load kubeconfig: %w", op, err)
	}

	// Tracking referenced resources
	usedResources := struct {
		Clusters map[string]bool
		Users    map[string]bool
	}{make(map[string]bool), make(map[string]bool)}

	// Collecting references from contexts
	for _, ctx := range config.Contexts {
		usedResources.Clusters[ctx.Cluster] = true
		usedResources.Users[ctx.AuthInfo] = true
	}

	// Displaying active contexts
	fmt.Printf("\033[36m\n===== Active Contexts (%d) =====\033[0m\n", len(config.Contexts))
	if len(config.Contexts) == 0 {
		fmt.Println("\033[33mNo contexts found.\033[0m")
	} else {
		// Sorting context names for consistent output
		var contextNames []string
		for ctxName := range config.Contexts {
			contextNames = append(contextNames, ctxName)
		}
		sort.Strings(contextNames)

		for _, ctxName := range contextNames {
			ctx := config.Contexts[ctxName]
			fmt.Printf("\n\033[32m● %s\033[0m\n", ctxName)
			fmt.Printf("  ├─ \033[33mCluster:\033[0m %s", ctx.Cluster)
			if cluster, ok := config.Clusters[ctx.Cluster]; ok {
				fmt.Printf(" (%s)", cluster.Server)
			} else {
				fmt.Printf(" \033[31m(missing)\033[0m")
			}
			fmt.Printf("\n  └─ \033[33mUser:\033[0m %s", ctx.AuthInfo)
			if _, ok := config.AuthInfos[ctx.AuthInfo]; !ok {
				fmt.Printf(" \033[31m(missing)\033[0m")
			}
			fmt.Println()
		}
	}

	// Identifying and displaying orphaned resources
	var orphanClusters, orphanUsers []string
	for name := range config.Clusters {
		if !usedResources.Clusters[name] {
			orphanClusters = append(orphanClusters, name)
		}
	}
	for name := range config.AuthInfos {
		if !usedResources.Users[name] {
			orphanUsers = append(orphanUsers, name)
		}
	}

	// Sorting orphaned resources for consistent output
	sort.Strings(orphanClusters)
	sort.Strings(orphanUsers)

	hasOrphans := len(orphanClusters) > 0 || len(orphanUsers) > 0
	if hasOrphans {
		fmt.Printf("\n\033[36m===== Orphaned Resources =====\033[0m\n")
		if len(orphanClusters) > 0 {
			fmt.Printf("\n\033[31mUnused Clusters (%d):\033[0m\n", len(orphanClusters))
			for _, name := range orphanClusters {
				if cluster, ok := config.Clusters[name]; ok {
					fmt.Printf("  × %s (%s)\n", name, cluster.Server)
				} else {
					fmt.Printf("  × %s \033[31m(corrupted)\033[0m\n", name)
				}
			}
		}
		if len(orphanUsers) > 0 {
			fmt.Printf("\n\033[31mUnused Users (%d):\033[0m\n", len(orphanUsers))
			for _, name := range orphanUsers {
				fmt.Printf("  × %s\n", name)
			}
		}
	} else {
		fmt.Printf("\n\033[36m===== No Orphaned Resources =====\033[0m\n")
		fmt.Println("\033[32mAll resources are properly referenced.\033[0m")
	}

	return nil
}

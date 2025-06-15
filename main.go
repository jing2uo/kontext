package main

import (
	"fmt"
	"os"

	"kontext/cmd"

	"github.com/spf13/cobra"
)

var (
	name   string
	server string
	token  string
	path   string
	scan   string
)

// Add an empty string to allow omitting the scan parameter
var validScans = []string{"alauda", ""}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kontext",
		Short: "Manage Kubernetes contexts efficiently",
		Long: `Kontext is a CLI tool for managing Kubernetes contexts in your kubectl configuration.
It provides commands to add, list, merge, delete, and clean Kubernetes contexts.`,
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new Kubernetes context",
		Long:  `Add a new Kubernetes context to the kubectl configuration using the provided name, server address, and authentication token.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("add command does not accept arguments, received: %v", args)
			}
			if err := validateName(name); err != nil {
				return fmt.Errorf("invalid name: %w", err)
			}
			if err := validateServer(server); err != nil {
				return fmt.Errorf("invalid server: %w", err)
			}
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}
			if err := validateScan(scan); err != nil {
				return fmt.Errorf("invalid scan value: %w", err)
			}
			var scanPtr *string
			if scan != "" {
				scanPtr = &scan
			}
			if err := cmd.AddContext(name, server, token, scanPtr); err != nil {
				return fmt.Errorf("failed to add context: %w", err)
			}
			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all Kubernetes contexts",
		Long:  `Displays all Kubernetes contexts currently configured in the kubectl configuration file.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("list command does not accept arguments, received: %v", args)
			}
			if err := cmd.ListContexts(); err != nil {
				return fmt.Errorf("failed to list contexts: %w", err)
			}
			return nil
		},
	}

	var mergeCmd = &cobra.Command{
		Use:   "merge",
		Short: "Merge a kubeconfig file",
		Long:  `Merges a specified kubeconfig YAML file into the existing kubectl configuration.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("merge command does not accept arguments, received: %v", args)
			}
			if path == "" {
				return fmt.Errorf("path to kubeconfig file is required")
			}
			if err := validateScan(scan); err != nil {
				return fmt.Errorf("invalid scan value: %w", err)
			}
			var scanPtr *string
			if scan != "" {
				scanPtr = &scan
			}
			if err := cmd.MergeContext(path, name, scanPtr); err != nil {
				return fmt.Errorf("failed to merge kubeconfig: %w", err)
			}
			return nil
		},
	}

	var cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Clean invalid Kubernetes contexts",
		Long:  `Validates and removes invalid or unreachable contexts from the kubectl configuration.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("clean command does not accept arguments, received: %v", args)
			}
			if err := cmd.CleanContextCmd(); err != nil {
				return fmt.Errorf("failed to clean contexts: %w", err)
			}
			return nil
		},
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete Kubernetes contexts",
		Long:  `Deletes one or more Kubernetes contexts from the kubectl configuration, supporting wildcard name patterns (e.g., name*).`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("delete command does not accept arguments, received: %v", args)
			}
			if err := validateName(name); err != nil {
				return fmt.Errorf("invalid name: %w", err)
			}
			if err := cmd.DeleteContext(name); err != nil {
				return fmt.Errorf("failed to delete context %q: %w", name, err)
			}
			return nil
		},
	}

	// Flag definitions
	addCmd.Flags().StringVar(&name, "name", "", "Name for the context, cluster, and user (required)")
	addCmd.Flags().StringVar(&server, "server", "", "Kubernetes API server address (required)")
	addCmd.Flags().StringVar(&token, "token", "", "Kubernetes authentication token (required)")
	addCmd.Flags().StringVar(&scan, "scan", "", "Cluster type to scan for sub-clusters (e.g., alauda)")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("server")
	addCmd.MarkFlagRequired("token")

	addCmd.RegisterFlagCompletionFunc("scan", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validScans, cobra.ShellCompDirectiveNoFileComp
	})

	mergeCmd.Flags().StringVar(&name, "name", "", "Optional name prefix for the context, cluster, and user")
	mergeCmd.Flags().StringVar(&path, "path", "", "Path to the kubeconfig file (required)")
	mergeCmd.Flags().StringVar(&scan, "scan", "", "Cluster type to scan for sub-clusters (e.g., alauda)")
	mergeCmd.MarkFlagRequired("path")
	mergeCmd.RegisterFlagCompletionFunc("scan", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validScans, cobra.ShellCompDirectiveNoFileComp
	})

	deleteCmd.Flags().StringVar(&name, "name", "", "Name of the context to delete (supports wildcard patterns, required)")
	deleteCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(addCmd, mergeCmd, deleteCmd, cleanCmd, listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// validateName ensures the context name is valid
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("context name too long (max 255 characters)")
	}
	return nil
}

// validateServer ensures the server address is valid
func validateServer(server string) error {
	if server == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	return nil
}

// validateScan ensures the scan type in validScans
func validateScan(scan string) error {
	for _, valid := range validScans {
		if scan == valid {
			return nil
		}
	}
	return fmt.Errorf("scan must be one of: %v", validScans)
}

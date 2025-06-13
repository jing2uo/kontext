package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kontext/cmd"
)

var (
	name  string
	host  string
	token string
	path  string
	scan  string
)

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
		Long: `Adds a new Kubernetes context to the kubectl configuration using the provided name, server host, and authentication token.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("add command does not accept arguments, received: %v", args)
			}
			if err := validateName(name); err != nil {
				return fmt.Errorf("invalid name: %w", err)
			}
			if err := validateHost(host); err != nil {
				return fmt.Errorf("invalid host: %w", err)
			}
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}
			var scanPtr *string
			if scan != "" {
				scanPtr = &scan
			}
			if err := cmd.AddContext(name, host, token, scanPtr); err != nil {
				return fmt.Errorf("failed to add context: %w", err)
			}
			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all Kubernetes contexts",
		Long: `Displays all Kubernetes contexts currently configured in the kubectl configuration file.`,
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
		Long: `Merges a specified kubeconfig YAML file into the existing kubectl configuration.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("merge command does not accept arguments, received: %v", args)
			}
			if path == "" {
				return fmt.Errorf("path to kubeconfig file is required")
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
		Long: `Validates and removes invalid or unreachable contexts from the kubectl configuration.`,
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
		Long: `Deletes one or more Kubernetes contexts from the kubectl configuration, supporting wildcard name patterns (e.g., name*).`,
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
	addCmd.Flags().StringVar(&host, "host", "", "Kubernetes API server address (required)")
	addCmd.Flags().StringVar(&token, "token", "", "Kubernetes authentication token (required)")
	addCmd.Flags().StringVar(&scan, "scan", "", "Cluster type to scan for sub-clusters (e.g., alauda)")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("host")
	addCmd.MarkFlagRequired("token")

	mergeCmd.Flags().StringVar(&name, "name", "", "Optional name prefix for the context, cluster, and user")
	mergeCmd.Flags().StringVar(&path, "path", "", "Path to the kubeconfig file (required)")
	mergeCmd.Flags().StringVar(&scan, "scan", "", "Cluster type to scan for sub-clusters (e.g., alauda)")
	mergeCmd.MarkFlagRequired("path")

	deleteCmd.Flags().StringVar(&name, "name", "", "Name of the context to delete (supports wildcard patterns, required)")
	deleteCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(addCmd, mergeCmd, deleteCmd, cleanCmd, listCmd)

	// Execute command and handle errors
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

// validateHost ensures the host address is valid
func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host address cannot be empty")
	}
	// Add more specific host validation if needed (e.g., URL format)
	return nil
}
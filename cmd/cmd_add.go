package cmd

import (
	"fmt"
	"strings"
)

// AddContext handles the add command, validating and adding contexts with optional sub-cluster scanning.
// It manages program output for the operation.
func AddContext(name, server, token string, scan *string) error {
	const op = "kubeconfig.AddContext"

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

	// Verifying cluster connectivity
	if err := ValidateClusterAccess(server, token); err != nil {
		return fmt.Errorf("%s: cluster validation failed [server=%s]: %w", op, server, err)
	}

	// Adding the primary context to the list
	var contexts []ContextConfig
	contexts = append(contexts, ContextConfig{
		Name:   name,
		Server: server,
		Token:  token,
	})

	// Scanning for sub-clusters if requested
	if scan != nil {
		scannedContexts, err := Scan(name, server, token, *scan)
		if err != nil {
			return fmt.Errorf("%s: failed to scan sub-clusters for type %q: %w", op, *scan, err)
		}
		if len(scannedContexts) == 0 {
			fmt.Printf("\033[33m[%s] No sub-clusters found for type %q\033[0m\n", op, *scan)
		} else {
			contexts = append(contexts, scannedContexts...)
		}
	}

	// Adding all contexts
	fmt.Printf("\033[36m[%s] Adding contexts...\033[0m\n", op)
	successCount := 0
	for _, ctx := range contexts {
		if err := NewContext(ctx.Name, ctx.Server, ctx.Token); err != nil {
			fmt.Printf("\033[31m  ✗ Failed to add context %s: %v\033[0m\n", ctx.Name, err)
			continue
		}
		fmt.Printf("\033[32m  ✓ Added context: %s (%s)\033[0m\n", ctx.Name, ctx.Server)
		successCount++
	}

	// Displaying summary
	fmt.Printf("\033[36m\n%s Summary:\n", op)
	fmt.Printf("  ✓ Added contexts: %d\n", successCount)
	fmt.Printf("  ✗ Failed contexts: %d\n", len(contexts)-successCount)
	fmt.Println(strings.Repeat("=", 50) + "\033[0m")

	return nil
}

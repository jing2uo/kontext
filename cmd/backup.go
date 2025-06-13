package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// BackupKubeConfig creates a backup of the provided kubeconfig configuration.
// It saves the backup under ~/.kube with a timestamp and retains up to 5 historical backups.
func BackupKubeConfig(config *api.Config, kubeconfigPath string) (string, error) {
	const op = "kubeconfig.BackupKubeConfig"

	// Validating inputs
	if config == nil {
		return "", fmt.Errorf("%s: config cannot be nil", op)
	}
	if kubeconfigPath == "" {
		return "", fmt.Errorf("%s: kubeconfig path cannot be empty", op)
	}

	// Serializing config
	configBytes, err := clientcmd.Write(*config)
	if err != nil {
		return "", fmt.Errorf("%s: failed to serialize kubeconfig: %w", op, err)
	}

	// Creating backup directory if it doesn't exist
	backupDir := filepath.Dir(kubeconfigPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("%s: failed to create backup directory %s: %w", op, backupDir, err)
	}

	// Generating backup file path
	backupPath := fmt.Sprintf("%s.backup-%s", kubeconfigPath, time.Now().Format("20060102-150405"))
	if err := os.WriteFile(backupPath, configBytes, 0600); err != nil {
		return "", fmt.Errorf("%s: failed to write backup file %s: %w", op, backupPath, err)
	}
	fmt.Printf("\033[32m[%s] Created backup: %s\033[0m\n", op, backupPath)

	// Managing historical backups (retain up to 5)
	if err := cleanupOldBackups(backupDir, kubeconfigPath); err != nil {
		fmt.Printf("\033[33m[%s] Warning: failed to clean old backups: %v\033[0m\n", op, err)
		// Non-fatal error, continue
	}

	return backupPath, nil
}

// cleanupOldBackups removes old backup files to retain only the 5 most recent.
func cleanupOldBackups(backupDir, kubeconfigPath string) error {
	const maxBackups = 5

	// Listing files in backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory %s: %w", backupDir, err)
	}

	// Filtering backup files
	var backups []string
	backupPrefix := filepath.Base(kubeconfigPath) + ".backup-"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), backupPrefix) {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	// Sorting backups by name (timestamp ensures chronological order)
	sort.Strings(backups)

	// Removing oldest backups if limit exceeded
	if len(backups) > maxBackups {
		for _, oldBackup := range backups[:len(backups)-maxBackups] {
			if err := os.Remove(oldBackup); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", oldBackup, err)
			}
			fmt.Printf("\033[33m[%s] Removed old backup: %s\033[0m\n", "kubeconfig.BackupKubeConfig", oldBackup)
		}
	}

	return nil
}
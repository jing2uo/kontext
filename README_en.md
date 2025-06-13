# Kontext CLI

Kontext is a command-line tool for efficiently managing Kubernetes contexts. It enables users to add, list, merge, delete, and clean Kubernetes contexts in the kubectl configuration file (typically `~/.kube/config`). Kontext supports sub-cluster scanning (e.g., Alauda cluster types) and wildcard operations, simplifying management in multi-cluster environments.

## Features

- **Add Contexts**: Quickly add new Kubernetes contexts using a name, server address, and token, with optional sub-cluster scanning.
- **Merge Configs**: Merge external kubeconfig files into the current configuration, supporting custom name prefixes and sub-cluster scanning.
- **List Contexts**: Display all current Kubernetes contexts and check for orphaned resources.
- **Delete Contexts**: Delete contexts by exact name or wildcard pattern, automatically cleaning up orphaned clusters and users.
- **Clean Contexts**: Validate and remove invalid or unreachable contexts, along with orphaned cluster and user resources.
- **Backup Management**: Automatically create backups before modifying the configuration (retains up to 5 historical backups in `~/.kube`).
- **Consistent Output**: Clear command-line output with operation progress, success/failure details, and summaries.

## Installation

Kontext is currently a development tool, recommended to be built from source. Follow these steps to install:

1. **Clone the Repository** (assuming hosted on a Git repository):

   ```bash
   git clone https://github.com/your-repo/kontext.git
   cd kontext
   ```

2. **Build and Install** (requires Go environment):

   ```bash
   go build -o kontext
   sudo mv kontext /usr/local/bin/
   ```

3. **Verify Installation**:
   ```bash
   kontext -h
   ```

## Usage Examples

The following examples operate on the `~/.kube/config` file, demonstrating Kontext's core functionality.

### 1. Add a Context

Add a context named `myenv` and scan for Alauda sub-clusters:

```bash
kontext add --name myenv --host "https://192.168.138.58/kubernetes/global" --token "<your-token>" --scan alauda
```

Output:

```
[kubeconfig.AddContext] Adding contexts...
  ✓ Added context: myenv (https://192.168.138.58/kubernetes/global)
  ✓ Added context: myenv-business-1 (https://192.168.138.58/kubernetes/business-1)

kubeconfig.AddContext Summary:
  ✓ Added contexts: 2
  ✗ Failed contexts: 0
==================================================
```

### 2. Merge a Kubeconfig File

Merge an external kubeconfig file with sub-cluster scanning:

```bash
kontext merge --path Downloads/asdf.yaml --scan alauda
```

Output:

```
[kubeconfig.MergeContext] Merging contexts...
  ✓ Added context: asdf-global (https://192.168.138.58/kubernetes/global)
  ✓ Added context: asdf-global-business-1 (https://192.168.138.58/kubernetes/business-1)

kubeconfig.MergeContext Summary:
  ✓ Added contexts: 2
  ✗ Skipped certificate-based contexts: 0
  ✗ Failed contexts: 0
==================================================
```

### 3. List Contexts

View all current contexts:

```bash
kontext list
```

Output:

```
===== Active Contexts (4) =====

● asdf-global
  ├─ Cluster: asdf-global (https://192.168.138.58/kubernetes/global)
  └─ User: asdf-global

● asdf-global-business-1
  ├─ Cluster: asdf-global-business-1 (https://192.168.138.58/kubernetes/business-1)
  └─ User: asdf-global-business-1

● myenv
  ├─ Cluster: myenv (https://192.168.138.58/kubernetes/global)
  └─ User: myenv

● myenv-business-1
  ├─ Cluster: myenv-business-1 (https://192.168.138.58/kubernetes/business-1)
  └─ User: myenv-business-1

===== No Orphaned Resources =====
All resources are properly referenced.
```

### 4. Delete Contexts

Delete contexts starting with `asdf` using a wildcard:

```bash
kontext delete --name "asdf*"
```

Output:

```
[kubeconfig.BackupKubeConfig] Created backup: /home/jing2uo/.kube/config.backup-20250613-104013
[kubeconfig.BackupKubeConfig] Removed old backup: /home/jing2uo/.kube/config.backup-20250613-102238
[kubeconfig.DeleteContext] Deleting contexts...
  ✓ Removed context: asdf-global
  ✓ Removed context: asdf-global-business-1
  ✓ Cleared current context setting
  ✓ Removed orphaned cluster: asdf-global
  ✓ Removed orphaned cluster: asdf-global-business-1
  ✓ Removed orphaned user: asdf-global
  ✓ Removed orphaned user: asdf-global-business-1

kubeconfig.DeleteContext Summary:
  ✓ Removed contexts: 2
  ✓ Current context reset
  ✓ Removed clusters: 2
  ✓ Removed users: 2
  ✓ Backup saved at: /home/jing2uo/.kube/config.backup-20250613-104013
==================================================
```

### 5. Clean Invalid Contexts

Validate and clean invalid or unreachable contexts:

```bash
kontext clean
```

Output(The config was manually altered to make this context as invalid."):

```
[kubeconfig.BackupKubeConfig] Created backup: /home/jing2uo/.kube/config.backup-20250613-104034
[kubeconfig.BackupKubeConfig] Removed old backup: /home/jing2uo/.kube/config.backup-20250613-103146
[kubeconfig.CleanContextCmd] Cleaning contexts...
  ✓ Removed invalid context: myenv-business-1
  ✓ Removed orphaned cluster: myenv-business-1
  ✓ Removed orphaned user: myenv-business-1

kubeconfig.CleanContextCmd Summary:
  ✓ Removed contexts: 1
  ✓ Removed clusters: 1
  ✓ Removed users: 1
  ✓ Backup saved at: /home/jing2uo/.kube/config.backup-20250613-104034
==================================================
```

## Command Reference

### `kontext add`

Add a new Kubernetes context.

```
kontext add --name <name> --host <host> --token <token> [--scan <type>]
```

- `--name`: Name for the context, cluster, and user (required).
- `--host`: Kubernetes API server address (required).
- `--token`: Authentication token (required).
- `--scan`: Cluster type for sub-cluster scanning (e.g., `alauda`).

### `kontext merge`

Merge an external kubeconfig file.

```
kontext merge --path <path> [--name <prefix>] [--scan <type>]
```

- `--path`: Path to the kubeconfig file (required).
- `--name`: Optional prefix for context names.
- `--scan`: Cluster type for sub-cluster scanning (e.g., `alauda`).

### `kontext list`

List all contexts and resource status.

```
kontext list
```

### `kontext delete`

Delete specified contexts, supporting wildcards.

```
kontext delete --name <name>
```

- `--name`: Context name to delete, supports wildcard patterns (e.g., `name*`) (required).

### `kontext clean`

Clean invalid or unreachable contexts and orphaned resources.

```
kontext clean
```

## Backup Management

- Backup files are stored as `~/.kube/config.backup-<timestamp>` (e.g., `config.backup-20250613-104034`).
- Up to 5 recent backups are retained, with older backups automatically deleted.
- Backups are created only for `delete` and `clean` commands when modifications occur.

## Dependencies

- Go 1.18 or higher (for building).
- Kubernetes client-go library (included in project dependencies).
- kubectl (for validating context connectivity).

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/xxx`).
3. Commit changes (`git commit -m 'Add feature xxx'`).
4. Push to the branch (`git push origin feature/xxx`).
5. Open a pull request.

## License

MIT License (to be confirmed, update based on actual project).

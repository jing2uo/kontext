# Kontext CLI

[![View on GitHub](https://img.shields.io/badge/GitHub-View%20on%20GitHub-blue?logo=github)](https://github.com/jing2uo/kontext)

Kontext is a command-line tool for efficiently managing Kubernetes contexts. It simplifies operations on `kubectl` configuration files (typically `~/.kube/config`) by enabling users to add, list, merge, delete, and clean Kubernetes contexts. Kontext supports sub-cluster scanning (e.g., Alauda clusters) and wildcard operations, making multi-cluster management seamless.

## Features

- **Add Contexts**: Quickly add Kubernetes contexts with name, server address, and token, including sub-cluster scanning.
- **List Contexts**: Display all current contexts and detect orphaned resources.
- **Delete Contexts**: Remove contexts by exact name or wildcard pattern, automatically cleaning orphaned clusters and users.
- **Clean Contexts**: Validate and remove invalid or unreachable contexts, along with orphaned clusters and users.
- **Merge Configs**: Merge external kubeconfig files into the current configuration, with optional name prefixes and sub-cluster scanning.
- **Backup Management**: Automatically create backups before modifying configurations (default: retain 5 backups in `~/.kube`).

## Installation

1. Download the appropriate binary from the [releases](https://github.com/jing2uo/kontext/releases) page, extract it, and move it to your `$PATH`:

   ```bash
   sudo mv kontext /usr/local/bin/
   ```

2. Verify the installation:

   ```bash
   kontext -h
   ```

## Command-Line Completion

To enable command completion (e.g., for Bash):

1. Generate the completion script:

   ```bash
   kontext completion bash > ~/.kontext-completion.bash
   ```

2. Add to your `~/.bashrc` or `~/.bash_profile`:

   ```bash
   source ~/.kontext-completion.bash
   ```

3. Reload your terminal:

   ```bash
   source ~/.bashrc
   ```

For Zsh or other shells, run `kontext completion <shell> -h` for details.

## Usage Examples

The following examples demonstrate Kontext’s core functionality using the `~/.kube/config` file.

### 1. Add a Context

Add a context named `myenv` with Alauda sub-cluster scanning:

```bash
kontext add --name myenv --host "https://192.168.138.58/kubernetes/global" --token "<your-token>" --scan alauda
```

**Output**:

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

**Output**:

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

**Output**:

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

**Output**:

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

Validate and remove invalid or unreachable contexts:

```bash
kontext clean
```

**Output** (assuming a manually invalidated context):

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

## Commands

### `kontext add`

Add a new Kubernetes context.

```
kontext add --name <name> --host <host> --token <token> [--scan <type>]
```

- `--name`: Context, cluster, and user name (required).
- `--host`: Kubernetes API server address (required).
- `--token`: Authentication token (required).
- `--scan`: Sub-cluster scan type (e.g., `alauda`).

### `kontext merge`

Merge an external kubeconfig file.

```
kontext merge --path <path> [--name <prefix>] [--scan <type>]
```

- `--path`: Path to kubeconfig file (required).
- `--name`: Context name prefix (optional).
- `--scan`: Sub-cluster scan type (e.g., `alauda`).

### `kontext list`

List all contexts and resource status.

```
kontext list
```

### `kontext delete`

Delete contexts, supporting wildcards.

```
kontext delete --name <name>
```

- `--name`: Context name to delete, supports wildcards (e.g., `name*`) (required).

### `kontext clean`

Remove invalid or unreachable contexts and orphaned resources.

```
kontext clean
```

## Backup Management

- Backups are stored as `~/.kube/config.backup-<timestamp>` (e.g., `config.backup-20250613-104034`).
- Retains the 5 most recent backups, automatically deleting older ones.
- Created only for `delete` and `clean` commands.

## Dependencies

- Go 1.18 or higher (for compilation).
- Kubernetes client-go library (included in project dependencies).
- `kubectl` (for context validation).

## Recommended Tools

Enhance your experience with:

1. [kubectx](https://github.com/ahmetb/kubectx): Quickly switch kubectl contexts and namespaces.
2. [fzf](https://github.com/junegunn/fzf): Enable interactive menus for kubectx.

## Contributing

Contributions are welcome! Follow these steps:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/xxx`).
3. Commit changes (`git commit -m 'Add feature xxx'`).
4. Push to the branch (`git push origin feature/xxx`).
5. Open a pull request.

## License

MIT License.

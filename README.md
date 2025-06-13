# Kontext CLI

Kontext 是一个用于高效管理 Kubernetes 上下文的命令行工具。它可以帮助用户在 kubectl 配置文件（通常位于 `~/.kube/config`）中添加、列出、合并、删除和清理 Kubernetes 上下文。Kontext 支持子集群扫描（如 Alauda 集群类型）和通配符操作，简化了多集群环境的管理。

## 功能特性

- **添加上下文**：通过名称、服务器地址和令牌快速添加新的 Kubernetes 上下文，支持子集群扫描。
- **合并配置文件**：将外部 kubeconfig 文件合并到当前配置中，支持自定义名称前缀和子集群扫描。
- **列出上下文**：显示所有当前的 Kubernetes 上下文，并检查是否有孤立资源。
- **删除上下文**：支持通过精确名称或通配符模式删除上下文，自动清理孤立集群和用户。
- **清理上下文**：验证并移除无效或不可达的上下文，以及孤立的集群和用户资源。
- **备份管理**：在修改配置前自动创建备份（默认保留 5 份历史备份，存储在 `~/.kube` 目录下）。
- **一致的输出**：清晰的命令行输出，包含操作进度、成功/失败详情和总结。

## 安装

目前，Kontext 是一个开发中的工具，建议从源代码编译安装。以下是安装步骤：

1. **克隆仓库**（假设项目托管在 Git 仓库中）：

   ```bash
   git clone https://github.com/your-repo/kontext.git
   cd kontext
   ```

2. **编译并安装**（需要 Go 环境）：

   ```bash
   go build -o kontext
   sudo mv kontext /usr/local/bin/
   ```

3. **验证安装**：
   ```bash
   kontext -h
   ```

## 使用示例

以下示例基于 `~/.kube/config` 配置文件，展示 Kontext 的核心功能。

### 1. 添加上下文

添加一个名为 `myenv` 的上下文，并扫描 Alauda 子集群：

```bash
kontext add --name myenv --host "https://192.168.138.58/kubernetes/global" --token "<your-token>" --scan alauda
```

输出：

```
[kubeconfig.AddContext] Adding contexts...
  ✓ Added context: myenv (https://192.168.138.58/kubernetes/global)
  ✓ Added context: myenv-business-1 (https://192.168.138.58/kubernetes/business-1)

kubeconfig.AddContext Summary:
  ✓ Added contexts: 2
  ✗ Failed contexts: 0
==================================================
```

### 2. 合并 kubeconfig 文件

合并一个外部 kubeconfig 文件，并扫描子集群：

```bash
kontext merge --path Downloads/asdf.yaml --scan alauda
```

输出：

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

### 3. 列出上下文

查看当前所有上下文：

```bash
kontext list
```

输出：

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

### 4. 删除上下文

使用通配符删除以 `asdf` 开头的上下文：

```bash
kontext delete --name "asdf*"
```

输出：

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

### 5. 清理无效上下文

验证并清理无效或不可达的上下文：

```bash
kontext clean
```

输出（这里手动修改了配置让这条 context 失效）：

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

## 命令详解

### `kontext add`

添加新的 Kubernetes 上下文。

```
kontext add --name <name> --host <host> --token <token> [--scan <type>]
```

- `--name`：上下文、集群和用户的名称（必填）。
- `--host`：Kubernetes API 服务器地址（必填）。
- `--token`：认证令牌（必填）。
- `--scan`：扫描子集群的类型（如 `alauda`）。

### `kontext merge`

合并外部 kubeconfig 文件。

```
kontext merge --path <path> [--name <prefix>] [--scan <type>]
```

- `--path`：kubeconfig 文件路径（必填）。
- `--name`：上下文名称前缀（可选）。
- `--scan`：扫描子集群的类型（如 `alauda`）。

### `kontext list`

列出所有上下文和资源状态。

```
kontext list
```

### `kontext delete`

删除指定的上下文，支持通配符。

```
kontext delete --name <name>
```

- `--name`：要删除的上下文名称，支持通配符（如 `name*`）（必填）。

### `kontext clean`

清理无效或不可达的上下文及孤立资源。

```
kontext clean
```

## 备份管理

- 备份文件存储在 `~/.kube/config.backup-<timestamp>`（如 `config.backup-20250613-104034`）。
- 默认保留最近 5 份备份，自动删除较旧的备份。
- 仅在 `delete` 和 `clean` 命令修改配置时创建备份。

## 依赖

- Go 1.18 或更高版本（用于编译）。
- Kubernetes client-go 库（自动包含在项目依赖中）。
- kubectl（用于验证上下文是否有效）。

## Tips

搭配以下两个工具使用体验更佳：

1.  kubectx: https://github.com/ahmetb/kubectx   快速切换 kubectl context 和 namespace 
2.  fzf: https://github.com/junegunn/fzf  让 kubectx 弹出选单

## 贡献

欢迎提交问题或拉取请求！请遵循以下步骤：

1. Fork 仓库。
2. 创建功能分支（`git checkout -b feature/xxx`）。
3. 提交更改（`git commit -m 'Add feature xxx'`）。
4. 推送到远程（`git push origin feature/xxx`）。
5. 创建拉取请求。

## 许可

MIT 许可证（待确认，建议根据实际项目更新）。

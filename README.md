# Kontext CLI

[![查看英文文档](https://img.shields.io/badge/GitHub-View%20English%20Version-blue?logo=github)](https://github.com/jing2uo/kontext/blob/main/README_en.md)

Kontext 是一个高效管理 Kubernetes 上下文的命令行工具，简化对 `kubectl` 配置文件（默认位于 `~/.kube/config`）的操作。支持添加、列出、合并、删除和清理 Kubernetes 上下文，具备子集群扫描（如 Alauda 集群）和通配符操作功能，适合多集群环境管理。

## 功能特性

- **添加上下文**：快速添加 Kubernetes 上下文，支持名称、服务器地址、令牌及子集群扫描。
- **列出上下文**：显示所有上下文，检测孤立资源。
- **删除上下文**：通过精确名称或通配符删除上下文，自动清理孤立集群和用户。
- **清理上下文**：验证并移除无效或不可达的上下文及孤立资源。
- **合并配置文件**：将外部 kubeconfig 文件合并到当前配置，支持名称前缀和子集群扫描。
- **备份管理**：修改配置前自动备份（默认保留 5 份，存储于 `~/.kube`）。

## 安装

1. 从 [releases](https://github.com/jing2uo/kontext/releases) 下载对应系统的二进制文件，解压后移至 `$PATH`：

   ```bash
   sudo mv kontext /usr/local/bin/
   ```

2. 验证安装：

   ```bash
   kontext -h
   ```

## 命令行补全

为提升使用体验，可启用 Cobra 命令补全（以 Bash 为例）：

1. 生成补全脚本：

   ```bash
   kontext completion bash > ~/.kontext-completion.bash
   ```

2. 添加到 `~/.bashrc` 或 `~/.bash_profile`：

   ```bash
   source ~/.kontext-completion.bash
   ```

3. 重新加载终端：

   ```bash
   source ~/.bashrc
   ```

查看其他 shell（如 Zsh）补全方法：`kontext completion <shell> -h`。

## 使用示例

以下示例基于 `~/.kube/config`，展示 Kontext 核心功能。

### 1. 添加上下文

添加名为 `myenv` 的上下文，并扫描 Alauda 子集群：

```bash
kontext add --name myenv --server "https://192.168.138.58/kubernetes/global" --token "<your-token>" --scan alauda
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

合并外部 kubeconfig 文件并扫描子集群：

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

查看所有上下文：

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
  ├─ Cluster‌خod: myenv-business-1 (https://192.168.138.58/kubernetes/business-1)
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

清理无效或不可达的上下文：

```bash
kontext clean
```

输出（假设手动使上下文失效）：

```
[kubeconfig.BackupKubeConfig] Created backup: /home/jing2uo/.kube/config.backup-20250613-104034
[kubeconfig.BackupKubekubeConfig] Removed old backup: /home/jing2uo/.kube/config.backup-20250613-103146
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

添加新 Kubernetes 上下文。

```
kontext add --name <name> --server <server> --token <token> [--scan <type>]
```

- `--name`：上下文、集群和用户名称（必填）。
- `--server`：Kubernetes API 服务器地址（必填）。
- `--token`：认证令牌（必填）。
- `--scan`：子集群扫描类型（如 `alauda`）。

### `kontext merge`

合并外部 kubeconfig 文件。

```
kontext merge --path <path> [--name <prefix>] [--scan <type>]
```

- `--path`：kubeconfig 文件路径（必填）。
- `--name`：上下文名称前缀（可选）。
- `--scan`：子集群扫描类型（如 `alauda`）。

### `kontext list`

列出所有上下文及资源状态。

```
kontext list
```

### `kontext delete`

删除指定上下文，支持通配符。

```
kontext delete --name <name>
```

- `--name`：上下文名称，支持通配符（如 `name*`）（必填）。

### `kontext clean`

清理无效或不可达的上下文及孤立资源。

```
kontext clean
```

## 备份管理

- 备份文件存储为 `~/.kube/config.backup-<timestamp>`（如 `config.backup-20250613-104034`）。
- 默认保留最近 5 份备份，自动删除较旧备份。
- 仅在 `delete` 和 `clean` 命令修改配置时生成备份。

## 依赖

- Go 1.18 或更高版本（用于编译）。
- Kubernetes client-go 库（自动包含）。
- `kubectl`（用于验证上下文有效性）。

## 推荐工具

搭配以下工具使用效果更佳：

1. [kubectx](https://github.com/ahmetb/kubectx)：快速切换 kubectl 上下文和命名空间。
2. [fzf](https://github.com/junegunn/fzf)：为 kubectx 提供交互式菜单。

## 贡献

欢迎提交问题或拉取请求：

1. Fork 仓库。
2. 创建功能分支：`git checkout -b feature/xxx`。
3. 提交更改：`git commit -m 'Add feature xxx'`。
4. 推送分支：`git push origin feature/xxx`。
5. 创建拉取请求。

## 许可

MIT 许可证。

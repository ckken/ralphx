# 安装说明

## 推荐：通过 GitHub Release 安装

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
```

## 运行依赖

必需：
- `codex`
- `curl` 或 `wget`

建议：
- `git`
- `gh`
- `bash`
- `python3`

可选：
- `jq`（仅 legacy 辅助；Go 主链路不依赖）

## 安装后的布局

稳定入口：
- `~/.local/bin/ralphx`
- `~/.local/bin/ralphx-doctor`

执行路径持久化：
- `~/.config/ralphx/current.env`

下载下来的版本二进制：
- `~/.local/share/ralphx/releases/`

如果 shell 找不到 `ralphx`，加入：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 验证

```bash
ralphx-doctor
ralphx --help
ralphx version
```

## 卸载

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/uninstall.sh | bash
```

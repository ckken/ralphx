# 安装说明

## 必需工具

- `bash`
- `jq`
- `python3`
- `codex`

## 建议工具

- `git`
- `gh`
- `timeout` 或 `gtimeout`

## 安装

```bash
git clone https://github.com/ckken/codex-ralph.git
cd codex-ralph
./install.sh
```

默认会把命令安装到 `~/.local/bin`。

如果 shell 找不到 `codex-ralph`，加入：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 验证

```bash
codex-ralph-doctor
codex-ralph --help
```

## 卸载

```bash
./uninstall.sh
```

# 安装说明

## 必需工具

必需：
- `go`
- `codex`

建议：
- `git`
- `gh`
- `bash`
- `python3`

可选：
- `jq`（仅 legacy 辅助；Go 主链路不依赖）

## 从源码安装

```bash
git clone https://github.com/ckken/ralphx.git
cd ralphx
./install.sh
```

默认会把命令安装到 `~/.local/bin`。

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

## 自定义安装目录

```bash
PREFIX=/custom/bin ./install.sh
```

## 卸载

```bash
./uninstall.sh
```

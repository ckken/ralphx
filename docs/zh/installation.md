# 安装说明

## 推荐：通过 GitHub Release 安装

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
```

安装器会下载 `SHA256SUMS`，校验通过后才激活二进制。

## 验证

```bash
ralphx-doctor
ralphx --help
ralphx version
ralphx current
```

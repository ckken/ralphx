# 生产 SOP

## 安装

使用 release 安装，不走源码安装。

最新版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
```

安装器会先校验 `SHA256SUMS`，再激活二进制。

当前激活执行路径持久化在：

```bash
~/.config/ralphx/current.env
```

查看当前持久化执行目标：

```bash
ralphx current
```

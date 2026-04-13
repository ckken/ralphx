     1|# 安装说明
     2|
     3|## 推荐：通过 GitHub Release 安装
     4|
     5|正常使用不需要源码安装。
     6|
     7|安装最新 release：
     8|
     9|```bash
    10|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    11|```
    12|
    13|安装指定版本：
    14|
    15|```bash
    16|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    17|```
    18|
    19|## 运行依赖
    20|
    21|必需：
    22|- `codex`
    23|- `curl` 或 `wget`
    24|
    25|建议：
    26|- `git`
    27|- `gh`
    28|- `bash`
    29|- `python3`
    30|
    31|可选：
    32|- `jq`（仅 legacy 辅助；Go 主链路不依赖）
    33|
    34|## 安装后的布局
    35|
    36|稳定入口：
    37|- `~/.local/bin/ralphx`
    38|- `~/.local/bin/ralphx-doctor`
    39|
    40|执行路径持久化：
    41|- `~/.config/ralphx/current.env`
    42|
    43|下载下来的版本二进制：
    44|- `~/.local/share/ralphx/releases/`
    45|
    46|如果 shell 找不到 `ralphx`，加入：
    47|
    48|```bash
    49|export PATH="$HOME/.local/bin:$PATH"
    50|```
    51|
    52|## 验证
    53|
    54|```bash
    55|ralphx-doctor
    56|ralphx --help
    57|ralphx version
    58|```
    59|
    60|## 卸载
    61|
    62|```bash
    63|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/uninstall.sh | bash
    64|```
    65|
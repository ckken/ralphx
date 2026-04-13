     1|# Installation
     2|
     3|## Preferred: install from GitHub release
     4|
     5|No source build is required for normal usage.
     6|
     7|Install the latest release:
     8|
     9|```bash
    10|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    11|```
    12|
    13|Install a specific version:
    14|
    15|```bash
    16|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    17|```
    18|
    19|## Runtime requirements
    20|
    21|Required:
    22|- `codex`
    23|- `curl` or `wget`
    24|
    25|Recommended:
    26|- `git`
    27|- `gh`
    28|- `bash`
    29|- `python3`
    30|
    31|Optional:
    32|- `jq` (legacy-only helper; not required by the Go-native main path)
    33|
    34|## Installed layout
    35|
    36|Wrappers:
    37|- `~/.local/bin/ralphx`
    38|- `~/.local/bin/ralphx-doctor`
    39|
    40|Persistent execution state:
    41|- `~/.config/ralphx/current.env`
    42|
    43|Downloaded release binaries:
    44|- `~/.local/share/ralphx/releases/`
    45|
    46|If your shell cannot find `ralphx`, add:
    47|
    48|```bash
    49|export PATH="$HOME/.local/bin:$PATH"
    50|```
    51|
    52|## Verify
    53|
    54|```bash
    55|ralphx-doctor
    56|ralphx --help
    57|ralphx version
    58|```
    59|
    60|## Uninstall
    61|
    62|```bash
    63|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/uninstall.sh | bash
    64|```
    65|
# ralphx Go 重写 MVP 计划

> 给 Hermes：使用“子代理驱动开发”来执行此计划；当任务彼此独立时可并行推进，但共享文件的编辑必须串行化。

目标：在不破坏当前 Bash 回退路径的前提下，原地启动 Go 重写，并落地一个可编译的单二进制 CLI，以及最小化的、面向多代理扩展的架构。

架构：保留当前仓库与资源，立即增加一个 Go 前门；保留 `.ralphx` 作为持久化本地状态契约；并为 `Agent`、`Runner` 以及未来的 `Scheduler` / `Worker` 引入清晰的边界。Go 的第一阶段里程碑应当能够编译，暴露 `run`、`doctor` 和 `version` 命令，并在 Go 原生运行时逐步完善的同时，通过同一个 CLI 将 `run` 委托给现有的 Bash 循环。

技术栈：Go 1.19+，优先使用标准库，本地文件系统状态，针对遗留执行路径保留 shell 回退，以及以 JSON 为优先的契约设计。

---

## MVP 范围

1. 添加 `go.mod` 和一个可编译的 `cmd/ralphx` 二进制。
2. 添加一个精简的、兼容 `cmd/ralphx-doctor` 的二进制。
3. 为 CLI 分发、配置、版本、doctor 与遗留执行实现最小内部包。
4. 临时保留当前 shell 运行时与相关资源。
5. 为即将到来的 runner 和 worker 协议添加初始的 Go 原生 state / contract 类型。
6. 添加初始的 scheduler / worker 脚手架，但暂不启用真正的并行执行。
7. 验证构建、帮助输出、doctor 输出，以及对遗留逻辑的委托。

## 本轮非目标

- 完整等价重写 `ralphx-loop.sh`
- TUI
- 网络服务 / 守护进程
- 插件系统
- 更丰富的规划器或任务拆解逻辑
- 真正的并行 worker 执行

## 本轮目标文件布局

- `go.mod`
- `cmd/ralphx/main.go`
- `cmd/ralphx-doctor/main.go`
- `internal/cli/app.go`
- `internal/config/config.go`
- `internal/version/version.go`
- `internal/doctor/doctor.go`
- `internal/legacy/exec.go`
- `internal/contracts/result.go`
- `internal/state/types.go`
- `internal/parallel/types.go`
- `docs/plans/2026-04-13-go-rewrite-mvp.md`

## 任务 1：初始化 Go module

目标：创建一个可编译的 Go module，并以最终 GitHub 路径作为根模块路径。

文件：

- 新建：`go.mod`
- 新建：`cmd/ralphx/main.go`
- 新建：`cmd/ralphx-doctor/main.go`

步骤：

1. 创建 `go.mod`，模块路径为 `github.com/ckken/ralphx`。
2. 为 `ralphx` 添加最小化的 `main.go`，调用 `internal/cli`。
3. 为 `ralphx-doctor` 添加最小化的 `main.go`，调用 `internal/doctor`。
4. 使用以下命令验证：
   - `go build ./cmd/ralphx`
   - `go build ./cmd/ralphx-doctor`

## 任务 2：补齐 CLI / config / version 管线

目标：为该二进制提供稳定的命令接口，以及环境变量 / 标志处理能力。

文件：

- 新建：`internal/cli/app.go`
- 新建：`internal/config/config.go`
- 新建：`internal/version/version.go`

步骤：

1. 定义初始命令：
   - `ralphx run`
   - `ralphx doctor`
   - `ralphx version`
2. 为兼容性起见，使裸调用 `ralphx --task ...` 映射到 `run`。
3. 支持遗留 Bash 流程中的关键环境变量。
4. 添加简单的构建信息 / 版本字符串。
5. 使用以下命令验证：
   - `go run ./cmd/ralphx --help`
   - `go run ./cmd/ralphx version`

## 任务 3：补齐 doctor 与 legacy 执行桥接

目标：即使在 Go 尚未实现功能对等之前，也要立即向用户交付一个可用的二进制。

文件：

- 新建：`internal/doctor/doctor.go`
- 新建：`internal/legacy/exec.go`

步骤：

1. 用 Go 原生方式实现对 `bash`、`python3`、`git`、`gh`、`codex` 以及可选的 `jq` 的 doctor 检查。
2. 添加遗留脚本执行辅助逻辑，以便定位仓库根目录下的脚本，并使用透传参数执行它们。
3. 先让 `ralphx run ...` 委托给 `./ralphx-loop.sh`。
4. 使用以下命令验证：
   - `go run ./cmd/ralphx-doctor`
   - `go run ./cmd/ralphx run --help` 或参数校验
   - `go run ./cmd/ralphx --task ./examples/sample-task.md --workdir .`（冒烟路径）

## 任务 4：补齐 Go 原生 contract 与 state 类型

目标：落地当前 Bash 运行时所隐含的稳定类型。

文件：

- 新建：`internal/contracts/result.go`
- 新建：`internal/state/types.go`

步骤：

1. 定义与当前 JSON schema 匹配的 `RoundResult`。
2. 定义与当前 `.ralphx` 文件足够接近的 `RunState` 和 `Stats` 类型，以支持低风险迁移。
3. 按需添加 JSON 标签以及辅助构造器 / 校验器。
4. 当前先用 `go build ./...` 验证；当测试补齐后，再使用 `go test ./...` 验证。

## 任务 5：补齐多 agent 基础类型

目标：让并行能力成为一等的未来模式，但现在不实现完整 scheduler。

文件：

- 新建：`internal/parallel/types.go`

步骤：

1. 定义 `Job`、`WorkerState`、`WorkerResult`，以及 `Scheduler` / `Worker` 接口。
2. 将这些类型保持为仅本地、基于文件的实现。
3. 确保设计保留以下约束：
   - leader 负责完成态
   - worker 只负责边界明确的切片任务
   - 每个 worker 的结果文件在追加 / 覆盖时都是安全的
4. 通过编译该包来验证，并在合适处从 CLI 包中引用这些类型。

## 任务 6：验证并记录 bootstrap 状态

目标：确认 Go 前门可用，且仓库拥有清晰的迁移方向。

文件：

- 如有需要后续修改：`README.md`

步骤：

1. 运行：
   - `gofmt -w ./cmd ./internal`
   - `go build ./...`
   - `go run ./cmd/ralphx --help`
   - `go run ./cmd/ralphx version`
   - `go run ./cmd/ralphx-doctor`
2. 记录所有缺口。
3. 在 Go 运行时替代 Bash 核心之前，尽量保持文档更新最小化。

## 本轮验收标准

- 仓库包含一个有效的 Go module。
- `ralphx` 与 `ralphx-doctor` 可以编译。
- `ralphx doctor` 可由 Go 原生实现并正常工作。
- `ralphx run` 已存在，并且可以桥接到遗留 Bash 运行时。
- 核心 Go contract / state / parallel 类型已具备，为下一轮实现铺路。
- 当前 Bash 实现仍可作为回退路径使用。

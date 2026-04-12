# 方法论

`codex-ralph` 是一个基于 Bash 的 Codex 外层控制器。它的方法可以概括为：

1. 任务文件作为唯一目标源
2. 使用非交互模式调用 Codex
3. 强制要求严格 JSON 返回
4. 拒绝弱完成信号
5. 每轮之间跑验证
6. 直到真实任务完成或出现真实阻塞才停止

## 核心原则

### 1. 外层控制

是否完成，不由模型单独决定，而由 Bash 外层循环控制：

- 循环次数
- 超时
- 验证
- 无进展判断
- checklist gate

### 2. Checklist 硬门

如果附带 markdown checklist：

- 任何未勾选项都代表剩余硬任务
- 即使局部切片完成，也不能结束总任务
- 只要 checklist 未清空，就拒绝 `complete`

### 3. 验证优先

每一轮可以挂低成本验证链，再决定是否继续。

典型验证：

- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`
- `bash scripts/verify-batch.sh --skip-build`

### 4. 防止过早完成

`codex-ralph` 会拒绝这类错误停止：

- Codex 返回 `complete`
- 但实际上没有新的代码改动

这能防止模型“总结了”却没有真正推进。

## 推荐工作流

1. 先写总任务文件
2. 如果任务跨多个里程碑，再配 checklist
3. 绑定最低成本但足够有用的验证链
4. 启动循环
5. 只有 checklist 清空且验证通过时，才接受最终完成

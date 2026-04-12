目标：完成当前 B 批次，不是只判断“当前切口是否已经差不多”，而是持续推进，直到 `shared/code-registry` 抽取达到明确终态。

最终终态：
- `src/render/shared/code-registry.ts` 成为唯一的公共基座，维护公共类型、公共 helper、公共 base class。
- `src/render/dsl-core/modules/code-registry.ts` 只保留 dsl-core 侧不可下沉的实例化差异。
- `src/render/dsl-react/generator/registry/code-registry/index.ts` 只保留 dsl-react 侧不可下沉的模板或实例化差异。
- 不再保留可以直接复用 shared 的重复类型、重复 helper、重复逻辑。
- 当前批次相关验证全部通过。

严格边界：
- 只处理 `src/render/shared/`、`src/render/dsl-core/modules/code-registry.ts`、`src/render/dsl-react/generator/registry/code-registry/index.ts` 以及它们的直接调用方。
- 可以修复在这一条抽取链上暴露出来的局部重复定义、错误转发、错误导入、明显 bug。
- 不进入 `contracts -> dsl-react -> dsl-base` 新切口。
- 不做大范围重命名。
- 不回退用户已有重构。

必须持续推进的子目标：
1. 找出仍然重复维护在 wrapper 层而本应归 shared 的内容，并完成迁移或删除。
2. 找出仍然依赖旧入口特殊行为的直接调用方，并把它们收敛到薄壳接口。
3. 删除这一条链路上已经无意义的重复类型定义或重复导出。
4. 每轮改动后都走低消耗验证，不通过就继续修，不允许在失败状态下结束。

低消耗验证链：
- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`
- `bash scripts/verify-batch.sh --skip-build`
- `python3 scripts/inventory-render-noise.py --json --out .omx/reports/render-noise-baseline.json`
- `node scripts/inventory-relative-imports.cjs --json --out .omx/reports/residual-import-inventory.json`

停止条件：
- 只有当下面所有条件同时成立，才允许返回 `status: "complete"` 和 `exit_signal: true`：
  - 这条抽取链上没有继续明显可做的重复定义、重复 helper、重复转发。
  - 两个 wrapper 都已经是纯薄壳，没有继续下沉到 shared 的明显机会。
  - 上述低消耗验证链全部通过。

禁止过早结束：
- 不能因为“当前看起来已经够薄”就结束。
- 不能因为“这轮没改动”就结束，除非已经明确证明终态成立。
- 不能把“当前切口已收口到合理停止点”当作完成条件，必须按本文件的最终终态判断。

输出要求：
- 只返回一个 JSON 对象。
- 如果还没达到最终终态，返回：
  - `status: "in_progress"`
  - `exit_signal: false`
- 如果存在阻塞且无法在当前边界内解决，返回：
  - `status: "blocked"`
  - `exit_signal: false`
  - `blockers` 写明原因
- 只有达到最终终态时，才返回：
  - `status: "complete"`
  - `exit_signal: true`

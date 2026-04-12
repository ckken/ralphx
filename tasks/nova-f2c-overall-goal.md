目标：完成当前 `nova-f2c` 迁移与收口任务，不是完成某一个局部子任务，而是持续推进直到整个当前目标达到可交付状态。

总体目标：
- 以 `docs/migration-plan.md` 和 `docs/workflow.md` 为准，完成当前迁移批次。
- 当前批次的终点不是某个局部文件“看起来差不多”，而是把当前迁移范围收口到一个一致、可验证、可继续演进的状态。

当前范围：
1. `src/cli`
2. `src/core/restapi-adapter`
3. `src/render/dsl-core`
4. `src/render/dsl-react`
5. `src/core/contracts`

当前阶段要求：
- `B` 已经收口，不要在当前批次上重复兜圈。
- 现在自动进入下一批次，继续推进整体迁移目标。
- 下一批次按现有文档优先走 `contracts` / `restapi-adapter` 最小闭包收敛主线。
- 不要把“上一批已完成”当作整个总任务完成。

整体完成标准：
- 已完成的批次会被记录，但不会因此结束总任务。
- 当前批次相关的重复定义、重复 helper、重复转发已经尽可能下沉或删除。
- `contracts` / `restapi-adapter` 主线上已经开始出现真实收敛，而不是只停留在分析或文档层。
- 当前工作流规定的低消耗验证链全部通过。
- 当前主线没有明显下一刀却被遗漏的直接重复面。

严格边界：
- 不回退用户已有改动。
- 不进入无关功能开发。
- 不做大范围重命名。
- 优先删除、复用和收口，而不是增加新抽象。
- 如果需要切到新的主线，必须先把当前主线收口到一个明确稳定点。

执行原则：
- 你必须把任务文件视为“总任务”，不是“一个子任务说明”。
- 如果你完成了某个批次，必须自动进入下一个批次，而不是直接结束。
- 如果你完成了某个局部文件的收口，但整体目标还没完成，必须返回 `in_progress`。
- 如果当前主线还有明显可继续推进的下一步，不能返回 `complete`。
- 只有在当前整体迁移目标达到稳定终态，或者出现真实阻塞无法继续时，才允许停止。

低消耗验证链：
- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`
- `bash scripts/verify-batch.sh --skip-build`
- `python3 scripts/inventory-render-noise.py --json --out .omx/reports/render-noise-baseline.json`
- `node scripts/inventory-relative-imports.cjs --json --out .omx/reports/residual-import-inventory.json`

停止条件：
- 只有整体目标完成时，返回：
  - `status: "complete"`
  - `exit_signal: true`
- 如果还有任何未完成的明显后续切口，返回：
  - `status: "in_progress"`
  - `exit_signal: false`
- 如果无法继续且阻塞真实存在，返回：
  - `status: "blocked"`
  - `exit_signal: false`
  - `blockers` 说明阻塞原因

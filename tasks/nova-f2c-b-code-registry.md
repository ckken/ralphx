目标：继续当前 B 批次，收口 shared/code-registry 抽取。

严格边界：
- 只处理 `src/render/shared/` 和 `code-registry` 相关抽取。
- 不进入 `contracts -> dsl-react -> dsl-base` 新切口。
- 不做大范围重命名。
- 不回退用户已有重构。

优先目标：
1. 让 `src/render/dsl-core/modules/code-registry.ts` 保持为薄壳，只保留 dsl-core 侧必须存在的实例化差异。
2. 让 `src/render/dsl-react/generator/registry/code-registry/index.ts` 保持为薄壳，只保留 dsl-react 侧模板或实例化差异。
3. 真正公共的类型、helper、基类只维护在 `src/render/shared/code-registry.ts`。
4. 优先删除重复定义或重复转发，不要引入新的抽象层。

验收要求：
- `bun src/index.ts --help` 可正常运行。
- 黄金样本回归不漂移。
- 当前低消耗验证链必须通过。

输出要求：
- 只返回一个 JSON 对象。
- 如果还没完全收口，就返回 `status: "in_progress"`。
- 只有在这一个切口已经收口到合理停止点时，才返回 `status: "complete"` 和 `exit_signal: true`。

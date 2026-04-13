# ralphx Autodrive Checklist

Goal: align `ralphx` with the useful parts of the Ralph workflow without adding a dedicated compact phase.

## Phase 1: Planning Entry

- [x] Add `ralphx plan --goal --out`
- [x] Add planner JSON schema
- [x] Write generated `task.md` and `task.checklist.md`
- [x] Support `ralphx plan --execute` to hand off to the existing runner
- [ ] Document planner usage in README and installation docs

## Phase 2: Replanning

- [x] Add `ralphx replan --task`
- [x] Read `.ralphx/summary.txt` and `state.json` as replanning context
- [x] Regenerate the next task/checklist when blocked or stale
- [ ] Preserve completed checklist items where safe

## Phase 3: Session Rollover

- [x] Add persisted session metadata under `.ralphx/`
- [x] Add `--resume`
- [x] Add `--session-expiry`
- [x] Roll to a fresh session using task + checklist + summary instead of full history

## Phase 4: Runner Integration

- [x] Trigger replanning after repeated blocked / no-progress rounds
- [ ] Surface next-step guidance in state files
- [ ] Keep checklist gating as the final completion authority

## Notes

- Do not add `compact` in the first pass.
- Prefer summary rollover over transcript compression.
- Keep planning and execution schemas separate.

# ralphx Hook Control Plan

Goal: add an OMX-style control layer above the current runner so `ralphx` can intercept unsafe exits and keep execution on the current branch of work instead of stopping at weak recommendations.

## Hook Surfaces

- `session-start`
- `pre-tool-use`
- `post-tool-use`
- `turn-complete`
- `stop`
- `session-end`

## First Bounded Slice

- [x] Add hook event vocabulary under `internal/hooks`
- [x] Add stop/session-end guard evaluator
- [x] Add tests for incomplete-work, missing-verification, and clean-complete outcomes
- [x] Wire the stop guard into a real hook entrypoint
- [x] Persist hook decisions into `.ralphx/state.json`

## Recommended Wiring Order

1. Stop/session-end guard
2. Turn-complete continuation nudge
3. Pre/Post tool safety hooks
4. Optional extensibility/plugin layer

## Design Notes

- Keep the stop guard deterministic and repo-local.
- Use tool-backed facts only: checklist count, latest result mode/status, verification state.
- Do not depend on natural-language interpretation alone at the hook layer.
- Use hook blocks to force continuation, not to generate plans directly.

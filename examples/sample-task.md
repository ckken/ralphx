Goal: finish the current migration batch.

Rules:
- Treat this as the total objective, not a local subtask.
- Prefer deleting duplication over adding abstraction.
- Keep validation green after each successful step.

Validation:
- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`

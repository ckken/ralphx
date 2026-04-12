# nova-f2c overall goal checklist

- [x] Finish the current `B` batch so the shared extraction has a stable boundary beyond the current `code-registry` slice.
- [x] Reduce remaining direct duplicate helper or type surfaces in the current render/shared extraction line where they are still obviously shared.
- [x] Confirm the current render/shared line no longer has another immediate same-shape extraction step before claiming the batch is done.
- [x] Keep the low-cost validation chain green while progressing the current batch.
- [x] Update migration/workflow docs only as needed to reflect the actual settled execution rules for the current batch.
- [x] Open the next batch explicitly and move from the completed `B` slice into the next migration mainline instead of stopping.
- [x] Start the `contracts` / `restapi-adapter` minimal-closure reduction with at least one real code-level consolidation, not just analysis text.
- [x] Keep the low-cost validation chain green while progressing the new batch.
- [x] Update migration/workflow docs only as needed to reflect the new active batch boundary after `B`.
- [x] Remove the now-unused `src/core/restapi-adapter/semantics/layout/type.ts` duplicate after inlining the local relative-position union.
- [x] Lift the shared `RelativePoisition` surface into `src/core/contracts/figma-types.ts` and delete the remaining dsl-react local copy.

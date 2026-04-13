# Architecture

This document explains the runtime flow of `ralphx` with a process diagram and a component chain diagram.

## Process Flow

```mermaid
flowchart TD
    A[Start ralphx] --> B[Load task file]
    B --> C{Checklist file exists?}
    C -- yes --> D[Load checklist and count open items]
    C -- no --> E[Continue without checklist]
    D --> F[Build prompt]
    E --> F
    F --> G[Run codex exec non-interactively]
    G --> H[Parse strict JSON result]
    H --> I{Valid JSON?}
    I -- no --> J[Mark blocked and stop]
    I -- yes --> K[Compare git status before and after]
    K --> L{Premature complete?}
    L -- yes --> M[Force in_progress]
    L -- no --> N[Keep model result]
    M --> O{Run tests?}
    N --> O
    O -- yes --> P[Run validation chain]
    O -- no --> Q{Checklist still open?}
    P --> R{Validation passed?}
    R -- no --> S[Mark blocked and stop]
    R -- yes --> Q
    Q -- yes --> T[Force in_progress]
    Q -- no --> U{status=complete and exit_signal=true?}
    T --> V[Next loop]
    U -- yes --> W[Stop successfully]
    U -- no --> X{status=blocked?}
    X -- yes --> Y[Stop blocked]
    X -- no --> V
```

## Component Chain

```mermaid
flowchart LR
    User[User] --> Task[Task file]
    User --> Checklist[Checklist file]
    Task --> Loop[ralphx-loop.sh]
    Checklist --> Loop
    Prompt[loop-system-prompt.md] --> Loop
    Schema[loop-output.schema.json] --> Loop
    Loop --> Codex[codex exec]
    Codex --> Json[Strict JSON result]
    Json --> Gate[Completion gates]
    Gate --> Validate[Validation chain]
    Validate --> State[.ralphx state files]
    State --> Loop
    Gate --> Stop[Stop only when total task is done]
```

## Why both diagrams matter

- The process flow explains control logic.
- The component chain explains what files and tools participate in the loop.
- Together they make the method easier to adopt in another repository.

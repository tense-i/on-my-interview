# AGENTS.md — Repo playbook for coding agents

This repository uses a traceable "Context Engineering" workflow:
every task MUST have a Context Manifest that records what context was loaded/excluded and why,
to keep reasoning bounded under the token window and reproducible. (Everything is Context)

Project development follows the **superpowers** approach and skill workflows.
Use relevant superpowers skills as guidance for planning, implementation, debugging, testing, and review.

---

## 0) Repo context layout (authoritative)

- `rules/` : small, always-on rules (autoload). Keep minimal.
- `memory/` : long-lived memory
  - `memory/FACTS.md`        : project facts (stable, atomic)
  - `memory/EXPERIENCE.md`   : pitfalls / playbooks (trigger -> symptom -> diagnosis -> fix -> prevention)
  - (optional) `memory/DECISIONS.md` : ADR-lite decisions
- `skills/` : task-triggered playbooks (load only when relevant)
- `docs/`   : on-demand background docs (load only specific sections)

---

## 1) Task closeout and memory updates

After completing a task, decide based on actual outcomes whether memory files should be updated.
Only write memory when the result is stable, reusable, and worth retaining.

Reference checklist:
- Updated `memory/FACTS.md`: <what/where>
- Updated `memory/EXPERIENCE.md`: <what/where>
- (optional) Updated `memory/DECISIONS.md`: <what/where>

---

## 2) Development Workflow Rules (Mandatory)

### Superpowers Development Workflow

This project uses the **superpowers** development model.
superpowers is responsible for planning, implementation, debugging, testing, and review activities.
All feature development must follow this workflow:

```
1. Use superpowers to clarify requirements and create the task plan
2. Implement against the task plan and keep progress tracked
3. After implementation, pass tests/builds/required verification
4. Complete verification before marking work as done
```

#### Key Rules

- **Do not skip planning and jump straight into coding**. Every new feature or module must be planned with superpowers before implementation starts.
- **Prefer the superpowers skill that best matches the task**. For example: planning, TDD, debugging, code review, and pre-completion verification should each use the corresponding skill.
- **Do not mark work as Done immediately after coding**. It must first pass tests and complete verification.
- **After backend Go development is complete, you must run the Go 1.26 `go fix` pass before final verification**. Recommended sequence: first run `cd server && GOTOOLCHAIN=go1.26.0 go fix -diff ./...` to inspect the refactors that would be applied, then run `cd server && GOTOOLCHAIN=go1.26.0 go fix ./...` to apply them, and only then proceed to tests/build/acceptance. Even if `go fix` produces no diff, it must still be run explicitly once.
- **UI changes must be verified with CDP (chrome-devtools MCP)**:
  - After modifying components, styles, or layouts, must verify effects via chrome-devtools MCP
  - Verification flow: `npm run dev` → open page via CDP at `http://localhost:3000` → screenshot to confirm rendering → check console for errors
  - Interactive changes (buttons, forms, navigation) require CDP click/input simulation with screenshots
  - Responsive layout changes require CDP device emulation to verify both desktop and mobile viewports

#### Recommended superpowers skill mapping

- **Task planning**: `writing-plans`, `product-requirements`
- **Task execution/orchestration**: `subagent-driven-development`, `executing-plans`
- **Implementation method**: `test-driven-development`, `dev`
- **Debugging/troubleshooting**: `systematic-debugging`, `root-cause-analyzer`
- **Code review/pre-completion verification**: `requesting-code-review`, `verification-before-completion`

For the detailed process, refer to the relevant superpowers skills under `skills/`.

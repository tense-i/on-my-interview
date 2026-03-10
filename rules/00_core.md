# Core Rules (autoload)

## Goal
- Keep work in this repository traceable, reproducible, and bounded under the Context Engineering workflow.
- Prefer small, verified changes that preserve current runtime behavior before expanding roadmap intent.

## Output Contract
- Start by stating the task and the first inspection step.
- When reporting results, cite exact files touched and the verification performed.
- When reverse-engineering behavior, separate `Observed`, `Inferred`, and `Open Question`.

## Context Loading Order
- Load the smallest useful context first:
  - `AGENTS.md`
  - `rules/00_core.md`
  - `memory/FACTS.md`
  - only the skill(s) relevant to the task
  - only the specific docs/code sections needed next
- Do not bulk-load `docs/`, logs, generated artifacts, or unrelated history.
- Every task must leave behind a Context Manifest under `docs/plans/`.

## Working Memory Discipline
- Use `today.md` as the scratchpad for temporary goals, hypotheses, commands, and raw findings.
- Promote only stable, reusable conclusions into:
  - `memory/FACTS.md` for project facts
  - `memory/EXPERIENCE.md` for pitfalls/playbooks
  - `memory/DECISIONS.md` for ADR-like decisions when needed

## Safety / Guardrails
- Do not present roadmap intent or temporary notes as implemented behavior.
- Do not guess missing facts; mark assumptions and provide a verification path.
- Do not claim completion without fresh verification evidence, or an explicit statement of what was not verified.

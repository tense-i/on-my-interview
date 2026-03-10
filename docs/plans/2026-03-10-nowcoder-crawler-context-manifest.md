# Context Manifest: Go Interview Crawler Service

**Task:** Build a Go HTTP service for crawling interview-experience posts, deduplicating into MySQL, extracting structured interview questions via pluggable LLM backends, and exposing query APIs.

**Date:** 2026-03-10

## Loaded Context

- `AGENTS.md`
  - Reason: repository workflow, mandatory skill usage, verification and closeout rules.
- `rules/00_core.md`
  - Reason: output contract, context discipline, scratchpad and manifest requirement.
- `rules/10_release.md`
  - Reason: release and commit guardrails.
- `memory/FACTS.md`
  - Reason: confirm any existing project invariants and avoid assuming unavailable structure.
- `memory/EXPERIENCE.md`
  - Reason: check reusable pitfalls and verification conventions.
- `newcoder-crawler/README.md`
  - Reason: understand current Python crawler scope and data assets.
- `newcoder-crawler/crawler.py`
  - Reason: inspect current NowCoder crawling and MySQL dedupe behavior.
- `newcoder-crawler/crawler_advanced.py`
  - Reason: inspect existing filtering and incremental crawling refinements.
- `newcoder-crawler/create_table_newcoder_search.sql`
  - Reason: understand current storage model and dedupe key.
- `today.md`
  - Reason: confirm active task scratchpad content.
- `/Users/zh/.codex/skills/using-superpowers/SKILL.md`
  - Reason: mandatory process skill.
- `/Users/zh/.codex/skills/brainstorming/SKILL.md`
  - Reason: required before feature design.
- `/Users/zh/.codex/skills/writing-plans/SKILL.md`
  - Reason: required before implementation for multi-step work.
- `/Users/zh/.codex/skills/test-driven-development/SKILL.md`
  - Reason: required implementation method.
- `/Users/zh/.codex/skills/subagent-driven-development/SKILL.md`
  - Reason: evaluate same-session plan execution workflow.
- `/Users/zh/.codex/skills/verification-before-completion/SKILL.md`
  - Reason: mandatory completion verification gate.

## Runtime Checks Performed

- `curl -s 'https://gw-c.nowcoder.com/api/sparta/pc/search' ...`
  - Reason: verify current NowCoder search API still returns post content inline.
  - Outcome: confirmed `contentData.title`, `contentData.content`, `createTime`, `editTime` are present in the response on 2026-03-10.

## Context Excluded

- `train.csv`, `test.csv`, `generated_pesudo_data.csv`
  - Reason: model-training artifacts are not needed for the Go service implementation path.
- `newcoder-crawler/bert_train.py`, `seq2seq.py`, `predict.py`, `generate.py`
  - Reason: training/inference code is out of scope for the first service version.
- Any nonexistent `server/`, `web/`, or `.git` metadata in this workspace
  - Reason: they are referenced in long-lived memory but not present in this checkout, so they cannot be treated as implemented context.

## Observed Constraints

- Current repository is not a git working tree root.
- No existing Go backend module is present.
- Existing implementation is a Python crawler prototype only.
- The first Go version should support multiple future platforms, not just NowCoder.

## Assumptions Chosen

- Build a single-process Go HTTP service under `server/`.
- Use environment-variable configuration for the first version.
- Use an in-process scheduler plus HTTP-triggered jobs.
- Keep LLM integration behind an interface; ship an OpenAI-compatible implementation first.

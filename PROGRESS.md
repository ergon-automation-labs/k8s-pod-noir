# Project progress

This file is a **small index** (target **≤80 lines**). Put dated detail in **[docs/progress/](docs/progress/)** by month—link out instead of growing this file.

## Snapshot

_Update when you ship player-facing, scenario, policy, or CI/tooling changes._

| Area | Current state (one line each) |
|------|-------------------------------|
| **Play** | 10 scenarios; solve mode + precinct; **four wire NPCs** + optional **HTTP LLM** wire copy (`POD_NOIR_LLM_CONTACT_WIRE`); REPL shortcuts & solve hints. |
| **Build / CI** | `lint` → `test` → `integration` (kind + digest-pinned node + `playtest-smoke.sh` with `CI=true`); **`make test-session`** for fast **`internal/session`** runs; session tests cover loop + golden **`help`**; optional **`core.hooksPath`** → **`githooks/`** (**pre-commit** + **pre-push** playtest); `make lint-docker` matches lint job. |
| **Docs** | [README](README.md), [AGENTS.md](AGENTS.md), [playtest checklist](docs/playtest-checklist.md), [northstar](docs/pod-noir-northstar.md), [architecture decisions](docs/architecture-decisions.md); Cursor rules: README + progress + northstar sync. |

## Log by month

| Month | File |
|-------|------|
| 2026-04 | [docs/progress/2026-04.md](docs/progress/2026-04.md) |

_Add a new row and new `docs/progress/YYYY-MM.md` when the calendar month rolls over._

## How to extend

1. **Same month:** append short bullets to `docs/progress/YYYY-MM.md`; adjust the **Snapshot** table here if the headline story changes.
2. **New month:** create `docs/progress/YYYY-MM.md` from a few bullets; add one table row above; trim or generalize **Snapshot** rows if they get stale.
3. Do **not** paste long specs, full command dumps, or multi-paragraph design here—link to PRs, issues, or `docs/*.md` instead.

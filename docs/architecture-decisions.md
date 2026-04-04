# Architecture & behavior decisions (pod-noir)

Append-only log of **sweeping** choices: solve policy, precinct rules, session/CI contracts, hooks, LLM wiring. Routine work lives in **`docs/progress/`** and **[PROGRESS.md](../PROGRESS.md)**.

| Date | ID | Area | Decision | Notes |
|------|-----|------|----------|--------|
| 2026-04-02 | AD-001 | Git hooks | `git config core.hooksPath githooks` (tracked `githooks/pre-commit` + `pre-push`) | Same playtest smoke script; `SKIP_PLAYTEST_SMOKE` skips both; `git push --no-verify` bypasses push only. |
| 2026-04-02 | AD-002 | Playtest / dev | `make playtest-smoke-ci` runs in Docker Compose (`playtest-smoke-ci` service); dev image includes `kubectl` | Parity with CI `CI=true` smoke without host Go; Linux binary built inside container. |
| 2026-04-02 | AD-003 | CI | Integration job runs `scripts/playtest-smoke.sh` with `CI=true` after `podnoir version` | Single doctor + kubectl report; no duplicate standalone `doctor` step. |
| 2026-04-02 | AD-004 | Contacts | Four NPCs; bare `hint` = roster; `hint senior` / sysadmin / network / archivist take the call | Unlocks: Senior = logs+trace or non-HOT accuse; Sysadmin = examine pod; Network = trace; Archivist = dossier. Per-scenario wire copy in `wire_messages.go`. One delivered hint per contact per case. |
| 2026-04-02 | AD-005 | LLM | `ContactWirer` / `HTTP.ContactWire` — LLM-generated wire messages with static anchor | `POD_NOIR_LLM_CONTACT_WIRE`; fallback to `contacts.StaticWireMessage` on mock, disabled, or HTTP error (when fallback enabled). |
| 2026-04-04 | AD-006 | Session / tests | `kubectl.Kube` interface — `*kubectl.Runner` implements it; session holds `kubectl.Kube` for REPL + debrief victory checks | Help text embedded from `internal/session/testdata/repl_help.golden`; refresh with `POD_NOIR_UPDATE_REPL_HELP_GOLDEN=1`. Unit tests use a `fakeKube` test double for observe/trace without a cluster. |

Add a row when you change **cross-cutting behavior**; link PRs or issues in **Notes** when helpful. See **`.cursor/rules/northstar-sync.mdc`**.

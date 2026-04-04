# Git hooks (version-controlled)

This directory is **committed with the repo**. Run **`make git-hooks`** or **`./scripts/setup-git-hooks.sh`** once per clone to set **`git config core.hooksPath githooks`** so Git executes these hooks.

| Hook | Role |
|------|------|
| **`pre-commit`** | Run playtest smoke before each commit. |
| **`pre-push`** | Same smoke before **`git push`** — catches commits made with **`git commit --no-verify`** or **`SKIP_PLAYTEST_SMOKE=1`**. |

To skip the push hook only (emergency): **`git push --no-verify`**.

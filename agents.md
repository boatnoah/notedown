## Agent skills

### Issue tracker

Issues live in GitHub Issues. See `docs/agents/issue-tracker.md`.

### Pull requests

- Run `bin/check` from the repo root and confirm it passes before opening a PR.
- Title must follow Conventional Commits (`type(scope): summary`, ≤70 chars).
- Body must end with `Closes #<issue-number>` to auto-close the linked issue on merge.
- Squash and merge is enforced — the PR title is the squash commit message.
- Always request `boatnoah` as reviewer when opening a PR (`gh pr create --reviewer boatnoah`). This applies to agents only — Noah is exempt.

Full rules in `docs/agents/issue-tracker.md`.

### Triage labels

Three-state labels: `triage`, `ready`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context repo — one `CONTEXT.md` + `docs/adr/` at the root. See `docs/agents/domain.md`.

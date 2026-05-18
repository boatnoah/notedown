# Issue tracker: GitHub

Issues and PRDs for this repo live as GitHub issues. Use the `gh` CLI for all operations.

## Conventions

- **Create an issue**: `gh issue create --title "..." --body "..."`. Use a heredoc for multi-line bodies.
- **Read an issue**: `gh issue view <number> --comments`, filtering comments by `jq` and also fetching labels.
- **List issues**: `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'` with appropriate `--label` and `--state` filters.
- **Comment on an issue**: `gh issue comment <number> --body "..."`
- **Apply / remove labels**: `gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **Close**: `gh issue close <number> --comment "..."`

Infer the repo from `git remote -v` — `gh` does this automatically when run inside a clone.

## When a skill says "publish to the issue tracker"

Create a GitHub issue.

## When a skill says "fetch the relevant ticket"

Run `gh issue view <number> --comments`.

## Pull request rules

These rules apply whenever an agent opens or updates a PR.

**Title — Conventional Commits, ≤70 chars:**
```
<type>(<scope>): <summary>
```
Common types: `feat`, `fix`, `ci`, `docs`, `refactor`, `test`, `chore`. The title becomes the squash-merge commit message, so it must be accurate and follow this format exactly.

**Body — always include a closing reference:**
```
Closes #<issue-number>
```
Place it at the end of the PR body. GitHub will auto-close the linked issue on merge.

**Full body template:**
```markdown
## Summary
- <bullet>

## Test plan
- [ ] <item>

Closes #<issue-number>
```

**Squash and merge is enforced** — the PR title is the only commit message that lands on `main`. Keep it precise.

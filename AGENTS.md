# AGENTS.md

This repository is a downstream fork of upstream `patriceckhart/zot`.

## Project

`zot` is a lightweight terminal coding agent harness written in Go. This checkout is the `mi-skam` downstream version and is locally branded as **`zot-fork`**.

Keep the upstream module/package imports intact (`github.com/patriceckhart/zot/...`) unless a task explicitly requires changing them. The fork should stay source-compatible with upstream wherever possible.

## Branch model

- `origin/main` points to upstream `patriceckhart/zot` and is the clean upstream base.
- `fork` is the downstream/base branch for local development in `mi-skam/zot`.
- `fork` intentionally carries a very small local branding delta so local builds are easy to identify as `zot-fork`.

Current intended local branding includes:

- local build version defaults to `fork`
- welcome/help text identifies the build as `zot-fork`

Do not casually expand the local branding patch. Keep `fork` and upstream `main` as close as possible.

## Upstream PR workflow

Pull requests to upstream **must not** be based on the `fork` branch.

For upstream PRs:

1. Start from upstream `main` / `origin/main`, not `fork`.
2. Rebase onto current upstream `main` before PR preparation whenever possible.
3. Ensure the PR branch does not include local `zot-fork` branding commits.
4. Keep the PR minimal and focused on the upstreamable change.

Recommended flow:

```bash
git fetch origin
git switch main
git reset --hard origin/main
git switch -c feature/my-upstream-change
# work, test, commit
git fetch origin
git rebase origin/main
```

If work started from `fork` by accident, rebase or cherry-pick the upstreamable commits onto `origin/main` before opening a PR.

## Updating the downstream fork branch

Keep `fork` close to upstream:

```bash
git switch fork
git fetch origin
git rebase origin/main
git push --force-with-lease mi-skam fork
```

Resolve conflicts by preserving only the small local branding delta unless explicitly asked otherwise.

## Local development workflow

For local-only downstream changes, branch from `fork`:

```bash
git switch fork
git switch -c local/my-change
```

For upstreamable changes, branch from `origin/main` instead.

## Quality checks

Use the existing project commands:

```bash
go test ./...
make build
```

Run narrower package tests during iteration when appropriate, but run the full suite before preparing PRs or updating `fork`.

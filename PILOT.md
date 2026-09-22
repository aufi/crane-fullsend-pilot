# Fullsend pilot sandbox

This repository is a **throwaway sandbox** for evaluating
[Fullsend](https://github.com/fullsend-ai/fullsend) `v0.43.0` against a realistic
Go codebase. It is a content copy of [migtools/crane](https://github.com/migtools/crane)
taken on 2026-09-21, not a fork and not a maintained downstream.

## What this repository is not

- It is **not** upstream Crane. Nothing merged here reaches `migtools/crane`.
- It is **not** a release source. No binaries, images, or tags are published from here.
- It does **not** track upstream. It will drift and that is expected.

Report real Crane bugs and features upstream at
<https://github.com/migtools/crane/issues>.

## Pilot scope

Agent roles, security boundaries, validation steps, acceptance criteria, and
rollback are defined in the runbook that governs this pilot:
`fullsend-ai/FULLSEND_V0.43.0_RUNBOOK.md` in
<https://github.com/aufi/move-crane>.

Rules that apply to every agent run here:

1. Never merge to `main` and never bypass branch protection or human review.
2. Never modify `.github/workflows/**`, `.fullsend/**`, `.github/CODEOWNERS`,
   `AGENTS.md`, or `PILOT.md` without human review.
3. Never open a pull request against `migtools/crane` or any other repository.
4. Never put credentials, tokens, kubeconfigs, or cluster output into issues,
   pull requests, commits, or comments.
5. Keep changes small and scoped to the issue that triggered the run.

## Differences from upstream

- Inherited GitHub Actions workflows were pruned to `go.yml`, `lint.yml`, and
  `pr-title-check.yml`. E2E, nightly, release, Slack, cherry-pick, and issue
  automation workflows were removed so pilot runs stay cheap and readable.
- `.github/CODEOWNERS` was added to protect the pilot's own configuration.
- This file and a banner in `AGENTS.md` were added.

## Teardown

When the pilot ends, delete this repository. Follow section 13 of the runbook
to remove the GitHub Apps, secrets, and WIF access first.

<!-- merge gate verification -->

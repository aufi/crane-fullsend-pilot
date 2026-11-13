# EnMaaS migration and activation

## Configuration

Prepared November 13, 2026 for `aufi/crane-fullsend-pilot`:

| Setting | Value |
| --- | --- |
| Fullsend workflow and CLI | v0.55.0 |
| Fleet harness revision | `3b40ac718d29c3190b71e68447b9f4e938f19974`, agents v0.55.0 |
| Runtime | pi, version 1.5.0 in the pinned fleet images |
| Inference gateway extension | 0.2.0 in the pinned fleet images |
| OpenShell expected by CI | 0.1.4 |
| Gateway | `https://api.enmaas.devshift.net` |
| Model ID | `rits/deepseek-ai/DeepSeek-V4.1-Flash-1017` |
| Fullsend model | `gateway/rits/deepseek-ai/DeepSeek-V4.1-Flash-1017` |
| API | OpenAI Chat Completions |
| Authentication | API key, Authorization Bearer |
| Repository secret | `FULLSEND_INFERENCE_GATEWAY_API_KEY` |

The secret was provisioned from the operator's existing environment. No key
value belongs in configuration, logs, issues, or pull requests.

The source URLs pin the fleet harnesses with both a commit and SHA-256 hashes.
Those harnesses pin sandbox images by digest. The local triage harness inherits
the pinned fleet triage harness and disables automatic code promotion.

The legacy scalar `tracker` and `create_issues` configuration is migrated to
`tracker.github.allowed_targets`. The prior explicit cross-repository target,
`fullsend-ai/fullsend`, is preserved. The current repository is implicitly
allowed; this migration adds no other issue targets.

## Activation

Triage is enabled. It runs on an authorized human `/fs-triage` comment,
not automatically on newly opened or edited issues. Triage-to-code promotion
is disabled in both runner and sandbox settings.

Review, code, and fix are now enabled with pinned harnesses and the DeepSeek
model. Their fleet triggers apply, including automatic PR reviews and eligible
review-triggered fixes. Retro and prioritize are explicitly disabled. Each of
the ten review personas in this fleet revision has an explicit gateway model
override, because `subagents.default` does not replace a persona's own model.

Configuration is loaded from the trusted base branch. The original migration
was merged as PR #16. Further activation changes take effect after human review
and merge. Automatic triage-to-code promotion remains off.

## Completed triage validation

The [first complete triage run](https://github.com/aufi/crane-fullsend-pilot/actions/runs/43444515213)
on issue #10 passed against merged commit
`737eadd00499192af93b5d98045c5d36a90b05a3`. The job confirmed Fullsend 0.55.0,
OpenShell 0.1.4, pi, and the exact DeepSeek gateway model. The agent completed
in 48.8 seconds with seven tool calls and schema-valid output on its first
iteration. The post-script published the answer, the run-scoped gateway
provider was deleted, and no code job was promoted. One unsupported `gh` JSON
field produced a recoverable tool error.

Runtime-reported estimated cost was USD 0.011671436. The transcript identifies
that estimate as catalog-based, not a provider-confirmed EnMaaS charge.

## Checks performed before merge

- The authenticated gateway model catalog returned HTTP 200 and included the
  exact selected DeepSeek ID.
- Direct streaming Chat Completions succeeded with `reasoning_effort: high`,
  function-call arguments, a tool result, and a subsequent user turn. Responses
  identified the selected DeepSeek model; a reasoning field was observed.
- pi 1.5.0 with pi-inference-gateway 0.2.0 successfully used a read-only `read`
  tool and returned the expected file contents through this gateway. This was
  an isolated runtime test, not an OpenShell or GitHub Actions agent run.
- `fullsend lint --forge github --format json` resolved all four configured
  agent harnesses while enabled for inspection: zero errors and nine warnings.
  Warnings concern fleet harness environment delivery, deprecated issue URL
  variables, and the review harness's deprecated forge block.
- `fullsend inference gateway status aufi/crane-fullsend-pilot` resolved the
  intended URL, model, authentication mode, and local credential presence.
  In API-key mode this command does not make a live inference request.

The complete triage run subsequently verified GitHub runner connectivity and
the OpenShell gateway credential route. Review delegation and the code/fix
lifecycle still require their own live validation after activation.

Context/output limits, reasoning-effort behavior beyond the tested requests,
quotas, and the pilot's cost ceiling still need confirmation with the EnMaaS
operator. Model metadata and costs inherited by pi from its catalog are not
confirmation of EnMaaS limits or billing rates.

## Review, code, and fix validation after human merge

1. Run a controlled review of the small existing PR #15 using `/fs-review`.
   Confirm each registered persona's resolved gateway model, actual child tool
   execution, and a final structured result. This command intentionally posts
   the agent's review and can trigger an eligible fix when changes are requested.
2. Create a focused pilot issue for `HasFilesRecursively` in `e2e-tests/utils`.
   Its current substring-based check can mistake a real filename containing
   `(no files)` for an empty directory. Ask the code agent to correct the helper
   and add cluster-free regression coverage using temporary directories.
3. Trigger `/fs-code` on that issue. Confirm the selected model, meaningful
   implementation and tests, a validated result, and a PR targeting this pilot.
4. Use `/fs-fix` on the new PR with a narrow follow-up, such as an additional
   nested-directory regression case. Verify that the fix agent reads CI,
   publishes the requested change, and leaves the resulting PR checks passing.
5. Record actual run URLs, tested commits, models, outcomes, errors, and usage.
   Keep automatic review/fix runs distinct from manual verification runs.

These checks validate the agent lifecycle without a live Kubernetes cluster.
Live E2E execution remains a separate step in the E2E pilot assignment.

Keep the existing `FULLSEND_OPENAI_API_KEY` secret during migration. After the
EnMaaS validation succeeds, remove it and revoke its key if it is dedicated to
this pilot. Its presence does not select OpenAI for an explicitly configured
gateway model.

## Rollback

Before activation, closing the configuration PR leaves the existing runtime
unchanged. The gateway secret alone does not change inference routing.

After activation, revert the workflow and configuration changes together
through a reviewed PR. The old tracker schema requires the old Fullsend
workflow version too. The retained OpenAI secret supports rollback only while
its key remains valid. After removing or revoking it, restoring OpenAI requires
fresh credentials.

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

## Initial activation

Only triage is enabled. It runs on an authorized human `/fs-triage` comment,
not automatically on newly opened or edited issues. Triage-to-code promotion
is disabled in both runner and sandbox settings.

Review, code, and fix have pinned harnesses and the DeepSeek model configured,
but `enabled: false`. Retro and prioritize are explicitly disabled. Each of
the ten review personas in this fleet revision has an explicit gateway model
override, because `subagents.default` does not replace a persona's own model.

Configuration is loaded from the trusted base branch. The migration becomes
active only after its configuration PR is reviewed and merged by a human.
The `fullsend/scaffold-install` branch name prevents the old base-branch
workflow from automatically reviewing the migration with OpenAI.

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
  variables, and the review harness's deprecated forge block. The final
  configuration disables review, code, and fix again.
- `fullsend inference gateway status aufi/crane-fullsend-pilot` resolved the
  intended URL, model, authentication mode, and local credential presence.
  In API-key mode this command does not make a live inference request.

No full Fullsend/OpenShell agent run or GitHub runner connectivity test has
been completed for this configuration yet. The operator's local OpenShell was
0.1.2 during preparation, older than the 0.1.4 version pinned by this release.
The isolated pi test does not validate sandbox networking or credential
placeholder handling.

Context/output limits, reasoning-effort behavior beyond the tested requests,
quotas, and the pilot's cost ceiling still need confirmation with the EnMaaS
operator. Model metadata and costs inherited by pi from its catalog are not
confirmation of EnMaaS limits or billing rates.

## Validation after human merge

1. Use a small pilot issue and add `/fs-triage` as a human with repository
   write access. This intentionally runs the triage post-script, which may
   update the issue's labels and triage comment.
2. Inspect the Actions run. Confirm Fullsend 0.55.0, pi, the exact gateway
   model, the pinned sandbox image, successful tool execution, and a validated
   triage result. Confirm that no code run was promoted automatically.
3. Check token usage and EnMaaS accounting. Record the run URL and outcome.
4. Enable review through a reviewed config change and run a controlled review.
   Check every registered persona's resolved model, completed result, and usage.
5. Enable and validate code, then fix, with small tasks. Follow the E2E pilot
   assignment for the first development tasks; live-cluster access is separate
   from this inference migration.

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

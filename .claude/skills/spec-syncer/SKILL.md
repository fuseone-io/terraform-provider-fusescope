---
name: spec-syncer
description: >-
  Catch up on drift in the FuseScope OpenAPI spec for every already-declared resource. Use this when the task is "the openapi spec changed", "regenerate the provider from the latest spec", "sync the generated code to fuseone-io/fusescope main", "pick up the new spec", or a periodic spec-drift catch-up. It does NOT add a new resource (use resource-creator for that) and makes **no** `skaff.yml` edits.
---

# Sync generated code to the FuseScope OpenAPI spec

The spec lives in `fuseone-io/fusescope` (`main` → `docs/api/openapi.yml`). This provider tracks it: periodically regenerate everything from the current spec and land the drift as a **catch-up PR**.

> **Blocked:** regeneration needs `skaff`, which is private to SigNoz and not available to this fork. Until a replacement generator is chosen, do Steps 1–2 only: measure the drift and record it in an issue in `fuseone-io/terraform-provider-fusescope`. Follow the `generated-code` rule for anything that cannot wait.

## Steps

### 1. Fetch the current spec and record its commit

```sh
mkdir -p tmp/spec
curl -fsSL https://raw.githubusercontent.com/fuseone-io/fusescope/main/docs/api/openapi.yml -o tmp/spec/openapi.yml
SPEC_SHA=$(gh api "repos/fuseone-io/fusescope/commits?path=docs/api/openapi.yml&per_page=1" --jq '.[0].sha')
echo "spec at fuseone-io/fusescope@$SPEC_SHA"
```

### 2. Measure the drift

For each resource declared in `skaff.yml`, compare its request and response schemas in the spec with the generated models under `internal/`. Note added, removed or retyped fields and changed enum values; that list is the body of the issue (or, once regeneration works, of the PR).

### 3. Regenerate (blocked)

When a generator is available: branch `chore/spec-sync-<date>`, regenerate every declared resource without touching `skaff.yml`, then run `go mod tidy` only if `go.mod` changed.

A wholesale regen legitimately pulls in unrelated drift (a shared schema gains a field). That is catch-up. A generated file **disappearing**, a resource's schema emptying out, or a schema the generator now skips is a breakage: stop and diagnose it against [`patterns.md`](../../docs/patterns.md) instead of committing a partial regen.

### 4. Gates, docs, commit

Run the [`go-checks`](../../rules/go-checks.md) rule, then the **docs-writer** skill. One catch-up commit with the regenerated code and docs:

```sh
git commit -am "chore(codegen): sync generated code to the FuseScope openapi spec

Regenerated against fuseone-io/fusescope@$SPEC_SHA."
```

### 5. Open the PR — confirm first

Summarize the drift and **confirm with the user before `gh pr create`**. Record `fuseone-io/fusescope@<sha>` in the PR body and follow the [`pull-requests`](../../rules/pull-requests.md) rule.

## Not this skill

- Adding a new resource → **resource-creator**.
- A dependency bump unrelated to the spec → a plain `build(deps)` PR.

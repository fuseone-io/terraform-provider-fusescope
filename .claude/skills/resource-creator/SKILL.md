---
name: resource-creator
description: >-
  End-to-end workflow for adding a new Terraform resource + data source to terraform-provider-fusescope via the skaff OpenAPI codegen pipeline. Use this whenever the task is to "add", "create", "generate", "scaffold", or "wire up" a new `signoz_<name>` resource/data source from a FuseScope API endpoint — even if the user just names the API (e.g. "dashboard v2 APIs", "downtime schedules", "service accounts") without saying "skaff" or "codegen". It covers declaring the resource in skaff.yml, running the skaff generators in the right order with the right flags, registering the resource in the live provider, generating examples + docs, the build/lint gates, and the two-commit split. Reach for it before hand-writing any resource — most resources are field-for-field and fully generated; a skip is a schema gap to fix upstream, never a convertor to hand-write.
---

# Adding a resource to terraform-provider-fusescope

This provider is **generated**, not hand-written. A tool called `skaff` reads the FuseScope OpenAPI spec plus a per-resource config (`skaff.yml`) and emits the `zz_generated_*` files across `internal/`. Adding a resource means *declaring* it and *running the pipeline* — you almost never write resource logic by hand.

> **Blocked:** `skaff` is private to SigNoz and not available to this fork, so Step 2 cannot run. Declare the resource (Step 1) and record the request in an issue in `fuseone-io/terraform-provider-fusescope`; follow the `generated-code` rule for anything that cannot wait.

Many resources already shipped this way. Follow their shape; don't invent a new one.

## Mental model: the layers

@.claude/docs/architecture.md

## Inputs (where things live)

Two inputs:

- **Spec:** the FuseScope OpenAPI spec (`fuseone-io/fusescope`, `main`, `docs/api/openapi.yml`). Fetch a fresh copy — never a stale in-repo copy:

```sh
mkdir -p tmp/spec
curl -fsSL https://raw.githubusercontent.com/fuseone-io/fusescope/main/docs/api/openapi.yml -o tmp/spec/openapi.yml
```

If you need a spec that isn't plain `main` — e.g. an open PR's contract merged in — hand-produce the merged spec into that same path, `tmp/spec/openapi.yml`.

- **Config:** `skaff.yml` at this repo root (you edit this in Step 1).

`skaff` is not available to this fork (see the note at the top).

## Step 1 — Declare in `skaff.yml`

Add one block under `resources:` (create/read/update/delete: each a `path` + `method`) and one under `data_sources:` (read only). Copy the nearest existing block. The path template uses `{id}`. Example (collection):

```yaml
resources:
  <name>:
    create: { path: /api/v1/<collection>, method: POST }
    read:   { path: /api/v1/<collection>/{id}, method: GET }
    update: { path: /api/v1/<collection>/{id}, method: PUT }
    delete: { path: /api/v1/<collection>/{id}, method: DELETE }
    schema:
      ignores:
        - data        # response envelope
        - status      # also drops a *meaningful* status field if the resource has one (accepted gap)
        - createdAt
        - createdBy
        - updatedAt
        - updatedBy
data_sources:
  <name>:
    read: { path: /api/v1/<collection>/{id}, method: GET }
    schema:
      ignores: [ data, status, createdAt, createdBy, updatedAt, updatedBy ]
```

**`schema.ignores` rules:**
- Drops the response envelope + server audit fields — see [`audit-fields.md`](../../docs/audit-fields.md) for why.
- To hide a nested attribute, add its **OpenAPI property name** to `ignores`. This keeps the model scalar and stops `skaff types` emitting custom types for it (service_account ignores `serviceAccountRoles`; user ignores `userRoles` and `frontendBaseUrl`).
- Ignoring everything nested → an all-scalar resource (route_policy, service_account-after-ignore) needs **no** customtypes and no per-customtype helpers — it still gets a `flex` entry point.

## Step 2 — Run the pipeline

Blocked: the pipeline needs `skaff`, which this fork does not have. When a replacement generator is chosen, this step becomes its runbook. Until then, stop after Step 1 and open the issue.

## Step 3 — Register in the live provider

Edit `internal/provider/provider.go`:
- add `services.New<Name>Resource` to `Resources()`
- add `services.New<Name>DataSource` to `DataSources()`

Keep the `services.*` entries grouped (alphabetical), before any hand-written `signoz.*` entries. The generated services set `TypeName = ProviderTypeName + "_<name>"`, so the Terraform type is `signoz_<name>`. Registration must happen before docs (Step 4) — tfplugindocs introspects the live provider schema.

## Step 4 — Examples + docs

Run the **docs-writer** skill.

## Step 5 — Gates

Run `go mod tidy` + `go mod vendor` **only if a new dependency appeared** (the first nested-object resource pulls in `terraform-plugin-go`). Then run the Go gates — see the [`go-checks`](../../rules/go-checks.md) rule.

## Step 6 — Integration tests

## Step 7 — Commit (three-commit split)

On a `feature/<name>` branch, split the work in two commits:
1. **regen output:** `skaff.yml` + the `internal/**/zz_generated_*` files
2. **wiring + docs:** `internal/provider/provider.go`, `examples/.../signoz_<name>/**`, `docs/**/<name>.md`
3. **integration tests:** `tests/`

Then open a PR following the **pull-requests** rule (`.claude/rules/pull-requests.md`) — it owns the PR template and body format.

## When skaff skips a schema — fix the schema, not skaff

Run `skaff types`; for each skipped schema it prints the name + failing OAS path + reason. Log it, fix it in the FuseScope schema (`fuseone-io/fusescope`, `docs/api/openapi.yml`), regenerate, and repeat until `skipped 0`.

The codegen references live in `.claude/docs/` — the catalog of known patterns and their fixes is in [`patterns.md`](../../docs/patterns.md). Read it before starting a complex resource, and recognize the shape there before hand-fixing.

# Integration tests

End-to-end tests that drive Terraform against a running FuseScope backend and a locally built provider, asserting there is no drift. Two suites run:

- [`test_examples.py`](integration/tests/test_examples.py) — every example in [`examples/`](../examples) through a full Terraform CRUD cycle.
- [`test_testdata.py`](integration/tests/test_testdata.py) — every scenario in [`testdata/`](testdata) through a full lifecycle, including edits (see [Test data](#test-data)).

## Requirements

- A running FuseScope backend (the suite does not start one)
- Go (to build the provider under test)
- Python ≥ 3.11 with [`uv`](https://docs.astral.sh/uv/)

The Terraform CLI is *not* a requirement — the suite downloads it (see [CLI under test](#cli-under-test)).

## Backend

Start FuseScope with the `ghcr.io/fuseone-io/fusescope` image (or from source) and a root
user provisioned on first boot:

```sh
SIGNOZ_USER_ROOT_ENABLED=true
SIGNOZ_USER_ROOT_EMAIL=admin@integration.test
SIGNOZ_USER_ROOT_PASSWORD='password123Z$'
SIGNOZ_SQLSTORE_SQLITE_TRANSACTION__MODE=immediate
```

The suite logs in as that root user, creates a `terraform-integration` service account
with the `signoz-admin` role and runs Terraform with its API key. The service account is
deleted at the end of the run.

| variable | default | meaning |
|----------|---------|---------|
| `FUSESCOPE_ENDPOINT` | — (required; tests skip without it) | backend root URL, e.g. `http://localhost:8080` |
| `FUSESCOPE_ACCESS_TOKEN` | — | use this API key instead of minting a service account |
| `FUSESCOPE_ROOT_EMAIL` | `admin@integration.test` | root user to log in as |
| `FUSESCOPE_ROOT_PASSWORD` | `password123Z$` | its password |

## Running

```sh
cd tests
uv sync
FUSESCOPE_ENDPOINT=http://localhost:8080 uv run pytest -vv
```

## CLI under test

The suite drives either Terraform or OpenTofu. It downloads the binary itself rather than taking one from `PATH`, so a run pins exactly what it tested:

| flag | default | meaning |
|------|---------|---------|
| `--tool` | `terraform` | `terraform` or `opentofu` |
| `--tool-version` | `latest` | an exact version (`1.9.8`), or `latest` to resolve the newest release |
| `--download-path` | `tmp/bin/` at the repo root | where the binary lands, under `<tool>/<version>/` |

`--tool-version` rather than `--version` because pytest already owns that flag.

```sh
uv run pytest integration/tests --tool opentofu --tool-version 1.8.5
```

A download is skipped when the binary is already present, so repeat runs are free. Both CLIs are driven through the same `dev_overrides` CLI config — they read it from the same `TF_CLI_CONFIG_FILE` variable and normalize the `fuseone-io/fusescope` source address against their own default registry host, so nothing in the workspace changes between them.

CI does not run this suite yet: it needs a FuseScope backend. Run it locally against one.

## Examples

Every config under [`examples/`](../examples) runs automatically — `test_examples.py` drives each example through a full create -> no-drift -> destroy cycle, **one file at a time**. So an example only passes if it can `terraform apply` on its own.

If an example depends on a resource it doesn't define (e.g. it references another resource by ID), pick one:

- **skip it** in the examples integration test, or
- **bundle the prerequisite into the same file** — define the dependency resource and the example's own resource together, so the single file applies cleanly.

Input variables (channel secrets, a service name) get fake values from `VARIABLES` in `test_examples.py`; URLs point at `example.com`, so nothing reaches a real Slack or PagerDuty. Each directory under [`examples/recipes/`](../examples/recipes) is applied as one configuration.

## Test data

`testdata/` holds Terraform configs that exercise the provider beyond the curated, user-facing [`examples/`](../examples). Same top-level layout — `resources/fusescope_<name>/` — so the suite can drive every config here.

Each two-digit directory under `fusescope_<name>/` is one scenario: a single base config plus, optionally, ordered JSON-patch edits. All scenarios are run by [`test_testdata.py`](integration/tests/test_testdata.py) through a full lifecycle.

```
fusescope_rule/00/
  01-<name>.tf            # base config, no edits

fusescope_rule/01/
  01-<name>.tf.json       # base config
  02-jsonpatch.json       # RFC 6902 patch applied on top of the base
  03-jsonpatch.json       # RFC 6902 patch applied on top of the previous state
```

The runner creates the base (plan shows a create, apply, re-plan is clean), applies each `NN-jsonpatch.json` in ascending order — re-plan must show changes, apply, re-plan must be clean — and finally destroys. A scenario with no patches is just create → no-drift → destroy; the full sequence is the end-to-end editing path: create, no-drift, edit, drift, converge, repeat, delete.

Naming:

- The base is the single `01-<name>.tf` or `01-<name>.tf.json` in the directory; the `<name>` says what the scenario exercises.
- A JSON Patch (RFC 6902) needs a JSON target, so any scenario with patches uses a **`.tf.json`** base (Terraform JSON syntax, read natively by Terraform). Patch `path`s are relative to the single resource's body (e.g. `/condition/thresholds/basic/spec/0/target`). A patch-free scenario may use a plain HCL `.tf` base.

A scenario that references another resource (e.g. a rule's `thresholds[*].channels`) must **bundle the prerequisite into the same config** — define a `fusescope_notification_channel` and reference it by `display_name` (rules and route policies resolve channels by display name). In a `.tf.json` base the prerequisite sits alongside the resource under test; patch paths stay relative to the resource matching the scenario directory (`fusescope_<name>`), which is how the runner picks it out.

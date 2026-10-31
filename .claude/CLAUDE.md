# terraform-provider-fusescope

A [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) provider for [FuseScope](https://github.com/fuseone-io/fusescope). FuseOne's independent fork of the SigNoz provider, under MPL-2.0 (see `NOTICE`); it does not track upstream.

Most of `internal/` is **generated** (`zz_generated_*`) from the FuseScope OpenAPI spec plus a per-resource config (`skaff.yml`). The generator, `skaff`, is private to SigNoz and is **not available to this fork**. Until a replacement generator is chosen, see the `generated-code` rule before changing anything under `internal/`.

The provider's type name, resource names (`signoz_<name>`), env vars (`SIGNOZ_ENDPOINT`, `SIGNOZ_ACCESS_TOKEN`) and the `SIGNOZ-API-KEY` header still carry upstream names. They are user-facing contracts; rename them only in a dedicated change.

## Architecture

@.claude/docs/architecture.md

## Conventions

@.claude/rules

## Skills

| Task | Skill |
|------|-------|
| Add / scaffold a new resource + data source from a FuseScope API | **resource-creator** |
| Catch up on drift in the FuseScope OpenAPI spec | **spec-syncer** |
| Regenerate or validate registry docs, add examples, write a guide | **docs-writer** |
| Run the e2e suite, add a testdata scenario | **integration-tester** |
| Triage a reported bug: reproduce it, find which repo owns the defect, prove it with a test | **issue-triager** |

Reach for the skill before hand-writing: it carries the runbook, the flags, the footguns, and the design/history references.

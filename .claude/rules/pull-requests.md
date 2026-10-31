# Opening pull requests

Open PRs with the `gh` CLI against `fuseone-io/terraform-provider-fusescope`. Never open issues or PRs in SigNoz repositories.

- **Run the checks first.** For a Go change, run the `go-checks` rule; for docs/examples, the `docs-writer` gates. Don't open a PR on red.
- **Follow the template** (`.github/pull_request_template.md`): put each change under the right heading (Features / Fixes / Chores / Refactors / Tests). Don't add sections the template doesn't have.
- **Keep only the headings that apply.** Delete every heading that has nothing under it, along with its `<Add ...>` placeholder. The body must never contain an empty heading.
- **Keep the description concise and human-readable.** A few plain bullets saying what changed and why, for a reviewer skimming it — not a wall of text, not a restatement of the diff, not generated boilerplate.
- **Conventional commits**, scoped and imperative: `feat(dashboard): …`, `fix(rule): …`. AI-assisted commits end with a `Co-Authored-By:` trailer naming the model, as in the other FuseScope repositories.
- **License:** this repository is MPL-2.0. Never copy its code into FuseScope's MIT or Apache-2.0 repositories.

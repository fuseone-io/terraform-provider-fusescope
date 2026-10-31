---
paths:
  - "**/*.go"
---

# Go checks before opening a PR

For any change touching Go, run the checks the `ci` workflow runs **before opening a PR**. All must pass.

```sh
gofmt -s -l $(git ls-files '*.go')   # must print nothing
go mod tidy -diff && go mod verify
go test ./...
go build ./...
golangci-lint run                    # v2.12.2, the version pinned in .github/workflows/ci.yml
```

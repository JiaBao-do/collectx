# Contributing

1. Enable the pre-push hook once: `git config core.hooksPath .githooks`. It runs gofmt, `go mod tidy`,
   `go vet`, `go test -race -shuffle=on`, `go build` and golangci-lint. Never bypass it.
2. Every change ships with tests (property tests against a naive model where possible).
3. Go floor is 1.24: do not use APIs added later (`sync.WaitGroup.Go`, `errors.AsType`, ...).
4. Zero third-party dependencies, no cgo, no `unsafe`.
5. Conventional Commits; add a line to `CHANGELOG.md` under Unreleased.

# collectx

Generic Guava-style collections. Go library, module `github.com/JiaBao-do/collectx`.
Fills the gap left by Guava / Eclipse Collections (Multimap, Multiset, BiMap, Table, RangeSet, RangeMap) in Go.

## Commands
- Enable hook once: `git config core.hooksPath .githooks`
- Test: `go test -race -shuffle=on ./...`
- Lint: `golangci-lint run`
- Bench: `go test -run=^$ -bench=. -benchmem ./...`
- Fuzz: `go test -run=^$ -fuzz=FuzzRangeSet -fuzztime=30s`
- Vuln: `govulncheck ./...`

## Architecture
Single package `collectx` at repo root, one file per type: multiset.go, multimap.go (List/Set), bimap.go, table.go,
range.go (Range + cuts), rangeset.go, rangemap.go, sync.go (Synchronized wrapper). `examples/` holds runnable programs.

## Conventions
- Go 1.24 floor. No API newer than 1.24 (no WaitGroup.Go, no errors.AsType). CI runs 1.24 and stable.
- Standard library only. No cgo, no unsafe, no init(), no global mutable state.
- Every exported identifier has a doc comment stating complexity; every type states goroutine-safety.
- Property tests against naive models, fuzz tests, Example* tests.
- Conventional Commits; update CHANGELOG.md.

## Invariants (do not break)
- Types are not goroutine-safe; only Synchronized[T] wraps with a lock.
- Zero values are usable. Methods never panic on caller input; range constructors return ErrInvalidRange.
- RangeSet holds sorted, disjoint, non-connected, non-empty ranges. RangeMap holds sorted disjoint ranges.
- Guava semantics: ranges are over a continuous domain, so [1,3] and [4,6] are NOT connected.

## Roadmap
Immutable views, sorted (tree) multimap keys, SubRangeSet views, LRU excluded (use hashicorp/golang-lru or otter).

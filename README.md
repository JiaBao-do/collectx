# collectx

[![ci](https://github.com/JiaBao-do/collectx/actions/workflows/ci.yml/badge.svg)](https://github.com/JiaBao-do/collectx/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/JiaBao-do/collectx.svg)](https://pkg.go.dev/github.com/JiaBao-do/collectx)

Guava-style generic collections for Go: **ListMultimap, SetMultimap, Multiset, BiMap, Table, RangeSet, RangeMap**.
Standard library only, no cgo, `iter.Seq` iteration, JSON support, documented complexity.

**Requires Go 1.24+.**

```sh
go get github.com/JiaBao-do/collectx
```

## Quick start

This is `examples/quickstart/main.go`, run in CI.

```go
package main

import (
	"fmt"
	"slices"

	"github.com/JiaBao-do/collectx"
)

func main() {
	tags := collectx.NewSetMultimap[string, string]()
	tags.Put("post-1", "go")
	tags.Put("post-1", "generics")
	tags.Put("post-1", "go") // duplicate pair: ignored
	tags.Put("post-2", "go")
	fmt.Println(tags.Len(), slices.Sorted(tags.Get("post-1")))

	codes := collectx.NewBiMap[string, int]()
	_ = codes.Put("ok", 200)
	_ = codes.Put("missing", 404)
	if err := codes.Put("gone", 404); err != nil { // value already used
		fmt.Println("error:", err)
	}
	name, _ := codes.Inverse().Get(404)
	fmt.Println(name)
}
```

## Types

| Type | Guava equivalent | Notes |
|------|------------------|-------|
| `ListMultimap[K,V]` | ArrayListMultimap | ordered values, duplicates allowed; `Get` copies |
| `SetMultimap[K,V]` | HashMultimap | each (key,value) pair at most once |
| `Multiset[E]` | HashMultiset | counts, algebra (union, intersection, sum, difference) |
| `BiMap[K,V]` | HashBiMap | `Put` errors on duplicate value, `ForcePut` evicts, `Inverse` is a live view |
| `Table[R,C,V]` | HashBasedTable | sparse grid, `Row`, `Column`, `Transpose` |
| `Range[T]`, `RangeSet[T]` | Range, TreeRangeSet | open/closed/unbounded bounds, coalescing, `Complement` |
| `RangeMap[K,V]` | TreeRangeMap | `Put` overwrites and splits, `PutCoalescing` merges equal neighbours |
| `Synchronized[T]` | (none) | RWMutex wrapper for any of the above |

Examples (each run and compared with `expected_output.txt` by `go test ./examples`): `examples/quickstart`,
`examples/table`, `examples/ranges`, `examples/multiset`, and `examples/demo` (inverted index plus room booking).

## Things to care about

Full list with wrong/right snippets, each backed by a test: [docs/PITFALLS.md](docs/PITFALLS.md).

- **Not goroutine-safe.** Guard with your own lock or wrap in `Synchronized[T]`.
- **Iteration order of hash-based types is unspecified.** Sort before printing or comparing. Do not modify a collection while ranging over it.
- **Copies vs views.** `ListMultimap.Get` returns a copy; `BiMap.Inverse` is a live view; `Clone` is independent.
- **Continuous domain (Guava semantics).** `[1,3]` and `[4,6]` are not connected, even for ints; `[1,4)` and `[4,6)` are.
- **Errors, not exceptions.** Invalid ranges return `ErrInvalidRange`; a taken BiMap value returns `ErrValueAlreadyPresent`.
- **Keys.** NaN never equals itself (cannot be found or removed); pointer keys compare by address; `any` keys holding a slice or map panic.
- **Range collections are slice-backed.** Lookups are O(log n); Add/Put/Remove shift the slice, O(n) worst case.
- **Go floor 1.24.**

## Complexity

Multimaps, Multiset, BiMap, Table point operations: O(1) average. `Table.Column`: O(rows); `ColumnKeys`, `Transpose`: O(cells).
RangeSet/RangeMap: `Contains`/`Get` O(log n), mutation O(log n) locate + O(n) shift, set algebra O(n+m).
Each method documents its cost on pkg.go.dev.

## Comparison (verified 2026-09-21 by reading the repositories)

| | collectx | [emirpasic/gods](https://github.com/emirpasic/gods) v2 | [samber/lo](https://github.com/samber/lo) | [abc-inc/goava](https://github.com/abc-inc/goava) |
|--|--|--|--|--|
| Multimap | yes (list, set) | not in the tree | no (helper functions, no collection types) | Guava port, last push 2023-03, 3 stars; features not verified |
| Multiset | yes | not in the tree | no | not verified |
| BiMap | yes | `hashbidimap`, `treebidimap` | no | not verified |
| Table | yes | not in the tree | no | not verified |
| RangeSet / RangeMap | yes | not in the tree | no | not verified |
| Generics | yes | yes (v2 module `github.com/emirpasic/gods/v2`, `go 1.21`) | yes | not verified |

Use **gods** for lists, queues, stacks, trees (red-black, AVL, B-tree), sorted maps/sets, and other structures collectx does not
have; use **lo** for slice/map/iterator helper functions; use `hashicorp/golang-lru` or `maypok86/otter` for caches;
`deckarep/golang-set` for plain sets. collectx does not replace them: it adds the Guava types they lack.

Measured on this repo's benchmarks (AMD Ryzen 7, Go 1.27, `go test -bench . -benchmem`): `RangeSet.Contains` at 100,000 ranges
143 ns vs 33,906 ns for a linear scan; at 10 ranges the linear scan is faster (34 ns vs 10 ns). `Multiset.Add` 26 ns vs 15 ns for a
plain `map[int]int` counter, both 0 allocs.

## Not included (v0.1)

Immutable/persistent collections, ordered (tree) multimaps, sub-range views, LRU/LFU caches, skip lists, Bloom filters
(maintained libraries already exist for those).

## License

MIT

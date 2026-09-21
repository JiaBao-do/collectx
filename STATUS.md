# collectx status

| stage | attempts | updated | evidence |
|-------|----------|---------|----------|
| validated | 0 | 2026-09-21 | Step 0 gate passed (PARTIAL): Table, RangeMap, RangeSet, Multiset, Multimap have no maintained, Guava-faithful, generic, iter.Seq option. Building those plus BiMap. |

## Step 0 coverage matrix (verified 2026-09-21 via gh api repo metadata, git trees, WebSearch; gh search API was rate-limited, so search-by-keyword used WebSearch)

| Type | Provided by (generic?) | Verdict |
|------|------------------------|---------|
| Multimap (list/set) | gods v2: none (tree has no multimap). lo: none. jwangsadinata/go-multimap: 38 stars, last push 2019, pre-generics. zyedidia/generic multimap: 1348 stars repo, last push 2024-01, generic, ordered-key focus, no iter.Seq. nkamenev/multimap: 0 stars, 2025-08, testify dep. fufuok/utils: 36 stars, bundle of utilities. | POORLY SERVED. Build |
| BiMap | gods v2 hashbidimap/treebidimap (generic, but no iter.Seq, go 1.21, only v2.0.0-alpha tag, last push 2025-03-12; latest non-alpha release v1.18.1 is 2022 non-generic). vishalkuo/bimap 47 stars 2025-08 (thread-safe only). jub0bs/bimap 0 stars 2022. zyedidia bimap. | Served in basics; build for Guava semantics (Put error on duplicate value, ForcePut, Inverse view) |
| Table (row/col/value) | gods: none. lo: none. zyedidia array2d is a dense grid, not Guava Table. No maintained generic library found. | GAP. Build |
| RangeSet | AndrewWPhillips/rangeset 2 stars (2025-04, int-focused); google/go-intervals 141 stars ARCHIVED; b97tsk/intervals 3 stars; brentp/rolyatmax small. gods/lo: none. | GAP (no Guava-faithful open/closed bound, complement, subRange). Build |
| RangeMap | none found (rdleal/intervalst 36 stars is an interval search tree, not a coalescing range map). | GAP. Build |
| Multiset | aadamandersson/multiset 0 stars 2023; godsadta small; soniakeys old. gods: none. lo: none (only word match). | GAP-ish (tiny/unmaintained). Build |
| Immutable / persistent | benbjohnson/immutable 744 stars, last push 2023-08 (List/Map/SortedMap, no iter.Seq). | STALE. Out of scope for v0.1 (Freeze/Copy value semantics for our types only) |
| Ordered map | elliotchance/orderedmap 1025 stars, active 2026-09. gods linkedhashmap/treemap. | COVERED. Skip |
| LRU/LFU cache | hashicorp/golang-lru 5123 stars 2026-09; maypok86/otter 2682 stars 2026-06. | COVERED. Skip |
| Skip list | huandu/skiplist 428 stars 2024-09; gods btree/rbtree/avl cover ordered. | COVERED enough. Skip |
| Bloom filter | bits-and-blooms/bloom 2812 stars 2026-07. | COVERED. Skip |
| Interner | stdlib `unique` package (Go 1.23). | COVERED. Skip |
| Generic set | deckarep/golang-set 4744 stars, active. | COVERED. Skip |

Unverified: keyword `gh search repos` and code search were rate-limited (HTTP 403) after the first burst; conclusions rely on WebSearch results, direct repo metadata and git trees of gods/lo/zyedidia. pkg.go.dev was reached only via WebSearch listings.

## Differentiator
Focused, zero-dependency module with Guava-faithful semantics: Multimap (list/set), Multiset, BiMap, Table, RangeSet, RangeMap; Go 1.24+ (iter.Seq/Seq2), documented complexity and goroutine-safety, JSON, property tests vs naive models.

## Version policy (coordinator correction)
go.mod `go 1.24`, CI matrix 1.24 + stable, no APIs newer than 1.24.

## Board
agentboard RUNNING.md does not exist; nothing synced.

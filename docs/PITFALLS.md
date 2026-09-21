# Things to care about

Every item is backed by a `TestPitfall*` test in `pitfalls_test.go`.

## 1. Not goroutine-safe

Wrong: two goroutines calling `Put` on one `SetMultimap` (data race, possible fatal "concurrent map writes").

```go
go m.Put("a", 1)
go m.Put("b", 2)
```

Right: wrap it.

```go
s := collectx.NewSynchronized(collectx.SetMultimap[string, int]{})
s.Update(func(m *collectx.SetMultimap[string, int]) { m.Put("a", 1) })
n := collectx.Read(s, func(m *collectx.SetMultimap[string, int]) int { return m.Len() })
```

Do not call `Update`/`View` from inside a callback on the same wrapper (not reentrant); consume iterators inside the callback.
(`TestPitfallSynchronizedIsTheRightWay`)

## 2. Iteration order is unspecified (hash-based types)

Wrong: printing `for k := range m.Keys()` and expecting stable output. Right: `slices.Sorted(m.Keys())`.
Ordered types (`RangeSet.Ranges`, `RangeMap.All`) iterate ascending; `ListMultimap` keeps insertion order per key.
Never modify a collection while ranging over it. (`TestPitfallSortForDeterministicOutput`)

## 3. Copies versus views

- `ListMultimap.Get` returns a copy: `m.Get(k)[0] = 9` does nothing to the map (Guava returns a live view).
- `BiMap.Inverse()` is a live view sharing storage: `inv.Put(...)` changes the original.
- `Clone()` is independent.

(`TestPitfallCopiesAndViews`)

## 4. Guava semantics you may not expect

- Continuous domain: `[1,3]` and `[4,6]` do not merge even for ints; `[1,4)` and `[4,6)` do. Use half-open ranges for integer data.
  (`TestPitfallContinuousDomain`)
- `RangeSet.Len()` is the number of ranges, not values. (`TestPitfallRangeSetLenCountsRanges`)
- Guava throws where collectx returns errors: invalid ranges give `ErrInvalidRange`, `BiMap.Put` of a taken value gives
  `ErrValueAlreadyPresent` (use `ForcePut`), `Range.Intersection` of disconnected ranges returns `ok=false`.
  (`TestPitfallErrorsInsteadOfExceptions`)
- No sub-range/sub-map live views, no immutable variants (v0.1). `Clone` for snapshots; `Range` values are immutable.
- `SetMultimap` and `Multiset` have no defined order (no LinkedHash variants).

## 5. Nil versus zero value

The zero value is ready to use (`var m collectx.Multiset[int]; m.Add(1, 1)`). A nil pointer is not: methods panic like any nil
dereference. (`TestPitfallNilVersusZeroValue`)

## 6. Key gotchas

- NaN as a float key never equals itself: it cannot be found or removed, only `Clear` frees it. (`TestPitfallNaNKeys`)
- Pointer keys compare by address, not by pointee. (`TestPitfallPointerKeysAreIdentity`)
- `any`/interface keys holding an uncomparable dynamic type (slice, map, func) panic at runtime, as with a Go map.
  (`TestPitfallUncomparableInterfaceKey`)

## 7. JSON

Decoding replaces the receiver's contents and leaves it untouched on error. Hash-based types encode in unspecified order: compare
with `Equal`, not bytes. Map-shaped formats need JSON object-key types (string, integers, TextMarshaler) for keys.
(`TestPitfallUnmarshalReplaces`, `TestPitfallJSONOrderUnspecified`)

## 8. Performance and memory

- `SetMultimap` allocates one inner map per key; with many keys of one value each, prefer `ListMultimap` or a `map[K]V`.
- `Table.Column` scans all rows (O(rows)); `ColumnKeys` and `Transpose` are O(cells). Put the hot axis on rows.
- `RangeSet`/`RangeMap` are slice-backed: lookups O(log n), inserts in the middle shift the slice (O(n)). Fine to hundreds of
  thousands of ranges when mostly read; for a heavy random-insert workload measure first. For tiny sets (about 10 ranges) a
  plain slice scan is faster (see README numbers).

## 9. Not supported

Concurrent-safe variants beyond `Synchronized`, immutable/persistent collections, sorted multimaps, sub-range views, caches.
Use `emirpasic/gods` (trees, lists, queues), `samber/lo` (helpers), `hashicorp/golang-lru` / `maypok86/otter` (caches).

## 10. Go version

Requires Go 1.24+ (`iter`, generics, `testing.B.Loop` in benchmarks). CI runs 1.24 and stable on Linux, macOS, Windows.

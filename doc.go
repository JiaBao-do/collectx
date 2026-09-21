// Package collectx provides generic, dependency-free collection types modelled
// on Google Guava: ListMultimap, SetMultimap, Multiset, BiMap, Table, RangeSet
// and RangeMap.
//
// # Goroutine safety
//
// None of the types are safe for concurrent use. Concurrent readers are fine as
// long as nobody writes; guard mutation with your own sync.RWMutex.
//
// # Zero values
//
// Every type is usable as its zero value (the first write allocates).
//
// # Iteration
//
// Iteration uses iter.Seq and iter.Seq2 (Go 1.23+). Iteration order of hash
// based types is unspecified. Do not mutate a collection while ranging over it
// unless the method documents that it is allowed.
//
// # Serialization
//
// Every type implements json.Marshaler and json.Unmarshaler. Formats are
// documented on each type.
//
// # Portability
//
// Standard library only, no cgo, no unsafe, no reflection beyond encoding/json,
// no init side effects.
package collectx

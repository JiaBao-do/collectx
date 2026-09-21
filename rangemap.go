package collectx

import (
	"cmp"
	"encoding/json"
	"iter"
	"slices"
	"sort"
)

// RangeEntry is one (range, value) pair of a RangeMap.
type RangeEntry[K cmp.Ordered, V comparable] struct {
	Range Range[K]
	Value V
}

// RangeMap maps disjoint ranges of keys to values, like Guava's TreeRangeMap.
// Put overwrites whatever overlapped, splitting existing entries at the edges
// of the new range. Ranges are never merged implicitly; use PutCoalescing to
// merge connected ranges that carry equal values.
//
// Get and GetEntry are O(log n) for n stored entries. Put, PutCoalescing and
// Remove are O(log n) to locate plus O(n) worst case to shift the backing
// slice. Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable RangeMap.
type RangeMap[K cmp.Ordered, V comparable] struct {
	es []RangeEntry[K, V]
}

// NewRangeMap returns an empty RangeMap.
func NewRangeMap[K cmp.Ordered, V comparable]() *RangeMap[K, V] { return &RangeMap[K, V]{} }

func (m *RangeMap[K, V]) firstHiAtLeast(c cut[K]) int {
	return sort.Search(len(m.es), func(i int) bool { return compareCuts(m.es[i].Range.hi, c) >= 0 })
}

// span returns the index range [i, j) of entries overlapping r (non-empty).
func (m *RangeMap[K, V]) span(r Range[K]) (i, j int) {
	i = sort.Search(len(m.es), func(i int) bool { return compareCuts(m.es[i].Range.hi, r.lo) > 0 })
	j = sort.Search(len(m.es), func(i int) bool { return compareCuts(m.es[i].Range.lo, r.hi) >= 0 })
	return i, max(i, j)
}

// splice replaces the entries overlapping r by their remainders outside r,
// plus mid when non-nil, and returns the index at which mid was placed.
func (m *RangeMap[K, V]) splice(r Range[K], mid *RangeEntry[K, V]) int {
	i, j := m.span(r)
	var repl [3]RangeEntry[K, V]
	n, at := 0, 0
	if i < j {
		if first := m.es[i]; compareCuts(first.Range.lo, r.lo) < 0 {
			repl[n] = RangeEntry[K, V]{Range[K]{first.Range.lo, r.lo}, first.Value}
			n++
		}
	}
	at = i + n
	if mid != nil {
		repl[n] = *mid
		n++
	}
	if i < j {
		if last := m.es[j-1]; compareCuts(r.hi, last.Range.hi) < 0 {
			repl[n] = RangeEntry[K, V]{Range[K]{r.hi, last.Range.hi}, last.Value}
			n++
		}
	}
	m.es = slices.Replace(m.es, i, j, repl[:n]...)
	return at
}

// Put maps every key of r to v, overwriting overlapping entries. Empty ranges
// are ignored.
func (m *RangeMap[K, V]) Put(r Range[K], v V) {
	if r.IsEmpty() {
		return
	}
	m.splice(r, &RangeEntry[K, V]{r, v})
}

// PutCoalescing is like Put but merges the new entry with connected neighbours
// holding an equal value.
func (m *RangeMap[K, V]) PutCoalescing(r Range[K], v V) {
	if r.IsEmpty() {
		return
	}
	at := m.splice(r, &RangeEntry[K, V]{r, v})
	lo, hi := at, at
	merged := r
	if at > 0 && m.es[at-1].Value == v && m.es[at-1].Range.IsConnected(r) {
		merged = merged.Span(m.es[at-1].Range)
		lo = at - 1
	}
	if at+1 < len(m.es) && m.es[at+1].Value == v && m.es[at+1].Range.IsConnected(r) {
		merged = merged.Span(m.es[at+1].Range)
		hi = at + 1
	}
	if lo != hi || lo != at {
		m.es = slices.Replace(m.es, lo, hi+1, RangeEntry[K, V]{merged, v})
	}
}

// Remove unmaps every key of r, splitting entries as needed.
func (m *RangeMap[K, V]) Remove(r Range[K]) {
	if r.IsEmpty() {
		return
	}
	m.splice(r, nil)
}

// GetEntry returns the range and value containing k. O(log n).
func (m *RangeMap[K, V]) GetEntry(k K) (Range[K], V, bool) {
	i := m.firstHiAtLeast(cut[K]{aboveValue, k})
	if i < len(m.es) && m.es[i].Range.Contains(k) {
		return m.es[i].Range, m.es[i].Value, true
	}
	var zero V
	return Range[K]{}, zero, false
}

// Get returns the value mapped for k. O(log n).
func (m *RangeMap[K, V]) Get(k K) (V, bool) {
	_, v, ok := m.GetEntry(k)
	return v, ok
}

// Span returns the smallest range enclosing all entries; ok is false if empty.
func (m *RangeMap[K, V]) Span() (Range[K], bool) {
	if len(m.es) == 0 {
		return Range[K]{}, false
	}
	return Range[K]{m.es[0].Range.lo, m.es[len(m.es)-1].Range.hi}, true
}

// Len returns the number of stored entries.
func (m *RangeMap[K, V]) Len() int { return len(m.es) }

// Clear removes everything.
func (m *RangeMap[K, V]) Clear() { m.es = nil }

// All iterates over (range, value) in ascending key order. The map must not
// be modified during iteration.
func (m *RangeMap[K, V]) All() iter.Seq2[Range[K], V] {
	return func(yield func(Range[K], V) bool) {
		for _, e := range m.es {
			if !yield(e.Range, e.Value) {
				return
			}
		}
	}
}

// Clone returns an independent copy.
func (m *RangeMap[K, V]) Clone() *RangeMap[K, V] { return &RangeMap[K, V]{es: slices.Clone(m.es)} }

// Equal reports whether both maps hold identical entries.
func (m *RangeMap[K, V]) Equal(o *RangeMap[K, V]) bool { return slices.Equal(m.es, o.es) }

type rangeEntryJSON[K cmp.Ordered, V comparable] struct {
	Range Range[K] `json:"range"`
	Value V        `json:"value"`
}

// MarshalJSON encodes an ascending array of {"range":...,"value":...}.
func (m *RangeMap[K, V]) MarshalJSON() ([]byte, error) {
	out := make([]rangeEntryJSON[K, V], len(m.es))
	for i, e := range m.es {
		out[i] = rangeEntryJSON[K, V](e)
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes the format of MarshalJSON. Later entries overwrite
// earlier ones where they overlap, as with Put.
func (m *RangeMap[K, V]) UnmarshalJSON(b []byte) error {
	var in []rangeEntryJSON[K, V]
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	fresh := RangeMap[K, V]{}
	for _, e := range in {
		fresh.Put(e.Range, e.Value)
	}
	*m = fresh
	return nil
}

package collectx

import (
	"cmp"
	"encoding/json"
	"iter"
	"slices"
	"sort"
)

// RangeSet is a set of values represented as a sorted list of disjoint,
// non-empty, non-connected ranges, like Guava's TreeRangeSet. Adding a range
// coalesces every connected range into one.
//
// Contains, RangeContaining, Encloses and Intersects are O(log n) for n stored
// ranges. Add and Remove are O(log n) to locate plus O(n) worst case to shift
// the backing slice (amortized fast when appending in order). Complement,
// Union, Intersection and Difference are O(n + m). Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable RangeSet.
type RangeSet[T cmp.Ordered] struct {
	rs []Range[T]
}

// NewRangeSet returns a RangeSet holding the given ranges.
func NewRangeSet[T cmp.Ordered](ranges ...Range[T]) *RangeSet[T] {
	s := &RangeSet[T]{}
	for _, r := range ranges {
		s.Add(r)
	}
	return s
}

// firstHiAtLeast returns the first index whose upper cut is >= c.
func (s *RangeSet[T]) firstHiAtLeast(c cut[T]) int {
	return sort.Search(len(s.rs), func(i int) bool { return compareCuts(s.rs[i].hi, c) >= 0 })
}

// firstHiAbove returns the first index whose upper cut is > c.
func (s *RangeSet[T]) firstHiAbove(c cut[T]) int {
	return sort.Search(len(s.rs), func(i int) bool { return compareCuts(s.rs[i].hi, c) > 0 })
}

// firstLoAbove returns the first index whose lower cut is > c.
func (s *RangeSet[T]) firstLoAbove(c cut[T]) int {
	return sort.Search(len(s.rs), func(i int) bool { return compareCuts(s.rs[i].lo, c) > 0 })
}

// firstLoAtLeast returns the first index whose lower cut is >= c.
func (s *RangeSet[T]) firstLoAtLeast(c cut[T]) int {
	return sort.Search(len(s.rs), func(i int) bool { return compareCuts(s.rs[i].lo, c) >= 0 })
}

// Add inserts every value of r. Empty ranges are ignored. Ranges connected to
// r are merged into it.
func (s *RangeSet[T]) Add(r Range[T]) {
	if r.IsEmpty() {
		return
	}
	i := s.firstHiAtLeast(r.lo)
	j := s.firstLoAbove(r.hi)
	merged := r
	if i < j {
		merged = Range[T]{minCut(r.lo, s.rs[i].lo), maxCut(r.hi, s.rs[j-1].hi)}
	}
	s.rs = slices.Replace(s.rs, i, max(i, j), merged)
}

// Remove deletes every value of r from the set, splitting ranges as needed.
func (s *RangeSet[T]) Remove(r Range[T]) {
	if r.IsEmpty() {
		return
	}
	i := s.firstHiAbove(r.lo)
	j := s.firstLoAtLeast(r.hi)
	if i >= j {
		return
	}
	var pieces [2]Range[T]
	n := 0
	if first := s.rs[i]; compareCuts(first.lo, r.lo) < 0 {
		pieces[n] = Range[T]{first.lo, r.lo}
		n++
	}
	if last := s.rs[j-1]; compareCuts(r.hi, last.hi) < 0 {
		pieces[n] = Range[T]{r.hi, last.hi}
		n++
	}
	s.rs = slices.Replace(s.rs, i, j, pieces[:n]...)
}

// AddAll adds every range of o.
func (s *RangeSet[T]) AddAll(o *RangeSet[T]) {
	for _, r := range o.rs {
		s.Add(r)
	}
}

// RemoveAll removes every range of o.
func (s *RangeSet[T]) RemoveAll(o *RangeSet[T]) {
	for _, r := range o.rs {
		s.Remove(r)
	}
}

// Contains reports whether v is in the set. O(log n).
func (s *RangeSet[T]) Contains(v T) bool {
	_, ok := s.RangeContaining(v)
	return ok
}

// RangeContaining returns the stored range containing v. O(log n).
func (s *RangeSet[T]) RangeContaining(v T) (Range[T], bool) {
	i := s.firstHiAtLeast(cut[T]{aboveValue, v})
	if i < len(s.rs) && s.rs[i].Contains(v) {
		return s.rs[i], true
	}
	return Range[T]{}, false
}

// Encloses reports whether every value of r is in the set (true for empty r).
func (s *RangeSet[T]) Encloses(r Range[T]) bool {
	if r.IsEmpty() {
		return true
	}
	i := s.firstHiAtLeast(r.hi)
	return i < len(s.rs) && s.rs[i].Encloses(r)
}

// Intersects reports whether the set has a value in common with r.
func (s *RangeSet[T]) Intersects(r Range[T]) bool {
	if r.IsEmpty() {
		return false
	}
	i := s.firstHiAbove(r.lo)
	return i < len(s.rs) && compareCuts(s.rs[i].lo, r.hi) < 0
}

// Span returns the smallest range enclosing the whole set; ok is false if the
// set is empty.
func (s *RangeSet[T]) Span() (Range[T], bool) {
	if len(s.rs) == 0 {
		return Range[T]{}, false
	}
	return Range[T]{s.rs[0].lo, s.rs[len(s.rs)-1].hi}, true
}

// Len returns the number of stored (coalesced) ranges, not the number of values.
func (s *RangeSet[T]) Len() int { return len(s.rs) }

// IsEmpty reports whether the set holds no values.
func (s *RangeSet[T]) IsEmpty() bool { return len(s.rs) == 0 }

// Clear removes everything.
func (s *RangeSet[T]) Clear() { s.rs = nil }

// Ranges iterates over the stored ranges in ascending order. The set must not
// be modified during iteration.
func (s *RangeSet[T]) Ranges() iter.Seq[Range[T]] {
	return func(yield func(Range[T]) bool) {
		for _, r := range s.rs {
			if !yield(r) {
				return
			}
		}
	}
}

// Complement returns a new set holding every value of the domain not in s.
func (s *RangeSet[T]) Complement() *RangeSet[T] {
	out := &RangeSet[T]{}
	prev := cut[T]{kind: belowAll}
	for _, r := range s.rs {
		if compareCuts(prev, r.lo) < 0 {
			out.rs = append(out.rs, Range[T]{prev, r.lo})
		}
		prev = r.hi
	}
	if end := (cut[T]{kind: aboveAll}); compareCuts(prev, end) < 0 {
		out.rs = append(out.rs, Range[T]{prev, end})
	}
	return out
}

// Union returns a new set holding the values of s or o.
func (s *RangeSet[T]) Union(o *RangeSet[T]) *RangeSet[T] {
	out := s.Clone()
	out.AddAll(o)
	return out
}

// Difference returns a new set holding the values of s that are not in o.
func (s *RangeSet[T]) Difference(o *RangeSet[T]) *RangeSet[T] {
	out := s.Clone()
	out.RemoveAll(o)
	return out
}

// Intersection returns a new set holding the values in both s and o. O(n + m).
func (s *RangeSet[T]) Intersection(o *RangeSet[T]) *RangeSet[T] {
	out := &RangeSet[T]{}
	i, j := 0, 0
	for i < len(s.rs) && j < len(o.rs) {
		a, b := s.rs[i], o.rs[j]
		lo, hi := maxCut(a.lo, b.lo), minCut(a.hi, b.hi)
		if compareCuts(lo, hi) < 0 {
			out.rs = append(out.rs, Range[T]{lo, hi})
		}
		if compareCuts(a.hi, b.hi) < 0 {
			i++
		} else {
			j++
		}
	}
	return out
}

// Clone returns an independent copy.
func (s *RangeSet[T]) Clone() *RangeSet[T] { return &RangeSet[T]{rs: slices.Clone(s.rs)} }

// Equal reports whether both sets hold the same values.
func (s *RangeSet[T]) Equal(o *RangeSet[T]) bool { return slices.Equal(s.rs, o.rs) }

// String renders the ranges, e.g. "{[1..3) [5..8]}".
func (s *RangeSet[T]) String() string {
	out := "{"
	for i, r := range s.rs {
		if i > 0 {
			out += " "
		}
		out += r.String()
	}
	return out + "}"
}

// MarshalJSON encodes the set as an ascending array of Range objects.
func (s *RangeSet[T]) MarshalJSON() ([]byte, error) {
	if s.rs == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(s.rs)
}

// UnmarshalJSON decodes an array of Range objects (in any order, coalescing as
// Add does), replacing the contents.
func (s *RangeSet[T]) UnmarshalJSON(b []byte) error {
	var in []Range[T]
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	*s = *NewRangeSet(in...)
	return nil
}

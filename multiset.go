package collectx

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
)

// Multiset is an unordered collection that may contain an element more than
// once. It is the Go counterpart of Guava's HashMultiset.
//
// Add, Remove, Count and SetCount are O(1) on average. Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable multiset.
type Multiset[E comparable] struct {
	counts map[E]int
	size   int
}

// NewMultiset returns a multiset holding one occurrence of each given element
// (repeats accumulate).
func NewMultiset[E comparable](elems ...E) *Multiset[E] {
	m := &Multiset[E]{}
	for _, e := range elems {
		m.Add(e, 1)
	}
	return m
}

// Add adds n occurrences of e and returns the count before the call.
// A non-positive n is a no-op.
func (m *Multiset[E]) Add(e E, n int) int {
	prev := m.counts[e]
	if n <= 0 {
		return prev
	}
	if m.counts == nil {
		m.counts = make(map[E]int)
	}
	m.counts[e] = prev + n
	m.size += n
	return prev
}

// Remove removes up to n occurrences of e and returns the count before the
// call. A non-positive n is a no-op.
func (m *Multiset[E]) Remove(e E, n int) int {
	prev := m.counts[e]
	if n <= 0 || prev == 0 {
		return prev
	}
	if n >= prev {
		delete(m.counts, e)
		m.size -= prev
		return prev
	}
	m.counts[e] = prev - n
	m.size -= n
	return prev
}

// SetCount sets the count of e to n (n < 0 is treated as 0) and returns the
// previous count.
func (m *Multiset[E]) SetCount(e E, n int) int {
	prev := m.counts[e]
	n = max(n, 0)
	if n == prev {
		return prev
	}
	if n == 0 {
		delete(m.counts, e)
	} else {
		if m.counts == nil {
			m.counts = make(map[E]int)
		}
		m.counts[e] = n
	}
	m.size += n - prev
	return prev
}

// Count returns the number of occurrences of e.
func (m *Multiset[E]) Count(e E) int { return m.counts[e] }

// Contains reports whether e occurs at least once.
func (m *Multiset[E]) Contains(e E) bool { return m.counts[e] > 0 }

// Len returns the total number of occurrences, counting duplicates.
func (m *Multiset[E]) Len() int { return m.size }

// DistinctLen returns the number of distinct elements.
func (m *Multiset[E]) DistinctLen() int { return len(m.counts) }

// Clear removes everything.
func (m *Multiset[E]) Clear() {
	clear(m.counts)
	m.size = 0
}

// All iterates over (element, count) pairs in unspecified order. The multiset
// must not be modified during iteration.
func (m *Multiset[E]) All() iter.Seq2[E, int] {
	return func(yield func(E, int) bool) {
		for e, n := range m.counts {
			if !yield(e, n) {
				return
			}
		}
	}
}

// Elements iterates over every occurrence, so an element with count 3 is
// yielded 3 times in a row. Order between distinct elements is unspecified.
func (m *Multiset[E]) Elements() iter.Seq[E] {
	return func(yield func(E) bool) {
		for e, n := range m.counts {
			for range n {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// Distinct iterates over the distinct elements in unspecified order.
func (m *Multiset[E]) Distinct() iter.Seq[E] {
	return func(yield func(E) bool) {
		for e := range m.counts {
			if !yield(e) {
				return
			}
		}
	}
}

// Equal reports whether both multisets hold the same counts.
func (m *Multiset[E]) Equal(o *Multiset[E]) bool {
	if m.size != o.size || len(m.counts) != len(o.counts) {
		return false
	}
	for e, n := range m.counts {
		if o.counts[e] != n {
			return false
		}
	}
	return true
}

// Clone returns an independent copy.
func (m *Multiset[E]) Clone() *Multiset[E] {
	c := &Multiset[E]{size: m.size}
	if len(m.counts) > 0 {
		c.counts = make(map[E]int, len(m.counts))
		for e, n := range m.counts {
			c.counts[e] = n
		}
	}
	return c
}

// ContainsAll reports whether o is a sub-multiset of m: for every element,
// m.Count(e) >= o.Count(e).
func (m *Multiset[E]) ContainsAll(o *Multiset[E]) bool {
	for e, n := range o.counts {
		if m.counts[e] < n {
			return false
		}
	}
	return true
}

// Union returns a new multiset holding the maximum count of each element.
func (m *Multiset[E]) Union(o *Multiset[E]) *Multiset[E] {
	r := m.Clone()
	for e, n := range o.counts {
		if n > r.counts[e] {
			r.SetCount(e, n)
		}
	}
	return r
}

// Intersection returns a new multiset holding the minimum count of each element.
func (m *Multiset[E]) Intersection(o *Multiset[E]) *Multiset[E] {
	r := &Multiset[E]{}
	for e, n := range m.counts {
		if k := min(n, o.counts[e]); k > 0 {
			r.Add(e, k)
		}
	}
	return r
}

// Sum returns a new multiset holding the sum of counts of each element.
func (m *Multiset[E]) Sum(o *Multiset[E]) *Multiset[E] {
	r := m.Clone()
	for e, n := range o.counts {
		r.Add(e, n)
	}
	return r
}

// Difference returns a new multiset with max(0, m.Count(e) - o.Count(e)).
func (m *Multiset[E]) Difference(o *Multiset[E]) *Multiset[E] {
	r := m.Clone()
	for e, n := range o.counts {
		r.Remove(e, n)
	}
	return r
}

type msEntry[E comparable] struct {
	Element E   `json:"element"`
	Count   int `json:"count"`
}

// MarshalJSON encodes the multiset as an array of {"element":e,"count":n}
// objects in unspecified order. It works for any JSON-encodable element type.
func (m *Multiset[E]) MarshalJSON() ([]byte, error) {
	out := make([]msEntry[E], 0, len(m.counts))
	for e, n := range m.counts {
		out = append(out, msEntry[E]{e, n})
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes the format written by MarshalJSON, replacing the
// contents. Counts must be positive.
func (m *Multiset[E]) UnmarshalJSON(b []byte) error {
	var in []msEntry[E]
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	fresh := Multiset[E]{}
	for _, en := range in {
		if en.Count <= 0 {
			return errors.New("collectx: multiset count must be positive")
		}
		fresh.Add(en.Element, en.Count)
	}
	*m = fresh
	return nil
}

// String implements fmt.Stringer, e.g. "multiset[a x2 b]" in unspecified order.
func (m *Multiset[E]) String() string {
	s := "multiset["
	first := true
	for e, n := range m.counts {
		if !first {
			s += " "
		}
		first = false
		if n == 1 {
			s += fmt.Sprint(e)
		} else {
			s += fmt.Sprintf("%v x%d", e, n)
		}
	}
	return s + "]"
}

package collectx

import (
	"encoding/json"
	"iter"
	"slices"
)

// ListMultimap maps each key to an ordered list of values, duplicates allowed
// (Guava's ArrayListMultimap). Values for one key keep insertion order.
//
// Put is amortized O(1); Get is O(v) for v values of the key (it copies);
// Remove is O(v); RemoveAll and ContainsKey are O(1); Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable multimap.
type ListMultimap[K, V comparable] struct {
	m    map[K][]V
	size int
}

// NewListMultimap returns an empty ListMultimap.
func NewListMultimap[K, V comparable]() *ListMultimap[K, V] { return &ListMultimap[K, V]{} }

// Put appends v to the values of k.
func (mm *ListMultimap[K, V]) Put(k K, v V) {
	if mm.m == nil {
		mm.m = make(map[K][]V)
	}
	mm.m[k] = append(mm.m[k], v)
	mm.size++
}

// PutAll appends all vs to the values of k, in order.
func (mm *ListMultimap[K, V]) PutAll(k K, vs ...V) {
	if len(vs) == 0 {
		return
	}
	if mm.m == nil {
		mm.m = make(map[K][]V)
	}
	mm.m[k] = append(mm.m[k], vs...)
	mm.size += len(vs)
}

// Get returns a copy of the values of k in insertion order, or nil.
func (mm *ListMultimap[K, V]) Get(k K) []V { return slices.Clone(mm.m[k]) }

// Remove removes the first occurrence of the entry (k, v) and reports whether
// one existed.
func (mm *ListMultimap[K, V]) Remove(k K, v V) bool {
	vs := mm.m[k]
	i := slices.Index(vs, v)
	if i < 0 {
		return false
	}
	if len(vs) == 1 {
		delete(mm.m, k)
	} else {
		mm.m[k] = slices.Delete(vs, i, i+1)
	}
	mm.size--
	return true
}

// RemoveAll removes every value of k and returns them.
func (mm *ListMultimap[K, V]) RemoveAll(k K) []V {
	vs := mm.m[k]
	delete(mm.m, k)
	mm.size -= len(vs)
	return vs
}

// ReplaceValues replaces all values of k with vs and returns the old ones.
func (mm *ListMultimap[K, V]) ReplaceValues(k K, vs ...V) []V {
	old := mm.RemoveAll(k)
	mm.PutAll(k, vs...)
	return old
}

// ContainsKey reports whether k has at least one value.
func (mm *ListMultimap[K, V]) ContainsKey(k K) bool { return len(mm.m[k]) > 0 }

// ContainsEntry reports whether (k, v) is present. O(values of k).
func (mm *ListMultimap[K, V]) ContainsEntry(k K, v V) bool {
	return slices.Contains(mm.m[k], v)
}

// Len returns the number of entries (key/value pairs, duplicates counted).
func (mm *ListMultimap[K, V]) Len() int { return mm.size }

// KeyLen returns the number of distinct keys.
func (mm *ListMultimap[K, V]) KeyLen() int { return len(mm.m) }

// Clear removes everything.
func (mm *ListMultimap[K, V]) Clear() {
	clear(mm.m)
	mm.size = 0
}

// Keys iterates over distinct keys in unspecified order.
func (mm *ListMultimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range mm.m {
			if !yield(k) {
				return
			}
		}
	}
}

// All iterates over every (key, value) entry. Key order is unspecified; values
// of one key are yielded in insertion order. The multimap must not be
// modified during iteration.
func (mm *ListMultimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vs := range mm.m {
			for _, v := range vs {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Values iterates over every value, in unspecified key order.
func (mm *ListMultimap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, vs := range mm.m {
			for _, v := range vs {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Equal reports whether both multimaps hold the same values, in the same order,
// for every key.
func (mm *ListMultimap[K, V]) Equal(o *ListMultimap[K, V]) bool {
	if mm.size != o.size || len(mm.m) != len(o.m) {
		return false
	}
	for k, vs := range mm.m {
		if !slices.Equal(vs, o.m[k]) {
			return false
		}
	}
	return true
}

// Clone returns an independent copy.
func (mm *ListMultimap[K, V]) Clone() *ListMultimap[K, V] {
	c := &ListMultimap[K, V]{size: mm.size}
	if len(mm.m) > 0 {
		c.m = make(map[K][]V, len(mm.m))
		for k, vs := range mm.m {
			c.m[k] = slices.Clone(vs)
		}
	}
	return c
}

// Invert returns a new multimap with keys and values swapped.
func (mm *ListMultimap[K, V]) Invert() *ListMultimap[V, K] {
	r := &ListMultimap[V, K]{}
	for k, v := range mm.All() {
		r.Put(v, k)
	}
	return r
}

// MarshalJSON encodes the multimap as an object of key to array of values.
// Keys must be JSON object-key types (string, integers, or TextMarshaler).
func (mm *ListMultimap[K, V]) MarshalJSON() ([]byte, error) {
	if mm.m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(mm.m)
}

// UnmarshalJSON decodes the format written by MarshalJSON, replacing the
// contents. Keys with an empty array are dropped.
func (mm *ListMultimap[K, V]) UnmarshalJSON(b []byte) error {
	var in map[K][]V
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	fresh := ListMultimap[K, V]{}
	for k, vs := range in {
		fresh.PutAll(k, vs...)
	}
	*mm = fresh
	return nil
}

// SetMultimap maps each key to a set of values: a (key, value) pair is stored
// at most once (Guava's HashMultimap). Value order per key is unspecified.
//
// Put, Remove, ContainsEntry, ContainsKey are O(1) average; RemoveAll is O(1)
// too; Get is O(v). Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable multimap.
type SetMultimap[K, V comparable] struct {
	m    map[K]map[V]struct{}
	size int
}

// NewSetMultimap returns an empty SetMultimap.
func NewSetMultimap[K, V comparable]() *SetMultimap[K, V] { return &SetMultimap[K, V]{} }

// Put adds (k, v) and reports whether the multimap changed.
func (mm *SetMultimap[K, V]) Put(k K, v V) bool {
	s := mm.m[k]
	if s == nil {
		if mm.m == nil {
			mm.m = make(map[K]map[V]struct{})
		}
		s = make(map[V]struct{})
		mm.m[k] = s
	}
	if _, ok := s[v]; ok {
		return false
	}
	s[v] = struct{}{}
	mm.size++
	return true
}

// PutAll adds (k, v) for each v and returns how many were new.
func (mm *SetMultimap[K, V]) PutAll(k K, vs ...V) int {
	n := 0
	for _, v := range vs {
		if mm.Put(k, v) {
			n++
		}
	}
	return n
}

// Get iterates over the values of k in unspecified order. The multimap must
// not be modified during iteration.
func (mm *SetMultimap[K, V]) Get(k K) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range mm.m[k] {
			if !yield(v) {
				return
			}
		}
	}
}

// GetLen returns the number of values of k.
func (mm *SetMultimap[K, V]) GetLen(k K) int { return len(mm.m[k]) }

// Remove removes (k, v) and reports whether it was present.
func (mm *SetMultimap[K, V]) Remove(k K, v V) bool {
	s := mm.m[k]
	if _, ok := s[v]; !ok {
		return false
	}
	delete(s, v)
	if len(s) == 0 {
		delete(mm.m, k)
	}
	mm.size--
	return true
}

// RemoveAll removes every value of k and returns how many were removed.
func (mm *SetMultimap[K, V]) RemoveAll(k K) int {
	n := len(mm.m[k])
	delete(mm.m, k)
	mm.size -= n
	return n
}

// ContainsKey reports whether k has at least one value.
func (mm *SetMultimap[K, V]) ContainsKey(k K) bool { return len(mm.m[k]) > 0 }

// ContainsEntry reports whether (k, v) is present.
func (mm *SetMultimap[K, V]) ContainsEntry(k K, v V) bool {
	_, ok := mm.m[k][v]
	return ok
}

// Len returns the number of (key, value) pairs.
func (mm *SetMultimap[K, V]) Len() int { return mm.size }

// KeyLen returns the number of distinct keys.
func (mm *SetMultimap[K, V]) KeyLen() int { return len(mm.m) }

// Clear removes everything.
func (mm *SetMultimap[K, V]) Clear() {
	clear(mm.m)
	mm.size = 0
}

// Keys iterates over distinct keys in unspecified order.
func (mm *SetMultimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range mm.m {
			if !yield(k) {
				return
			}
		}
	}
}

// All iterates over every (key, value) pair in unspecified order. The
// multimap must not be modified during iteration.
func (mm *SetMultimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, s := range mm.m {
			for v := range s {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Equal reports whether both multimaps hold the same pairs.
func (mm *SetMultimap[K, V]) Equal(o *SetMultimap[K, V]) bool {
	if mm.size != o.size || len(mm.m) != len(o.m) {
		return false
	}
	for k, s := range mm.m {
		os := o.m[k]
		if len(os) != len(s) {
			return false
		}
		for v := range s {
			if _, ok := os[v]; !ok {
				return false
			}
		}
	}
	return true
}

// Clone returns an independent copy.
func (mm *SetMultimap[K, V]) Clone() *SetMultimap[K, V] {
	c := &SetMultimap[K, V]{}
	for k, v := range mm.All() {
		c.Put(k, v)
	}
	return c
}

// Invert returns a new multimap with keys and values swapped.
func (mm *SetMultimap[K, V]) Invert() *SetMultimap[V, K] {
	r := &SetMultimap[V, K]{}
	for k, v := range mm.All() {
		r.Put(v, k)
	}
	return r
}

// MarshalJSON encodes the multimap as an object of key to array of values.
// Value order inside each array is unspecified.
func (mm *SetMultimap[K, V]) MarshalJSON() ([]byte, error) {
	out := make(map[K][]V, len(mm.m))
	for k, s := range mm.m {
		vs := make([]V, 0, len(s))
		for v := range s {
			vs = append(vs, v)
		}
		out[k] = vs
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes the format written by MarshalJSON, replacing the
// contents. Duplicate values are collapsed.
func (mm *SetMultimap[K, V]) UnmarshalJSON(b []byte) error {
	var in map[K][]V
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	fresh := SetMultimap[K, V]{}
	for k, vs := range in {
		fresh.PutAll(k, vs...)
	}
	*mm = fresh
	return nil
}

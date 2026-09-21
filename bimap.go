package collectx

import (
	"encoding/json"
	"fmt"
	"iter"
)

// BiMap is a map that also maintains the inverse value-to-key mapping, so keys
// and values are both unique (Guava's HashBiMap).
//
// Put, ForcePut, Get, GetKey, Remove and RemoveValue are O(1) average; memory
// is two maps. Inverse returns a live view in O(1).
// Not safe for concurrent use. The zero value is an empty, usable BiMap.
type BiMap[K, V comparable] struct {
	fwd map[K]V
	inv map[V]K
}

// NewBiMap returns an empty BiMap.
func NewBiMap[K, V comparable]() *BiMap[K, V] { return &BiMap[K, V]{} }

func (b *BiMap[K, V]) init() {
	if b.fwd == nil {
		b.fwd = make(map[K]V)
		b.inv = make(map[V]K)
	}
}

// Put binds k to v. If k is already bound to another value that binding is
// replaced. If v is already bound to a different key, Put changes nothing and
// returns ErrValueAlreadyPresent (like Guava's put, which throws
// IllegalArgumentException); use ForcePut to evict that other key instead.
func (b *BiMap[K, V]) Put(k K, v V) error {
	if other, ok := b.inv[v]; ok && other != k {
		return fmt.Errorf("%w: %v is bound to key %v", ErrValueAlreadyPresent, v, other)
	}
	b.init()
	if old, ok := b.fwd[k]; ok {
		delete(b.inv, old)
	}
	b.fwd[k] = v
	b.inv[v] = k
	return nil
}

// ForcePut binds k to v, evicting any existing binding of v to another key.
// It returns the key that was evicted, if any.
func (b *BiMap[K, V]) ForcePut(k K, v V) (evicted K, ok bool) {
	if other, has := b.inv[v]; has && other != k {
		delete(b.fwd, other)
		evicted, ok = other, true
	}
	b.init()
	if old, has := b.fwd[k]; has {
		delete(b.inv, old)
	}
	b.fwd[k] = v
	b.inv[v] = k
	return evicted, ok
}

// Get returns the value bound to k.
func (b *BiMap[K, V]) Get(k K) (V, bool) {
	v, ok := b.fwd[k]
	return v, ok
}

// GetKey returns the key bound to v.
func (b *BiMap[K, V]) GetKey(v V) (K, bool) {
	k, ok := b.inv[v]
	return k, ok
}

// ContainsKey reports whether k is bound.
func (b *BiMap[K, V]) ContainsKey(k K) bool { _, ok := b.fwd[k]; return ok }

// ContainsValue reports whether v is bound.
func (b *BiMap[K, V]) ContainsValue(v V) bool { _, ok := b.inv[v]; return ok }

// Remove deletes the binding of k and returns its value.
func (b *BiMap[K, V]) Remove(k K) (V, bool) {
	v, ok := b.fwd[k]
	if ok {
		delete(b.fwd, k)
		delete(b.inv, v)
	}
	return v, ok
}

// RemoveValue deletes the binding of v and returns its key.
func (b *BiMap[K, V]) RemoveValue(v V) (K, bool) {
	k, ok := b.inv[v]
	if ok {
		delete(b.inv, v)
		delete(b.fwd, k)
	}
	return k, ok
}

// Len returns the number of bindings.
func (b *BiMap[K, V]) Len() int { return len(b.fwd) }

// Clear removes every binding.
func (b *BiMap[K, V]) Clear() {
	clear(b.fwd)
	clear(b.inv)
}

// Inverse returns a live view with keys and values swapped. It shares storage
// with b: writes through either are visible in both. The view of a zero-value
// BiMap is initialized on demand and stays linked.
func (b *BiMap[K, V]) Inverse() *BiMap[V, K] {
	b.init()
	return &BiMap[V, K]{fwd: b.inv, inv: b.fwd}
}

// All iterates over (key, value) pairs in unspecified order. The BiMap must
// not be modified during iteration.
func (b *BiMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range b.fwd {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys iterates over keys in unspecified order.
func (b *BiMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range b.fwd {
			if !yield(k) {
				return
			}
		}
	}
}

// Values iterates over values in unspecified order.
func (b *BiMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range b.inv {
			if !yield(v) {
				return
			}
		}
	}
}

// Clone returns an independent copy.
func (b *BiMap[K, V]) Clone() *BiMap[K, V] {
	c := &BiMap[K, V]{}
	for k, v := range b.fwd {
		c.init()
		c.fwd[k] = v
		c.inv[v] = k
	}
	return c
}

// Equal reports whether both hold the same bindings.
func (b *BiMap[K, V]) Equal(o *BiMap[K, V]) bool {
	if len(b.fwd) != len(o.fwd) {
		return false
	}
	for k, v := range b.fwd {
		if ov, ok := o.fwd[k]; !ok || ov != v {
			return false
		}
	}
	return true
}

// MarshalJSON encodes the BiMap as a JSON object of key to value. Keys must be
// JSON object-key types.
func (b *BiMap[K, V]) MarshalJSON() ([]byte, error) {
	if b.fwd == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(b.fwd)
}

// UnmarshalJSON decodes an object, replacing the contents. It fails with
// ErrValueAlreadyPresent if two keys share a value and leaves b unchanged.
func (b *BiMap[K, V]) UnmarshalJSON(data []byte) error {
	var in map[K]V
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	fresh := BiMap[K, V]{}
	for k, v := range in {
		if err := fresh.Put(k, v); err != nil {
			return err
		}
	}
	*b = fresh
	return nil
}

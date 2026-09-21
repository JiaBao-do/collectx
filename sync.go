package collectx

import "sync"

// Synchronized guards any collection of this package (or any other value)
// with a sync.RWMutex. It is the opt-in answer to "the types are not
// goroutine-safe": all access goes through Update (exclusive) or View (shared).
//
// Wrap the struct value (every zero value is usable), not a pointer:
//
//	sm := collectx.NewSynchronized(collectx.SetMultimap[string, int]{})
//	sm.Update(func(m *collectx.SetMultimap[string, int]) { m.Put("a", 1) })
//	n := collectx.Read(sm, func(m *collectx.SetMultimap[string, int]) int { return m.Len() })
//
// The callback must not retain the pointer after it returns, and must not call
// Update or View on the same Synchronized (the lock is not reentrant).
// Iterators created inside a callback must be fully consumed inside it.
type Synchronized[T any] struct {
	mu sync.RWMutex
	v  T
}

// NewSynchronized wraps v (a struct value such as collectx.Multiset[int]{}).
// Do not keep another reference to the same underlying storage.
func NewSynchronized[T any](v T) *Synchronized[T] { return &Synchronized[T]{v: v} }

// Update runs f with exclusive access.
func (s *Synchronized[T]) Update(f func(v *T)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(&s.v)
}

// View runs f with shared access; f must only call read-only methods.
func (s *Synchronized[T]) View(f func(v *T)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f(&s.v)
}

// With runs f with exclusive access and returns its result.
func With[T, R any](s *Synchronized[T], f func(v *T) R) R {
	s.mu.Lock()
	defer s.mu.Unlock()
	return f(&s.v)
}

// Read runs f with shared access and returns its result; f must only call
// read-only methods.
func Read[T, R any](s *Synchronized[T], f func(v *T) R) R {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return f(&s.v)
}

package collectx

import (
	"sync"
	"testing"
)

func TestSynchronizedConcurrent(t *testing.T) {
	s := NewSynchronized(Multiset[int]{})
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 500 {
				s.Update(func(m *Multiset[int]) { m.Add(i%10, 1) })
				_ = Read(s, func(m *Multiset[int]) int { return m.Count(g) })
				s.View(func(m *Multiset[int]) { _ = m.Len() })
			}
		}()
	}
	wg.Wait()
	if got := Read(s, func(m *Multiset[int]) int { return m.Len() }); got != 4000 {
		t.Fatalf("Len = %d, want 4000", got)
	}
	if got := With(s, func(m *Multiset[int]) int { return m.Remove(0, 1) }); got != 400 {
		t.Fatalf("Remove prev = %d, want 400", got)
	}
}

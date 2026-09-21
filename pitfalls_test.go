package collectx

import (
	"encoding/json"
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
)

// Each TestPitfall* backs one claim in docs/PITFALLS.md. Change behavior,
// change both.

// Pitfall 1: not goroutine-safe; wrap in Synchronized.
func TestPitfallSynchronizedIsTheRightWay(t *testing.T) {
	s := NewSynchronized(SetMultimap[int, int]{})
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range 100 {
				s.Update(func(m *SetMultimap[int, int]) { m.Put(i, j) })
			}
		}()
	}
	wg.Wait()
	if n := Read(s, func(m *SetMultimap[int, int]) int { return m.Len() }); n != 400 {
		t.Fatalf("Len = %d", n)
	}
}

// Pitfall 2: hash-based iteration order is unspecified; sort before comparing.
func TestPitfallSortForDeterministicOutput(t *testing.T) {
	m := NewMultiset("b", "a", "c")
	got := slices.Sorted(m.Distinct())
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatal(got)
	}
}

// Pitfall 3: copies versus live views.
func TestPitfallCopiesAndViews(t *testing.T) {
	// ListMultimap.Get returns a copy: writing to it does not change the map.
	lm := NewListMultimap[string, int]()
	lm.PutAll("k", 1, 2)
	lm.Get("k")[0] = 99
	if !slices.Equal(lm.Get("k"), []int{1, 2}) {
		t.Fatal("Get must return a copy")
	}
	// BiMap.Inverse is a live view: writes through it change the original.
	b := NewBiMap[string, int]()
	_ = b.Inverse().Put(1, "one")
	if v, _ := b.Get("one"); v != 1 {
		t.Fatal("Inverse must be a live view")
	}
	// Clone is independent.
	c := b.Clone()
	c.Clear()
	if b.Len() != 1 {
		t.Fatal("Clone must be independent")
	}
}

// Pitfall 4: the domain is continuous (Guava semantics), so integer ranges
// that only look adjacent are NOT merged.
func TestPitfallContinuousDomain(t *testing.T) {
	a, _ := Closed(1, 3)
	b, _ := Closed(4, 6)
	if NewRangeSet(a, b).Len() != 2 {
		t.Fatal("[1,3] and [4,6] must stay separate")
	}
	c, _ := ClosedOpen(1, 4)
	d, _ := ClosedOpen(4, 6)
	if NewRangeSet(c, d).Len() != 1 {
		t.Fatal("[1,4) and [4,6) touch and must merge")
	}
}

// Pitfall 5: invalid input is an error, never a panic (Guava throws).
func TestPitfallErrorsInsteadOfExceptions(t *testing.T) {
	if _, err := Closed(5, 1); !errors.Is(err, ErrInvalidRange) {
		t.Fatal("Closed(5,1)")
	}
	b := NewBiMap[string, int]()
	_ = b.Put("a", 1)
	if err := b.Put("b", 1); !errors.Is(err, ErrValueAlreadyPresent) {
		t.Fatal("BiMap.Put duplicate value")
	}
	x, _ := Closed(1, 2)
	y, _ := Closed(5, 6)
	if _, ok := x.Intersection(y); ok {
		t.Fatal("disconnected Intersection must report ok=false")
	}
}

// Pitfall 6: a nil pointer receiver panics like any nil pointer; the zero
// value (not the nil pointer) is what is usable.
func TestPitfallNilVersusZeroValue(t *testing.T) {
	var zero Multiset[int]
	zero.Add(1, 1) // fine
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil-pointer panic")
		}
	}()
	var nilPtr *Multiset[int]
	nilPtr.Count(1)
}

// Pitfall 7: NaN keys never equal themselves, so they cannot be found or removed.
func TestPitfallNaNKeys(t *testing.T) {
	m := NewMultiset[float64]()
	m.Add(math.NaN(), 1)
	m.Add(math.NaN(), 1)
	if m.DistinctLen() != 2 || m.Count(math.NaN()) != 0 {
		t.Fatalf("distinct=%d count=%d", m.DistinctLen(), m.Count(math.NaN()))
	}
	m.Remove(math.NaN(), 1)
	if m.Len() != 2 {
		t.Fatal("NaN entries cannot be removed; only Clear frees them")
	}
}

// Pitfall 8: interface-typed keys with an uncomparable dynamic type panic.
func TestPitfallUncomparableInterfaceKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected runtime panic for []int key")
		}
	}()
	m := NewMultiset[any]()
	m.Add([]int{1}, 1)
}

// Pitfall 9: pointer keys compare by identity, not by pointee.
func TestPitfallPointerKeysAreIdentity(t *testing.T) {
	a, b := new(int), new(int)
	m := NewMultiset(a)
	if m.Contains(b) || !m.Contains(a) {
		t.Fatal("pointer keys must compare by address")
	}
}

// Pitfall 10: RangeSet.Len counts stored ranges, not values.
func TestPitfallRangeSetLenCountsRanges(t *testing.T) {
	r, _ := Closed(0, 1_000_000)
	if NewRangeSet(r).Len() != 1 {
		t.Fatal("Len is the number of ranges")
	}
}

// Pitfall 11: JSON decoding replaces the contents and is atomic on error.
func TestPitfallUnmarshalReplaces(t *testing.T) {
	m := NewMultiset("old")
	if err := json.Unmarshal([]byte(`[{"element":"new","count":2}]`), m); err != nil {
		t.Fatal(err)
	}
	if m.Contains("old") || m.Count("new") != 2 {
		t.Fatal("Unmarshal must replace")
	}
	if err := json.Unmarshal([]byte(`[{"element":"bad","count":0}]`), m); err == nil || m.Count("new") != 2 {
		t.Fatal("failed Unmarshal must leave the value unchanged")
	}
}

// Pitfall 12: Multiset JSON order is unspecified; compare with Equal, not bytes.
func TestPitfallJSONOrderUnspecified(t *testing.T) {
	a := NewMultiset("x", "y", "z")
	data, _ := json.Marshal(a)
	var b Multiset[string]
	if err := json.Unmarshal(data, &b); err != nil || !a.Equal(&b) {
		t.Fatal("use Equal after a round trip")
	}
}

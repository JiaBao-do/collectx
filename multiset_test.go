package collectx

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestMultisetBasics(t *testing.T) {
	var m Multiset[string] // zero value
	if m.Len() != 0 || m.Count("a") != 0 || m.Contains("a") {
		t.Fatal("zero value not empty")
	}
	if got := m.Add("a", 2); got != 0 {
		t.Fatalf("Add prev = %d", got)
	}
	if got := m.Add("a", 0); got != 2 {
		t.Fatalf("Add(0) prev = %d", got)
	}
	m.Add("a", -3) // no-op
	m.Add("b", 1)
	if m.Len() != 3 || m.DistinctLen() != 2 || m.Count("a") != 2 {
		t.Fatalf("len=%d distinct=%d", m.Len(), m.DistinctLen())
	}
	if got := m.Remove("a", 1); got != 2 || m.Count("a") != 1 {
		t.Fatal("Remove partial")
	}
	if got := m.Remove("a", 10); got != 1 || m.Contains("a") || m.Len() != 1 {
		t.Fatal("Remove all")
	}
	if got := m.Remove("zz", 1); got != 0 {
		t.Fatal("Remove missing")
	}
	m.Remove("b", 0)
	if m.SetCount("b", 5) != 1 || m.Len() != 5 {
		t.Fatal("SetCount up")
	}
	m.SetCount("b", 5)
	m.SetCount("b", -1)
	if m.Len() != 0 || m.DistinctLen() != 0 {
		t.Fatal("SetCount to zero")
	}
	m.SetCount("q", 2)
	m.Clear()
	if m.Len() != 0 {
		t.Fatal("Clear")
	}
}

func TestMultisetIterationAndString(t *testing.T) {
	m := NewMultiset("a", "a", "b")
	total, distinct := 0, 0
	for range m.Elements() {
		total++
	}
	for range m.Distinct() {
		distinct++
	}
	pairs := 0
	for e, n := range m.All() {
		pairs += n
		if e == "a" && n != 2 {
			t.Fatal("count of a")
		}
	}
	if total != 3 || distinct != 2 || pairs != 3 {
		t.Fatalf("total=%d distinct=%d pairs=%d", total, distinct, pairs)
	}
	// early break paths
	for range m.Elements() {
		break
	}
	for range m.Distinct() {
		break
	}
	for range m.All() {
		break
	}
	s := m.String()
	if !strings.HasPrefix(s, "multiset[") || !strings.Contains(s, "a x2") {
		t.Fatalf("String = %q", s)
	}
}

func TestMultisetAlgebra(t *testing.T) {
	a := NewMultiset(1, 1, 2, 3)
	b := NewMultiset(1, 2, 2, 4)
	if u := a.Union(b); u.Count(1) != 2 || u.Count(2) != 2 || u.Count(3) != 1 || u.Count(4) != 1 || u.Len() != 6 {
		t.Fatalf("union %v", u)
	}
	if i := a.Intersection(b); i.Count(1) != 1 || i.Count(2) != 1 || i.Len() != 2 {
		t.Fatalf("intersection %v", i)
	}
	if s := a.Sum(b); s.Len() != 8 || s.Count(1) != 3 {
		t.Fatalf("sum %v", s)
	}
	if d := a.Difference(b); d.Count(1) != 1 || d.Count(2) != 0 || d.Count(3) != 1 || d.Len() != 2 {
		t.Fatalf("difference %v", d)
	}
	if !a.ContainsAll(NewMultiset(1, 1)) || a.ContainsAll(NewMultiset(1, 1, 1)) {
		t.Fatal("ContainsAll")
	}
	c := a.Clone()
	c.Add(9, 1)
	if a.Contains(9) || !a.Equal(a.Clone()) || a.Equal(c) || a.Equal(b) {
		t.Fatal("Clone/Equal")
	}
	var empty Multiset[int]
	if !empty.Clone().Equal(&empty) {
		t.Fatal("empty clone")
	}
}

func TestMultisetJSON(t *testing.T) {
	m := NewMultiset("x", "x", "y")
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var back Multiset[string]
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !back.Equal(m) {
		t.Fatalf("roundtrip %s", b)
	}
	if err := json.Unmarshal([]byte(`[{"element":"a","count":0}]`), &back); err == nil {
		t.Fatal("expected error for zero count")
	}
	if err := json.Unmarshal([]byte(`{`), &back); err == nil {
		t.Fatal("expected syntax error")
	}
	if !back.Equal(m) {
		t.Fatal("failed unmarshal must not modify receiver")
	}
	var e Multiset[int]
	if b, _ := json.Marshal(&e); string(b) != "[]" {
		t.Fatalf("empty = %s", b)
	}
}

// TestMultisetModel checks random operation sequences against a map model.
func TestMultisetModel(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for range 200 {
		var m Multiset[int]
		model := map[int]int{}
		for range 300 {
			e, n := r.IntN(8), r.IntN(5)-1
			switch r.IntN(3) {
			case 0:
				m.Add(e, n)
				if n > 0 {
					model[e] += n
				}
			case 1:
				m.Remove(e, n)
				if n > 0 {
					model[e] = max(0, model[e]-n)
				}
			default:
				m.SetCount(e, n)
				model[e] = max(n, 0)
			}
		}
		total := 0
		for e, n := range model {
			total += n
			if m.Count(e) != n {
				t.Fatalf("count(%d)=%d want %d", e, m.Count(e), n)
			}
		}
		if m.Len() != total {
			t.Fatalf("len=%d want %d", m.Len(), total)
		}
		got := 0
		for range m.Elements() {
			got++
		}
		if got != total {
			t.Fatalf("elements=%d want %d", got, total)
		}
	}
}

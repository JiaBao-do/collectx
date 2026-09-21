package collectx

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
)

// Grid model: endpoints are integers in [0,10]; the model tracks membership of
// every half-integer point in [-1, 11] (indices 0..24 for x = i/2 - 1). Since
// every endpoint is an integer, two ranges are connected in Guava's sense
// exactly when the model cannot separate them by a missing grid point.
const gridN = 25

func gridX(i int) float64 { return float64(i)/2 - 1 }

// spec describes a range independently of the implementation.
type spec struct {
	lo, hi int
	lk, hk int // 0 unbounded, 1 open, 2 closed
}

func (s spec) member(x float64) bool {
	if s.lk == 1 && !(x > float64(s.lo)) || s.lk == 2 && !(x >= float64(s.lo)) {
		return false
	}
	if s.hk == 1 && !(x < float64(s.hi)) || s.hk == 2 && !(x <= float64(s.hi)) {
		return false
	}
	return true
}

func (s spec) build() (Range[float64], error) {
	lo, hi := float64(s.lo), float64(s.hi)
	switch {
	case s.lk == 0 && s.hk == 0:
		return AllValues[float64](), nil
	case s.lk == 0 && s.hk == 1:
		return LessThan(hi), nil
	case s.lk == 0 && s.hk == 2:
		return AtMost(hi), nil
	case s.hk == 0 && s.lk == 1:
		return GreaterThan(lo), nil
	case s.hk == 0 && s.lk == 2:
		return AtLeast(lo), nil
	case s.lk == 1 && s.hk == 1:
		return Open(lo, hi)
	case s.lk == 1 && s.hk == 2:
		return OpenClosed(lo, hi)
	case s.lk == 2 && s.hk == 1:
		return ClosedOpen(lo, hi)
	}
	return Closed(lo, hi)
}

func randSpec(r *rand.Rand) spec {
	s := spec{lo: r.IntN(11), hi: r.IntN(11), lk: r.IntN(3), hk: r.IntN(3)}
	if s.lk != 0 && s.hk != 0 && s.lo > s.hi {
		s.lo, s.hi = s.hi, s.lo
	}
	if s.lk == 1 && s.hk == 1 && s.lo == s.hi { // (x,x) is invalid
		s.hi++
	}
	return s
}

type model [gridN]bool

func (m *model) apply(s spec, on bool) {
	for i := range gridN {
		if s.member(gridX(i)) {
			m[i] = on
		}
	}
}

func (m *model) runs() int {
	n := 0
	for i := range gridN {
		if m[i] && (i == 0 || !m[i-1]) {
			n++
		}
	}
	return n
}

func checkSet(t *testing.T, s *RangeSet[float64], m *model) {
	t.Helper()
	for i := range gridN {
		if s.Contains(gridX(i)) != m[i] {
			t.Fatalf("Contains(%v) = %v, model %v; set %v", gridX(i), !m[i], m[i], s)
		}
	}
	if s.Len() != m.runs() {
		t.Fatalf("Len %d, model runs %d; set %v", s.Len(), m.runs(), s)
	}
	var prev Range[float64]
	first := true
	for r := range s.Ranges() {
		if r.IsEmpty() {
			t.Fatalf("empty range stored: %v", s)
		}
		if !first && compareCuts(prev.hi, r.lo) >= 0 {
			t.Fatalf("ranges not sorted/disjoint/non-connected: %v", s)
		}
		prev, first = r, false
	}
}

func TestRangeSetModel(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 12))
	for range 400 {
		var s RangeSet[float64]
		var m model
		for range 40 {
			sp := randSpec(r)
			rg, err := sp.build()
			if err != nil {
				t.Fatalf("build %+v: %v", sp, err)
			}
			switch r.IntN(3) {
			case 0, 1:
				s.Add(rg)
				m.apply(sp, true)
			default:
				s.Remove(rg)
				m.apply(sp, false)
			}
			checkSet(t, &s, &m)

			// Encloses / Intersects against the model.
			q := randSpec(r)
			qr, _ := q.build()
			enc, inter := true, false
			for i := range gridN {
				if q.member(gridX(i)) {
					enc = enc && m[i]
					inter = inter || m[i]
				}
			}
			if s.Encloses(qr) != enc || s.Intersects(qr) != inter {
				t.Fatalf("Encloses/Intersects mismatch for %v in %v: got %v/%v want %v/%v",
					qr, &s, s.Encloses(qr), s.Intersects(qr), enc, inter)
			}
		}
		// Complement and binary operations.
		comp := s.Complement()
		var cm model
		for i := range gridN {
			cm[i] = !m[i]
		}
		checkSet(t, comp, &cm)
		if !comp.Complement().Equal(&s) {
			t.Fatalf("double complement %v vs %v", comp.Complement(), &s)
		}
		var o RangeSet[float64]
		var om model
		for range 5 {
			sp := randSpec(r)
			rg, _ := sp.build()
			o.Add(rg)
			om.apply(sp, true)
		}
		var um, im, dm model
		for i := range gridN {
			um[i], im[i], dm[i] = m[i] || om[i], m[i] && om[i], m[i] && !om[i]
		}
		checkSet(t, s.Union(&o), &um)
		checkSet(t, s.Intersection(&o), &im)
		checkSet(t, s.Difference(&o), &dm)
	}
}

func TestRangeSetBasics(t *testing.T) {
	var s RangeSet[int]
	if !s.IsEmpty() || s.Len() != 0 || s.Contains(1) || s.String() != "{}" {
		t.Fatal("zero value")
	}
	if _, ok := s.Span(); ok {
		t.Fatal("Span of empty")
	}
	if _, ok := s.RangeContaining(1); ok {
		t.Fatal("RangeContaining on empty")
	}
	a, _ := Closed(1, 3)
	b, _ := Closed(4, 6)
	s.Add(a)
	s.Add(b)
	if s.Len() != 2 { // Guava semantics: [1,3] and [4,6] not connected
		t.Fatalf("continuous domain: want 2 ranges, got %v", &s)
	}
	c, _ := ClosedOpen(3, 4)
	_ = c
	e, _ := ClosedOpen(5, 5)
	s.Add(e) // empty ignored
	s.Add(Range[int]{})
	if s.Len() != 2 {
		t.Fatal("empty range must be ignored")
	}
	if got := s.String(); got != "{[1..3] [4..6]}" {
		t.Fatalf("String = %s", got)
	}
	sp, _ := s.Span()
	if sp.String() != "[1..6]" {
		t.Fatalf("Span = %v", sp)
	}
	if r, ok := s.RangeContaining(5); !ok || r != b {
		t.Fatal("RangeContaining")
	}
	x := NewRangeSet(a)
	x.AddAll(NewRangeSet(b))
	if !x.Equal(&s) {
		t.Fatal("AddAll/Equal")
	}
	x.RemoveAll(NewRangeSet(b))
	if x.Len() != 1 {
		t.Fatal("RemoveAll")
	}
	for range s.Ranges() {
		break
	}
	x.Clear()
	if !x.IsEmpty() {
		t.Fatal("Clear")
	}
}

func TestRangeSetJSON(t *testing.T) {
	a, _ := Closed(1, 3)
	s := NewRangeSet(a, GreaterThan(10), LessThan(-5))
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back RangeSet[int]
	if err := json.Unmarshal(data, &back); err != nil || !back.Equal(s) {
		t.Fatalf("roundtrip %s: %v", data, err)
	}
	if err := json.Unmarshal([]byte(`[{"lower":{"value":5,"closed":true},"upper":{"value":1,"closed":true}}]`), &back); !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("want ErrInvalidRange, got %v", err)
	}
	if err := json.Unmarshal([]byte(`x`), &back); err == nil {
		t.Fatal("syntax")
	}
	var z RangeSet[int]
	if data, _ := json.Marshal(&z); string(data) != "[]" {
		t.Fatalf("zero = %s", data)
	}
}

func TestRangeAPI(t *testing.T) {
	if _, err := Closed(3, 1); !errors.Is(err, ErrInvalidRange) {
		t.Fatal("Closed(3,1)")
	}
	if _, err := Open(2, 2); !errors.Is(err, ErrInvalidRange) {
		t.Fatal("Open(2,2)")
	}
	if e, err := ClosedOpen(2, 2); err != nil || !e.IsEmpty() {
		t.Fatal("[2,2) is empty but valid")
	}
	if e, err := OpenClosed(2, 2); err != nil || !e.IsEmpty() || e.Contains(2) {
		t.Fatal("(2,2] is empty but valid")
	}
	if !Singleton(4).Contains(4) || Singleton(4).Contains(5) {
		t.Fatal("Singleton")
	}
	co, _ := ClosedOpen(1, 3)
	oc, _ := OpenClosed(1, 3)
	cl, _ := Closed(1, 3)
	op, _ := Open(1, 3)
	if !co.Contains(1) || co.Contains(3) || oc.Contains(1) || !oc.Contains(3) || op.Contains(1) || op.Contains(3) || !cl.Contains(3) {
		t.Fatal("bounds")
	}
	a, _ := ClosedOpen(1, 3)
	b, _ := ClosedOpen(3, 5)
	c, _ := Open(3, 5)
	if !a.IsConnected(b) || a.IsConnected(c) {
		t.Fatal("IsConnected")
	}
	if in, ok := a.Intersection(b); !ok || !in.IsEmpty() {
		t.Fatalf("touching intersection is empty: %v %v", in, ok)
	}
	if _, ok := a.Intersection(c); ok {
		t.Fatal("disconnected intersection must report !ok")
	}
	if sp := a.Span(c); sp.String() != "[1..5)" {
		t.Fatalf("Span = %v", sp)
	}
	if sp := a.Span(Range[int]{}); sp != a {
		t.Fatal("Span with zero")
	}
	if sp := (Range[int]{}).Span(a); sp != a {
		t.Fatal("zero.Span")
	}
	all := AllValues[int]()
	if !all.Encloses(a) || a.Encloses(all) || !a.Encloses(Range[int]{}) || !all.Contains(-99) {
		t.Fatal("Encloses")
	}
	if a.IsConnected(Range[int]{}) {
		t.Fatal("zero connected to nothing")
	}
	if _, _, ok := all.Lower(); ok {
		t.Fatal("unbounded Lower")
	}
	if _, _, ok := all.Upper(); ok {
		t.Fatal("unbounded Upper")
	}
	if v, closed, ok := a.Lower(); !ok || v != 1 || !closed {
		t.Fatal("Lower")
	}
	if v, closed, ok := a.Upper(); !ok || v != 3 || closed {
		t.Fatal("Upper")
	}
	for r, want := range map[string]Range[int]{
		"[1..3)": a, "(-∞..+∞)": all, "(-∞..3]": AtMost(3), "(2..+∞)": GreaterThan(2),
		"[2..+∞)": AtLeast(2), "(-∞..2)": LessThan(2), "(1..3]": oc, "[empty]": {},
	} {
		if want.String() != r {
			t.Fatalf("String = %s want %s", want.String(), r)
		}
	}
}

func TestRangeJSON(t *testing.T) {
	a, _ := ClosedOpen(1.5, 3.0)
	for _, r := range []Range[float64]{a, AllValues[float64](), AtMost(2.0), GreaterThan(1.0), {}} {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var back Range[float64]
		if err := json.Unmarshal(data, &back); err != nil || back != r {
			t.Fatalf("roundtrip %s: %v %v", data, back, err)
		}
	}
	var r Range[int]
	if err := json.Unmarshal([]byte(`[`), &r); err == nil {
		t.Fatal("syntax")
	}
}

func TestRangeMapModel(t *testing.T) {
	rnd := rand.New(rand.NewPCG(21, 22))
	for range 400 {
		var m RangeMap[float64, int]
		var vals [gridN]int // 0 = unmapped
		for range 40 {
			sp := randSpec(rnd)
			rg, _ := sp.build()
			v := rnd.IntN(3) + 1
			switch rnd.IntN(4) {
			case 0:
				m.Remove(rg)
				for i := range gridN {
					if sp.member(gridX(i)) {
						vals[i] = 0
					}
				}
			case 1:
				m.PutCoalescing(rg, v)
				for i := range gridN {
					if sp.member(gridX(i)) {
						vals[i] = v
					}
				}
			default:
				m.Put(rg, v)
				for i := range gridN {
					if sp.member(gridX(i)) {
						vals[i] = v
					}
				}
			}
			for i := range gridN {
				got, ok := m.Get(gridX(i))
				if (got != vals[i]) || ok != (vals[i] != 0) {
					t.Fatalf("Get(%v) = %v,%v want %v; map %v", gridX(i), got, ok, vals[i], dump(&m))
				}
			}
			var prev Range[float64]
			first := true
			for r := range m.All() {
				if r.IsEmpty() || (!first && compareCuts(prev.hi, r.lo) > 0) {
					t.Fatalf("bad entry order: %v", dump(&m))
				}
				prev, first = r, false
			}
		}
	}
}

func dump(m *RangeMap[float64, int]) []string {
	var out []string
	for r, v := range m.All() {
		out = append(out, r.String()+"="+string(rune('0'+v)))
	}
	return out
}

func TestRangeMapCoalescing(t *testing.T) {
	var m RangeMap[int, string]
	a, _ := ClosedOpen(0, 10)
	b, _ := ClosedOpen(10, 20)
	c, _ := ClosedOpen(20, 30)
	m.PutCoalescing(a, "x")
	m.PutCoalescing(c, "x")
	if m.Len() != 2 {
		t.Fatal("not connected yet")
	}
	m.PutCoalescing(b, "x")
	if m.Len() != 1 {
		t.Fatalf("want one merged entry, got %v", m.Len())
	}
	if r, v, ok := m.GetEntry(15); !ok || v != "x" || r.String() != "[0..30)" {
		t.Fatalf("entry %v %v %v", r, v, ok)
	}
	// Put does not coalesce.
	m2 := NewRangeMap[int, string]()
	m2.Put(a, "x")
	m2.Put(b, "x")
	if m2.Len() != 2 {
		t.Fatal("Put must not coalesce")
	}
	// Overwrite in the middle splits the neighbour.
	mid, _ := ClosedOpen(5, 8)
	m2.Put(mid, "y")
	if m2.Len() != 4 {
		t.Fatalf("split: %v", m2.Len())
	}
	if v, _ := m2.Get(4); v != "x" {
		t.Fatal("left remainder")
	}
	if v, _ := m2.Get(8); v != "x" {
		t.Fatal("right remainder")
	}
	if sp, ok := m2.Span(); !ok || sp.String() != "[0..20)" {
		t.Fatalf("Span %v", sp)
	}
	m2.Put(Range[int]{}, "z") // ignored
	m2.Remove(Range[int]{})
	m2.PutCoalescing(Range[int]{}, "z")
	if m2.Len() != 4 {
		t.Fatal("empty ranges ignored")
	}
	if !m2.Clone().Equal(m2) {
		t.Fatal("Clone/Equal")
	}
	for range m2.All() {
		break
	}
	m2.Clear()
	if _, ok := m2.Get(1); ok || m2.Len() != 0 {
		t.Fatal("Clear")
	}
	if _, ok := m2.Span(); ok {
		t.Fatal("Span empty")
	}
}

func TestRangeMapJSON(t *testing.T) {
	m := NewRangeMap[int, string]()
	a, _ := ClosedOpen(0, 10)
	m.Put(a, "lo")
	m.Put(AtLeast(10), "hi")
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var back RangeMap[int, string]
	if err := json.Unmarshal(data, &back); err != nil || !back.Equal(m) {
		t.Fatalf("roundtrip %s: %v", data, err)
	}
	if err := json.Unmarshal([]byte(`{`), &back); err == nil {
		t.Fatal("syntax")
	}
	if !slices.Equal([]int{1}, []int{1}) {
		t.Fatal("sanity")
	}
}

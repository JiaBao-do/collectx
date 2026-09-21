package collectx

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestListMultimap(t *testing.T) {
	var mm ListMultimap[string, int]
	if mm.Len() != 0 || mm.Get("a") != nil || mm.ContainsKey("a") {
		t.Fatal("zero value")
	}
	mm.Put("a", 1)
	mm.Put("a", 1) // duplicates allowed
	mm.Put("a", 2)
	mm.PutAll("b", 5, 6)
	mm.PutAll("c")
	if mm.Len() != 5 || mm.KeyLen() != 2 || mm.ContainsKey("c") {
		t.Fatalf("len=%d keys=%d", mm.Len(), mm.KeyLen())
	}
	if got := mm.Get("a"); !slices.Equal(got, []int{1, 1, 2}) {
		t.Fatalf("Get = %v", got)
	}
	// Get returns a copy.
	g := mm.Get("a")
	g[0] = 99
	if !mm.ContainsEntry("a", 1) || mm.ContainsEntry("a", 99) {
		t.Fatal("Get must copy")
	}
	if !mm.Remove("a", 1) || !slices.Equal(mm.Get("a"), []int{1, 2}) || mm.Remove("a", 42) || mm.Remove("zz", 1) {
		t.Fatal("Remove")
	}
	mm.Remove("a", 1)
	mm.Remove("a", 2)
	if mm.ContainsKey("a") || mm.KeyLen() != 1 {
		t.Fatal("key must vanish with last value")
	}
	if old := mm.ReplaceValues("b", 7); !slices.Equal(old, []int{5, 6}) || !slices.Equal(mm.Get("b"), []int{7}) {
		t.Fatal("ReplaceValues")
	}
	if rm := mm.RemoveAll("b"); len(rm) != 1 || mm.Len() != 0 {
		t.Fatal("RemoveAll")
	}
	mm.Put("k", 1)
	mm.Clear()
	if mm.Len() != 0 || NewListMultimap[int, int]().Len() != 0 {
		t.Fatal("Clear")
	}
}

func TestListMultimapIterEqualClone(t *testing.T) {
	mm := NewListMultimap[string, int]()
	mm.PutAll("a", 1, 2, 3)
	mm.PutAll("b", 4)
	n := 0
	for k, v := range mm.All() {
		n++
		if (k == "b") != (v == 4) {
			t.Fatal("pair mismatch")
		}
	}
	keys := slices.Sorted(mm.Keys())
	vals := slices.Sorted(mm.Values())
	if n != 4 || !slices.Equal(keys, []string{"a", "b"}) || !slices.Equal(vals, []int{1, 2, 3, 4}) {
		t.Fatalf("n=%d keys=%v vals=%v", n, keys, vals)
	}
	for range mm.All() {
		break
	}
	for range mm.Keys() {
		break
	}
	for range mm.Values() {
		break
	}
	c := mm.Clone()
	if !c.Equal(mm) {
		t.Fatal("Clone/Equal")
	}
	c.Put("a", 9)
	if c.Equal(mm) || mm.ContainsEntry("a", 9) {
		t.Fatal("Clone must be independent")
	}
	o := NewListMultimap[string, int]()
	o.PutAll("a", 3, 2, 1)
	o.Put("b", 4)
	if o.Equal(mm) {
		t.Fatal("order matters for lists")
	}
	inv := mm.Invert()
	if !slices.Equal(inv.Get(4), []string{"b"}) || inv.Len() != 4 {
		t.Fatal("Invert")
	}
	var empty ListMultimap[int, int]
	if !empty.Clone().Equal(&empty) {
		t.Fatal("empty clone")
	}
}

func TestListMultimapJSON(t *testing.T) {
	mm := NewListMultimap[string, int]()
	mm.PutAll("a", 1, 2)
	b, err := json.Marshal(mm)
	if err != nil || string(b) != `{"a":[1,2]}` {
		t.Fatalf("%s %v", b, err)
	}
	var back ListMultimap[string, int]
	if err := json.Unmarshal([]byte(`{"a":[1,2],"z":[]}`), &back); err != nil || !back.Equal(mm) {
		t.Fatalf("roundtrip %v", err)
	}
	if err := json.Unmarshal([]byte(`[`), &back); err == nil {
		t.Fatal("want error")
	}
	var zero ListMultimap[string, int]
	if b, _ := json.Marshal(&zero); string(b) != "{}" {
		t.Fatalf("zero = %s", b)
	}
}

func TestSetMultimap(t *testing.T) {
	var mm SetMultimap[string, int]
	if mm.Len() != 0 || mm.ContainsKey("a") || mm.ContainsEntry("a", 1) || mm.GetLen("a") != 0 {
		t.Fatal("zero value")
	}
	if !mm.Put("a", 1) || mm.Put("a", 1) || !mm.Put("a", 2) {
		t.Fatal("Put semantics")
	}
	if mm.PutAll("b", 1, 1, 2) != 2 || mm.Len() != 4 || mm.KeyLen() != 2 || mm.GetLen("b") != 2 {
		t.Fatal("PutAll")
	}
	if got := slices.Sorted(mm.Get("a")); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("Get = %v", got)
	}
	for range mm.Get("a") {
		break
	}
	if !mm.Remove("a", 1) || mm.Remove("a", 1) || mm.Remove("zz", 1) {
		t.Fatal("Remove")
	}
	mm.Remove("a", 2)
	if mm.ContainsKey("a") || mm.KeyLen() != 1 {
		t.Fatal("key must vanish")
	}
	if mm.RemoveAll("b") != 2 || mm.Len() != 0 || mm.RemoveAll("b") != 0 {
		t.Fatal("RemoveAll")
	}
	mm.Put("x", 1)
	mm.Clear()
	if mm.Len() != 0 {
		t.Fatal("Clear")
	}
}

func TestSetMultimapIterEqualCloneJSON(t *testing.T) {
	mm := NewSetMultimap[string, int]()
	mm.PutAll("a", 1, 2)
	mm.Put("b", 3)
	n := 0
	for range mm.All() {
		n++
	}
	if n != 3 || !slices.Equal(slices.Sorted(mm.Keys()), []string{"a", "b"}) {
		t.Fatal("iteration")
	}
	for range mm.All() {
		break
	}
	for range mm.Keys() {
		break
	}
	c := mm.Clone()
	if !c.Equal(mm) {
		t.Fatal("Clone")
	}
	c.Put("b", 4)
	if c.Equal(mm) {
		t.Fatal("must differ")
	}
	d := mm.Clone()
	d.Remove("b", 3)
	d.Put("b", 5)
	if d.Equal(mm) {
		t.Fatal("same size different pair")
	}
	e := mm.Clone()
	e.Remove("b", 3)
	e.Put("c", 3)
	if e.Equal(mm) {
		t.Fatal("same size different key")
	}
	if inv := mm.Invert(); !inv.ContainsEntry(3, "b") || inv.Len() != 3 {
		t.Fatal("Invert")
	}
	b, err := json.Marshal(mm)
	if err != nil {
		t.Fatal(err)
	}
	var back SetMultimap[string, int]
	if err := json.Unmarshal(b, &back); err != nil || !back.Equal(mm) {
		t.Fatalf("roundtrip %s %v", b, err)
	}
	if err := json.Unmarshal([]byte(`{"a":[1,1]}`), &back); err != nil || back.Len() != 1 {
		t.Fatal("duplicates collapse")
	}
	if err := json.Unmarshal([]byte(`x`), &back); err == nil {
		t.Fatal("want error")
	}
}

// TestMultimapModel drives both multimaps and a naive model with the same
// random operations.
func TestMultimapModel(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for range 200 {
		var lm ListMultimap[int, int]
		var sm SetMultimap[int, int]
		lmodel := map[int][]int{}
		smodel := map[[2]int]bool{}
		for range 300 {
			k, v := r.IntN(5), r.IntN(5)
			switch r.IntN(4) {
			case 0, 1:
				lm.Put(k, v)
				lmodel[k] = append(lmodel[k], v)
				sm.Put(k, v)
				smodel[[2]int{k, v}] = true
			case 2:
				got := lm.Remove(k, v)
				i := slices.Index(lmodel[k], v)
				if got != (i >= 0) {
					t.Fatal("list Remove result")
				}
				if i >= 0 {
					lmodel[k] = slices.Delete(lmodel[k], i, i+1)
				}
				if sm.Remove(k, v) != smodel[[2]int{k, v}] {
					t.Fatal("set Remove result")
				}
				delete(smodel, [2]int{k, v})
			default:
				lm.RemoveAll(k)
				delete(lmodel, k)
				sm.RemoveAll(k)
				for kv := range smodel {
					if kv[0] == k {
						delete(smodel, kv)
					}
				}
			}
		}
		total := 0
		for k, vs := range lmodel {
			total += len(vs)
			if !slices.Equal(lm.Get(k), vs) && len(vs) > 0 {
				t.Fatalf("list key %d: %v vs %v", k, lm.Get(k), vs)
			}
		}
		if lm.Len() != total {
			t.Fatalf("list len %d want %d", lm.Len(), total)
		}
		if sm.Len() != len(smodel) {
			t.Fatalf("set len %d want %d", sm.Len(), len(smodel))
		}
		for kv := range smodel {
			if !sm.ContainsEntry(kv[0], kv[1]) {
				t.Fatalf("set missing %v", kv)
			}
		}
	}
}

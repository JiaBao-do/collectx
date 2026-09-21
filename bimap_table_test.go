package collectx

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestBiMapPut(t *testing.T) {
	var b BiMap[string, int]
	if b.Len() != 0 || b.ContainsKey("a") || b.ContainsValue(1) {
		t.Fatal("zero value")
	}
	if _, ok := b.Get("a"); ok {
		t.Fatal("Get on zero")
	}
	if err := b.Put("a", 1); err != nil {
		t.Fatal(err)
	}
	if err := b.Put("b", 1); !errors.Is(err, ErrValueAlreadyPresent) {
		t.Fatalf("want ErrValueAlreadyPresent, got %v", err)
	}
	if k, _ := b.GetKey(1); k != "a" || b.Len() != 1 {
		t.Fatal("failed Put must not change state")
	}
	if err := b.Put("a", 1); err != nil { // same binding is fine
		t.Fatal(err)
	}
	if err := b.Put("a", 2); err != nil { // rebind key
		t.Fatal(err)
	}
	if b.ContainsValue(1) || !b.ContainsValue(2) || b.Len() != 1 {
		t.Fatal("rebind must drop old inverse")
	}
}

func TestBiMapForcePutRemove(t *testing.T) {
	b := NewBiMap[string, int]()
	b.ForcePut("a", 1)
	b.ForcePut("b", 2)
	if k, ok := b.ForcePut("c", 1); !ok || k != "a" {
		t.Fatalf("evicted %v %v", k, ok)
	}
	if b.ContainsKey("a") || b.Len() != 2 {
		t.Fatal("a must be evicted")
	}
	if _, ok := b.ForcePut("c", 1); ok {
		t.Fatal("no eviction expected")
	}
	if _, ok := b.ForcePut("c", 3); ok || b.ContainsValue(1) {
		t.Fatal("rebind via ForcePut")
	}
	if v, ok := b.Remove("c"); !ok || v != 3 || b.ContainsValue(3) {
		t.Fatal("Remove")
	}
	if _, ok := b.Remove("zz"); ok {
		t.Fatal("Remove missing")
	}
	if k, ok := b.RemoveValue(2); !ok || k != "b" || b.Len() != 0 {
		t.Fatal("RemoveValue")
	}
	if _, ok := b.RemoveValue(9); ok {
		t.Fatal("RemoveValue missing")
	}
	b.ForcePut("x", 1)
	b.Clear()
	if b.Len() != 0 || b.ContainsValue(1) {
		t.Fatal("Clear")
	}
}

func TestBiMapInverseIsLiveView(t *testing.T) {
	var b BiMap[string, int] // zero value must link correctly
	inv := b.Inverse()
	if err := inv.Put(7, "seven"); err != nil {
		t.Fatal(err)
	}
	if v, ok := b.Get("seven"); !ok || v != 7 {
		t.Fatal("write through view not visible")
	}
	b.ForcePut("eight", 8)
	if k, ok := inv.Get(8); !ok || k != "eight" {
		t.Fatal("write to parent not visible in view")
	}
	if inv.Inverse().Len() != b.Len() {
		t.Fatal("double inverse")
	}
}

func TestBiMapMisc(t *testing.T) {
	b := NewBiMap[string, int]()
	_ = b.Put("a", 1)
	_ = b.Put("b", 2)
	if !slices.Equal(slices.Sorted(b.Keys()), []string{"a", "b"}) || !slices.Equal(slices.Sorted(b.Values()), []int{1, 2}) {
		t.Fatal("Keys/Values")
	}
	n := 0
	for range b.All() {
		n++
	}
	for range b.All() {
		break
	}
	for range b.Keys() {
		break
	}
	for range b.Values() {
		break
	}
	c := b.Clone()
	if n != 2 || !c.Equal(b) {
		t.Fatal("Clone/Equal")
	}
	c.ForcePut("z", 1)
	if c.Equal(b) || !b.ContainsKey("a") {
		t.Fatal("clone must be independent")
	}
	d := b.Clone()
	d.Remove("a")
	_ = d.Put("q", 1)
	if d.Equal(b) {
		t.Fatal("same size different keys")
	}
	var e BiMap[int, int]
	if !e.Clone().Equal(&e) {
		t.Fatal("empty")
	}
}

func TestBiMapJSON(t *testing.T) {
	b := NewBiMap[string, int]()
	_ = b.Put("a", 1)
	_ = b.Put("b", 2)
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var back BiMap[string, int]
	if err := json.Unmarshal(data, &back); err != nil || !back.Equal(b) {
		t.Fatalf("roundtrip %s %v", data, err)
	}
	if v, _ := back.GetKey(2); v != "b" {
		t.Fatal("inverse rebuilt")
	}
	if err := json.Unmarshal([]byte(`{"a":1,"b":1}`), &back); !errors.Is(err, ErrValueAlreadyPresent) {
		t.Fatalf("dup value: %v", err)
	}
	if !back.Equal(b) {
		t.Fatal("failed decode must not modify")
	}
	if err := json.Unmarshal([]byte(`[`), &back); err == nil {
		t.Fatal("syntax error expected")
	}
	var z BiMap[string, int]
	if data, _ := json.Marshal(&z); string(data) != "{}" {
		t.Fatalf("zero = %s", data)
	}
}

// TestBiMapModel checks the bijection invariant under random ops.
func TestBiMapModel(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	for range 200 {
		var b BiMap[int, int]
		for range 300 {
			k, v := r.IntN(6), r.IntN(6)
			switch r.IntN(4) {
			case 0:
				_ = b.Put(k, v)
			case 1:
				b.ForcePut(k, v)
			case 2:
				b.Remove(k)
			default:
				b.RemoveValue(v)
			}
			if len(b.fwd) != len(b.inv) {
				t.Fatalf("size mismatch %v %v", b.fwd, b.inv)
			}
			for k, v := range b.fwd {
				if b.inv[v] != k {
					t.Fatalf("not a bijection %v %v", b.fwd, b.inv)
				}
			}
		}
	}
}

func TestTable(t *testing.T) {
	var tb Table[string, string, int]
	if tb.Len() != 0 || tb.Contains("r", "c") {
		t.Fatal("zero")
	}
	if _, rep := tb.Put("r1", "c1", 1); rep {
		t.Fatal("first put replaced")
	}
	if old, rep := tb.Put("r1", "c1", 5); !rep || old != 1 {
		t.Fatal("replace")
	}
	tb.Put("r1", "c2", 2)
	tb.Put("r2", "c1", 3)
	if tb.Len() != 3 || tb.RowLen() != 2 {
		t.Fatalf("len=%d rows=%d", tb.Len(), tb.RowLen())
	}
	if v, ok := tb.Get("r1", "c1"); !ok || v != 5 {
		t.Fatal("Get")
	}
	rowSum, colSum := 0, 0
	for _, v := range tb.Row("r1") {
		rowSum += v
	}
	for r, v := range tb.Column("c1") {
		colSum += v
		if r != "r1" && r != "r2" {
			t.Fatal("column row key")
		}
	}
	if rowSum != 7 || colSum != 8 {
		t.Fatalf("row=%d col=%d", rowSum, colSum)
	}
	cols := tb.ColumnKeys()
	slices.Sort(cols)
	if !slices.Equal(cols, []string{"c1", "c2"}) || !slices.Equal(slices.Sorted(tb.RowKeys()), []string{"r1", "r2"}) {
		t.Fatal("keys")
	}
	for range tb.Row("r1") {
		break
	}
	for range tb.Column("c1") {
		break
	}
	for range tb.RowKeys() {
		break
	}
	for range tb.Cells() {
		break
	}
	if _, ok := tb.Remove("r9", "c"); ok {
		t.Fatal("remove missing")
	}
	if v, ok := tb.Remove("r2", "c1"); !ok || v != 3 || tb.RowLen() != 1 {
		t.Fatal("Remove drops empty row")
	}
	if tb.RemoveRow("r1") != 2 || tb.Len() != 0 || tb.RemoveRow("r1") != 0 {
		t.Fatal("RemoveRow")
	}
	tb.Put("a", "b", 1)
	tb.Clear()
	if tb.Len() != 0 || NewTable[int, int, int]().Len() != 0 {
		t.Fatal("Clear")
	}
}

func TestTableTransposeCloneEqualJSON(t *testing.T) {
	tb := NewTable[string, int, string]()
	tb.Put("a", 1, "x")
	tb.Put("a", 2, "y")
	tb.Put("b", 1, "z")
	tr := Transpose(tb)
	if v, ok := tr.Get(2, "a"); !ok || v != "y" || tr.Len() != 3 {
		t.Fatal("Transpose")
	}
	eq := func(a, b string) bool { return a == b }
	c := tb.Clone()
	if !c.Equal(tb, eq) {
		t.Fatal("Clone")
	}
	c.Put("a", 1, "changed")
	if c.Equal(tb, eq) {
		t.Fatal("value differs")
	}
	c.Put("a", 1, "x")
	c.Remove("a", 2)
	c.Put("q", 9, "y")
	if c.Equal(tb, eq) {
		t.Fatal("cell differs")
	}
	data, err := json.Marshal(tb)
	if err != nil {
		t.Fatal(err)
	}
	var back Table[string, int, string]
	if err := json.Unmarshal(data, &back); err != nil || !back.Equal(tb, eq) {
		t.Fatalf("roundtrip %s %v", data, err)
	}
	if err := json.Unmarshal([]byte(`{"a":{"x":1}}`), &back); err == nil {
		t.Fatal("bad column key must error")
	}
	var z Table[int, int, int]
	if data, _ := json.Marshal(&z); string(data) != "{}" {
		t.Fatalf("zero = %s", data)
	}
}

func TestTableModel(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	for range 200 {
		var tb Table[int, int, int]
		model := map[[2]int]int{}
		for range 300 {
			k := [2]int{r.IntN(5), r.IntN(5)}
			switch r.IntN(3) {
			case 0, 1:
				v := r.IntN(100)
				old, rep := tb.Put(k[0], k[1], v)
				mo, mrep := model[k]
				if rep != mrep || old != mo {
					t.Fatal("Put result")
				}
				model[k] = v
			default:
				v, ok := tb.Remove(k[0], k[1])
				mv, mok := model[k]
				if ok != mok || v != mv {
					t.Fatal("Remove result")
				}
				delete(model, k)
			}
		}
		if tb.Len() != len(model) {
			t.Fatalf("len %d want %d", tb.Len(), len(model))
		}
		got := 0
		for cell := range tb.Cells() {
			got++
			if model[[2]int{cell.Row, cell.Column}] != cell.Value {
				t.Fatal("cell mismatch")
			}
		}
		if got != len(model) {
			t.Fatal("cells count")
		}
	}
}

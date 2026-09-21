package collectx

import "testing"

// FuzzRangeSet interprets the input as a stream of 3-byte operations against a
// RangeSet and the grid model from range_test.go.
func FuzzRangeSet(f *testing.F) {
	f.Add([]byte{0, 1, 5, 1, 3, 4, 2, 0, 9})
	f.Add([]byte{0, 0, 0, 0, 22, 20})
	f.Fuzz(func(t *testing.T, data []byte) {
		var s RangeSet[float64]
		var m model
		for len(data) >= 3 {
			op, a, b := data[0], data[1], data[2]
			data = data[3:]
			sp := spec{lo: int(a % 11), hi: int(b % 11), lk: int(a/11) % 3, hk: int(b/11) % 3}
			if sp.lk != 0 && sp.hk != 0 && sp.lo > sp.hi {
				sp.lo, sp.hi = sp.hi, sp.lo
			}
			if sp.lk == 1 && sp.hk == 1 && sp.lo == sp.hi {
				sp.hi++
			}
			rg, err := sp.build()
			if err != nil {
				t.Fatalf("build %+v: %v", sp, err)
			}
			if op%2 == 0 {
				s.Add(rg)
				m.apply(sp, true)
			} else {
				s.Remove(rg)
				m.apply(sp, false)
			}
			checkSet(t, &s, &m)
		}
	})
}

// FuzzMultiset checks the size invariant against a map model.
func FuzzMultiset(f *testing.F) {
	f.Add([]byte{0, 1, 3, 1, 1, 2, 2, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		var ms Multiset[byte]
		model := map[byte]int{}
		for len(data) >= 3 {
			op, e, n := data[0]%3, data[1]%8, int(int8(data[2]))
			data = data[3:]
			switch op {
			case 0:
				ms.Add(e, n)
				if n > 0 {
					model[e] += n
				}
			case 1:
				ms.Remove(e, n)
				if n > 0 {
					model[e] = max(0, model[e]-n)
				}
			default:
				ms.SetCount(e, n)
				model[e] = max(0, n)
			}
		}
		total := 0
		for e, n := range model {
			total += n
			if ms.Count(e) != n {
				t.Fatalf("count(%d) = %d, want %d", e, ms.Count(e), n)
			}
		}
		if ms.Len() != total {
			t.Fatalf("Len = %d, want %d", ms.Len(), total)
		}
	})
}

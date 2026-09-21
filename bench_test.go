package collectx

import (
	"fmt"
	"testing"
)

// The naive baselines are what people write by hand without this library.

func BenchmarkListMultimapPut(b *testing.B) {
	b.ReportAllocs()
	mm := NewListMultimap[int, int]()
	for i := 0; b.Loop(); i++ {
		mm.Put(i&1023, i)
	}
}

func BenchmarkNaiveMapOfSlicesPut(b *testing.B) {
	b.ReportAllocs()
	m := map[int][]int{}
	for i := 0; b.Loop(); i++ {
		m[i&1023] = append(m[i&1023], i)
	}
}

func BenchmarkSetMultimapPut(b *testing.B) {
	b.ReportAllocs()
	mm := NewSetMultimap[int, int]()
	for i := 0; b.Loop(); i++ {
		mm.Put(i&1023, i&4095)
	}
}

func BenchmarkMultisetAdd(b *testing.B) {
	b.ReportAllocs()
	var ms Multiset[int]
	for i := 0; b.Loop(); i++ {
		ms.Add(i&1023, 1)
	}
}

func BenchmarkNaiveCountMapAdd(b *testing.B) {
	b.ReportAllocs()
	m := map[int]int{}
	for i := 0; b.Loop(); i++ {
		m[i&1023]++
	}
}

func BenchmarkBiMapPut(b *testing.B) {
	b.ReportAllocs()
	var bm BiMap[int, int]
	for i := 0; b.Loop(); i++ {
		_ = bm.Put(i&1023, i&1023)
	}
}

func BenchmarkTablePutGet(b *testing.B) {
	b.ReportAllocs()
	var t Table[int, int, int]
	for i := 0; b.Loop(); i++ {
		t.Put(i&63, i&127, i)
		t.Get(i&63, i&127)
	}
}

func rangeSetOf(n int) *RangeSet[int] {
	s := &RangeSet[int]{}
	for i := range n {
		r, _ := ClosedOpen(i*10, i*10+5)
		s.Add(r)
	}
	return s
}

func BenchmarkRangeSetContains(b *testing.B) {
	for _, n := range []int{10, 1000, 100000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			s := rangeSetOf(n)
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				s.Contains((i * 7) % (n * 10))
			}
		})
	}
}

// BenchmarkNaiveSliceScanContains is the linear-scan baseline for the above.
func BenchmarkNaiveSliceScanContains(b *testing.B) {
	for _, n := range []int{10, 1000, 100000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			type iv struct{ lo, hi int }
			ivs := make([]iv, n)
			for i := range n {
				ivs[i] = iv{i * 10, i*10 + 5}
			}
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				v := (i * 7) % (n * 10)
				for _, x := range ivs {
					if v >= x.lo && v < x.hi {
						break
					}
				}
			}
		})
	}
}

func BenchmarkRangeSetAddAppend(b *testing.B) {
	b.ReportAllocs()
	s := &RangeSet[int]{}
	for i := 0; b.Loop(); i++ {
		r, _ := ClosedOpen(i*10, i*10+5)
		s.Add(r)
	}
}

func BenchmarkRangeMapGet(b *testing.B) {
	m := NewRangeMap[int, int]()
	for i := range 1000 {
		r, _ := ClosedOpen(i*10, i*10+10)
		m.Put(r, i)
	}
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		m.Get((i * 13) % 10000)
	}
}

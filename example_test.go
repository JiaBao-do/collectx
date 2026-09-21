package collectx_test

import (
	"fmt"

	"github.com/JiaBao-do/collectx"
)

func ExampleListMultimap() {
	mm := collectx.NewListMultimap[string, int]()
	mm.Put("fruit", 1)
	mm.Put("fruit", 2)
	mm.Put("fruit", 1) // duplicates are kept
	mm.Put("veg", 7)
	fmt.Println(mm.Get("fruit"), mm.Len(), mm.KeyLen())
	// Output: [1 2 1] 4 2
}

func ExampleSetMultimap() {
	mm := collectx.NewSetMultimap[string, int]()
	fmt.Println(mm.Put("a", 1), mm.Put("a", 1), mm.Put("a", 2))
	fmt.Println(mm.Len(), mm.ContainsEntry("a", 2))
	// Output:
	// true false true
	// 2 true
}

func ExampleMultiset() {
	m := collectx.NewMultiset("a", "b", "a", "a")
	fmt.Println(m.Count("a"), m.Count("b"), m.Count("z"), m.Len(), m.DistinctLen())
	m.Remove("a", 2)
	fmt.Println(m.Count("a"))
	// Output:
	// 3 1 0 4 2
	// 1
}

func ExampleBiMap() {
	b := collectx.NewBiMap[string, int]()
	_ = b.Put("one", 1)
	_ = b.Put("two", 2)
	err := b.Put("uno", 1) // value 1 is taken
	fmt.Println(err != nil)
	k, _ := b.Inverse().Get(2)
	fmt.Println(k)
	// Output:
	// true
	// two
}

func ExampleTable() {
	t := collectx.NewTable[string, string, int]()
	t.Put("alice", "math", 90)
	t.Put("alice", "art", 70)
	t.Put("bob", "math", 80)
	sum := 0
	for _, score := range t.Column("math") {
		sum += score
	}
	fmt.Println(sum, t.Len())
	// Output: 170 3
}

func ExampleRangeSet() {
	a, _ := collectx.ClosedOpen(1, 5)
	b, _ := collectx.ClosedOpen(5, 9) // connected to a: merged
	c, _ := collectx.Closed(20, 30)
	s := collectx.NewRangeSet(a, b, c)
	fmt.Println(s, s.Contains(7), s.Contains(9))
	hole, _ := collectx.Closed(3, 4)
	s.Remove(hole)
	fmt.Println(s)
	fmt.Println(s.Complement())
	// Output:
	// {[1..9) [20..30]} true false
	// {[1..3) (4..9) [20..30]}
	// {(-∞..1) [3..4] [9..20) (30..+∞)}
}

func ExampleRangeMap() {
	m := collectx.NewRangeMap[int, string]()
	low, _ := collectx.ClosedOpen(0, 60)
	m.Put(low, "fail")
	m.Put(collectx.AtLeast(60), "pass")
	mid, _ := collectx.ClosedOpen(90, 100)
	m.Put(mid, "honors")
	for _, score := range []int{10, 60, 95, 100} {
		g, _ := m.Get(score)
		fmt.Println(score, g)
	}
	// Output:
	// 10 fail
	// 60 pass
	// 95 honors
	// 100 pass
}

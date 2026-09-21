// RangeSet and RangeMap: coalescing intervals and a step function.
// Run: go run ./examples/ranges
package main

import (
	"fmt"

	"github.com/JiaBao-do/collectx"
)

func main() {
	// Guava semantics: the domain is continuous. [9,12) and [12,13) touch, so
	// they merge; [1,3] and [4,6] would not.
	booked := collectx.NewRangeSet[int]()
	a, _ := collectx.ClosedOpen(9, 12)
	b, _ := collectx.ClosedOpen(12, 13)
	c, _ := collectx.ClosedOpen(15, 17)
	booked.Add(a)
	booked.Add(b)
	booked.Add(c)
	fmt.Println("booked:", booked)

	free := booked.Complement()
	free.Remove(collectx.LessThan(9))
	free.Remove(collectx.AtLeast(18))
	workday, _ := collectx.ClosedOpen(9, 18)
	fmt.Println("free:  ", free, "within workday:", workday.Encloses(spanOf(free)))

	// RangeMap: tax brackets.
	rates := collectx.NewRangeMap[int, string]()
	low, _ := collectx.ClosedOpen(0, 10000)
	mid, _ := collectx.ClosedOpen(10000, 50000)
	rates.Put(low, "0%")
	rates.Put(mid, "10%")
	rates.Put(collectx.AtLeast(50000), "30%")
	for _, income := range []int{500, 10000, 49999, 80000} {
		r, _ := rates.Get(income)
		fmt.Println(income, r)
	}
}

func spanOf(s *collectx.RangeSet[int]) collectx.Range[int] {
	r, _ := s.Span()
	return r
}

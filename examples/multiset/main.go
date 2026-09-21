// Multiset: word frequencies and multiset algebra.
// Run: go run ./examples/multiset
package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/JiaBao-do/collectx"
)

func main() {
	words := strings.Fields("to be or not to be that is the question to")
	freq := collectx.NewMultiset(words...)
	fmt.Println("total", freq.Len(), "distinct", freq.DistinctLen(), "to:", freq.Count("to"))

	// Sorted report: iteration order is unspecified, so sort before printing.
	type wc struct {
		w string
		n int
	}
	var all []wc
	for w, n := range freq.All() {
		all = append(all, wc{w, n})
	}
	slices.SortFunc(all, func(a, b wc) int {
		if a.n != b.n {
			return b.n - a.n
		}
		return strings.Compare(a.w, b.w)
	})
	fmt.Println(all[:3])

	other := collectx.NewMultiset("to", "to", "be")
	fmt.Println("intersection", freq.Intersection(other).Len(), "contains all:", freq.ContainsAll(other))
}

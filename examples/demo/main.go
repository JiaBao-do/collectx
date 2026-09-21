// Demo: a tiny search index and meeting-room scheduler built from the collections.
//   - SetMultimap: inverted index term -> documents
//   - Multiset:    term frequency per document
//   - Table:       (room, day) -> booked ranges
//   - RangeSet:    reject overlapping bookings
//
// Run: go run ./examples/demo
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/JiaBao-do/collectx"
)

func main() {
	docs := map[string]string{
		"a.txt": "go generics make collections easy",
		"b.txt": "collections in go use iter seq",
		"c.txt": "generics and iter",
	}
	index := collectx.NewSetMultimap[string, string]()
	tf := collectx.NewTable[string, string, int]() // doc x term -> count
	for _, name := range slices.Sorted(maps.Keys(docs)) {
		terms := collectx.NewMultiset(strings.Fields(docs[name])...)
		for term, n := range terms.All() {
			index.Put(term, name)
			tf.Put(name, term, n)
		}
	}
	fmt.Println("go:", slices.Sorted(index.Get("go")))
	fmt.Println("generics:", slices.Sorted(index.Get("generics")))

	// Room booking: reject a request that intersects an existing booking.
	rooms := collectx.NewTable[string, string, *collectx.RangeSet[int]]()
	book := func(room, day string, from, to int) {
		req, err := collectx.ClosedOpen(from, to)
		if err != nil {
			fmt.Println("invalid range:", err)
			return
		}
		set, ok := rooms.Get(room, day)
		if !ok {
			set = collectx.NewRangeSet[int]()
			rooms.Put(room, day, set)
		}
		if set.Intersects(req) {
			fmt.Printf("%s %s %d-%d: conflict\n", room, day, from, to)
			return
		}
		set.Add(req)
		fmt.Printf("%s %s %d-%d: booked\n", room, day, from, to)
	}
	book("r1", "mon", 9, 11)
	book("r1", "mon", 10, 12) // overlaps
	book("r1", "mon", 11, 12) // touches: allowed, merges
	book("r1", "tue", 9, 10)
	book("r1", "mon", 5, 3) // invalid
	set, _ := rooms.Get("r1", "mon")
	fmt.Println("r1 mon:", set)
}

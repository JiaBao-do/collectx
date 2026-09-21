// Quickstart: a multimap and a bimap in under 30 lines.
// Run: go run ./examples/quickstart
package main

import (
	"fmt"
	"slices"

	"github.com/JiaBao-do/collectx"
)

func main() {
	tags := collectx.NewSetMultimap[string, string]()
	tags.Put("post-1", "go")
	tags.Put("post-1", "generics")
	tags.Put("post-1", "go") // duplicate pair: ignored
	tags.Put("post-2", "go")
	fmt.Println(tags.Len(), slices.Sorted(tags.Get("post-1")))

	codes := collectx.NewBiMap[string, int]()
	_ = codes.Put("ok", 200)
	_ = codes.Put("missing", 404)
	if err := codes.Put("gone", 404); err != nil { // value already used
		fmt.Println("error:", err)
	}
	name, _ := codes.Inverse().Get(404)
	fmt.Println(name)
}

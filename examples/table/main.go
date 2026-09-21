// Table: sparse (row, column) -> value, e.g. student grades.
// Run: go run ./examples/table
package main

import (
	"fmt"
	"maps"
	"slices"

	"github.com/JiaBao-do/collectx"
)

func main() {
	grades := collectx.NewTable[string, string, int]()
	grades.Put("alice", "math", 90)
	grades.Put("alice", "art", 70)
	grades.Put("bob", "math", 80)

	// Row: everything for one student (sorted here, iteration order is unspecified).
	row := maps.Collect(grades.Row("alice"))
	for _, subj := range slices.Sorted(maps.Keys(row)) {
		fmt.Println("alice", subj, row[subj])
	}
	// Column: one subject across students.
	total := 0
	for _, v := range grades.Column("math") {
		total += v
	}
	fmt.Println("math total", total, "cells", grades.Len())

	bySubject := collectx.Transpose(grades) // subject -> student -> grade
	v, ok := bySubject.Get("art", "alice")
	fmt.Println(v, ok)
}

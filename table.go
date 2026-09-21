package collectx

import (
	"encoding/json"
	"iter"
)

// Cell is one (row, column, value) entry of a Table.
type Cell[R, C comparable, V any] struct {
	Row    R
	Column C
	Value  V
}

// Table is a sparse two-dimensional map from (row, column) to value, like
// Guava's HashBasedTable. Rows are indexed, columns are not.
//
// Put, Get, Contains and Remove are O(1) average. Row iteration is O(cells in
// the row). Column, ColumnKeys and Transpose scan every row: O(rows) resp.
// O(cells). Len is O(1).
// Not safe for concurrent use. The zero value is an empty, usable Table.
type Table[R, C comparable, V any] struct {
	rows map[R]map[C]V
	size int
}

// NewTable returns an empty Table.
func NewTable[R, C comparable, V any]() *Table[R, C, V] { return &Table[R, C, V]{} }

// Put stores v at (r, c) and returns the previous value, if any.
func (t *Table[R, C, V]) Put(r R, c C, v V) (old V, replaced bool) {
	row := t.rows[r]
	if row == nil {
		if t.rows == nil {
			t.rows = make(map[R]map[C]V)
		}
		row = make(map[C]V)
		t.rows[r] = row
	}
	old, replaced = row[c]
	row[c] = v
	if !replaced {
		t.size++
	}
	return old, replaced
}

// Get returns the value at (r, c).
func (t *Table[R, C, V]) Get(r R, c C) (V, bool) {
	v, ok := t.rows[r][c]
	return v, ok
}

// Contains reports whether (r, c) has a value.
func (t *Table[R, C, V]) Contains(r R, c C) bool {
	_, ok := t.rows[r][c]
	return ok
}

// Remove deletes (r, c) and returns its value. Empty rows are dropped.
func (t *Table[R, C, V]) Remove(r R, c C) (V, bool) {
	row := t.rows[r]
	v, ok := row[c]
	if !ok {
		return v, false
	}
	delete(row, c)
	if len(row) == 0 {
		delete(t.rows, r)
	}
	t.size--
	return v, true
}

// RemoveRow deletes a whole row and returns how many cells it held.
func (t *Table[R, C, V]) RemoveRow(r R) int {
	n := len(t.rows[r])
	delete(t.rows, r)
	t.size -= n
	return n
}

// Len returns the number of cells.
func (t *Table[R, C, V]) Len() int { return t.size }

// RowLen returns the number of non-empty rows.
func (t *Table[R, C, V]) RowLen() int { return len(t.rows) }

// Clear removes every cell.
func (t *Table[R, C, V]) Clear() {
	clear(t.rows)
	t.size = 0
}

// Row iterates over (column, value) pairs of row r in unspecified order. The
// table must not be modified during iteration.
func (t *Table[R, C, V]) Row(r R) iter.Seq2[C, V] {
	return func(yield func(C, V) bool) {
		for c, v := range t.rows[r] {
			if !yield(c, v) {
				return
			}
		}
	}
}

// Column iterates over (row, value) pairs of column c in unspecified order.
// It scans every row: O(rows). The table must not be modified during iteration.
func (t *Table[R, C, V]) Column(c C) iter.Seq2[R, V] {
	return func(yield func(R, V) bool) {
		for r, row := range t.rows {
			if v, ok := row[c]; ok {
				if !yield(r, v) {
					return
				}
			}
		}
	}
}

// RowKeys iterates over non-empty row keys in unspecified order.
func (t *Table[R, C, V]) RowKeys() iter.Seq[R] {
	return func(yield func(R) bool) {
		for r := range t.rows {
			if !yield(r) {
				return
			}
		}
	}
}

// ColumnKeys returns the distinct column keys in unspecified order. O(cells).
func (t *Table[R, C, V]) ColumnKeys() []C {
	seen := make(map[C]struct{})
	var out []C
	for _, row := range t.rows {
		for c := range row {
			if _, ok := seen[c]; !ok {
				seen[c] = struct{}{}
				out = append(out, c)
			}
		}
	}
	return out
}

// Cells iterates over every cell in unspecified order. The table must not be
// modified during iteration.
func (t *Table[R, C, V]) Cells() iter.Seq[Cell[R, C, V]] {
	return func(yield func(Cell[R, C, V]) bool) {
		for r, row := range t.rows {
			for c, v := range row {
				if !yield(Cell[R, C, V]{r, c, v}) {
					return
				}
			}
		}
	}
}

// Clone returns an independent copy (values are copied by assignment).
func (t *Table[R, C, V]) Clone() *Table[R, C, V] {
	n := &Table[R, C, V]{}
	for cell := range t.Cells() {
		n.Put(cell.Row, cell.Column, cell.Value)
	}
	return n
}

// Transpose returns a new table with rows and columns swapped. O(cells).
func Transpose[R, C comparable, V any](t *Table[R, C, V]) *Table[C, R, V] {
	n := &Table[C, R, V]{}
	for cell := range t.Cells() {
		n.Put(cell.Column, cell.Row, cell.Value)
	}
	return n
}

// Equal reports whether both tables hold the same cells, comparing values with
// eq.
func (t *Table[R, C, V]) Equal(o *Table[R, C, V], eq func(a, b V) bool) bool {
	if t.size != o.size {
		return false
	}
	for cell := range t.Cells() {
		ov, ok := o.Get(cell.Row, cell.Column)
		if !ok || !eq(cell.Value, ov) {
			return false
		}
	}
	return true
}

// MarshalJSON encodes the table as {"row":{"column":value}}. Row and column
// keys must be JSON object-key types.
func (t *Table[R, C, V]) MarshalJSON() ([]byte, error) {
	if t.rows == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(t.rows)
}

// UnmarshalJSON decodes the format written by MarshalJSON, replacing the
// contents. Empty rows are dropped.
func (t *Table[R, C, V]) UnmarshalJSON(b []byte) error {
	var in map[R]map[C]V
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}
	fresh := Table[R, C, V]{}
	for r, row := range in {
		for c, v := range row {
			fresh.Put(r, c, v)
		}
	}
	*t = fresh
	return nil
}

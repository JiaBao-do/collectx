package collectx

import (
	"cmp"
	"encoding/json"
	"fmt"
)

// cutKind orders the four kinds of cut: below everything, just below a value,
// just above a value, above everything.
type cutKind int8

const (
	belowAll cutKind = iota
	belowValue
	aboveValue
	aboveAll
)

// cut is a point between two adjacent values of a continuous domain, exactly
// as in Guava's Cut. A range is the half-open span [lo, hi) of cuts, which
// makes every set operation a plain comparison.
type cut[T cmp.Ordered] struct {
	kind cutKind
	v    T
}

func compareCuts[T cmp.Ordered](a, b cut[T]) int {
	if a.kind == belowValue || a.kind == aboveValue {
		if b.kind == belowValue || b.kind == aboveValue {
			if c := cmp.Compare(a.v, b.v); c != 0 {
				return c
			}
		}
	}
	return cmp.Compare(a.kind, b.kind)
}

func minCut[T cmp.Ordered](a, b cut[T]) cut[T] {
	if compareCuts(a, b) <= 0 {
		return a
	}
	return b
}

func maxCut[T cmp.Ordered](a, b cut[T]) cut[T] {
	if compareCuts(a, b) >= 0 {
		return a
	}
	return b
}

// Range is an interval over a continuous ordered domain with each end open,
// closed or unbounded, like Guava's Range. Ranges are immutable values and are
// safe to copy, compare with == (except when T is a float type holding NaN),
// and use as map keys.
//
// Guava semantics: the domain is continuous, so [1, 3] and [4, 6] are not
// connected even for integers. The zero Range is the empty range.
type Range[T cmp.Ordered] struct {
	lo, hi cut[T]
}

func makeRange[T cmp.Ordered](lo, hi cut[T]) (Range[T], error) {
	if compareCuts(lo, hi) > 0 {
		return Range[T]{}, ErrInvalidRange
	}
	return Range[T]{lo, hi}, nil
}

// Closed returns [lower, upper]. It fails with ErrInvalidRange if lower > upper.
func Closed[T cmp.Ordered](lower, upper T) (Range[T], error) {
	return makeRange(cut[T]{belowValue, lower}, cut[T]{aboveValue, upper})
}

// Open returns (lower, upper). It fails with ErrInvalidRange if lower >= upper.
func Open[T cmp.Ordered](lower, upper T) (Range[T], error) {
	return makeRange(cut[T]{aboveValue, lower}, cut[T]{belowValue, upper})
}

// ClosedOpen returns [lower, upper). It fails with ErrInvalidRange if
// lower > upper. [x, x) is a valid empty range.
func ClosedOpen[T cmp.Ordered](lower, upper T) (Range[T], error) {
	return makeRange(cut[T]{belowValue, lower}, cut[T]{belowValue, upper})
}

// OpenClosed returns (lower, upper]. It fails with ErrInvalidRange if
// lower > upper. (x, x] is a valid empty range.
func OpenClosed[T cmp.Ordered](lower, upper T) (Range[T], error) {
	return makeRange(cut[T]{aboveValue, lower}, cut[T]{aboveValue, upper})
}

// Singleton returns [v, v].
func Singleton[T cmp.Ordered](v T) Range[T] {
	return Range[T]{cut[T]{belowValue, v}, cut[T]{aboveValue, v}}
}

// AtLeast returns [v, +inf).
func AtLeast[T cmp.Ordered](v T) Range[T] {
	return Range[T]{cut[T]{belowValue, v}, cut[T]{kind: aboveAll}}
}

// GreaterThan returns (v, +inf).
func GreaterThan[T cmp.Ordered](v T) Range[T] {
	return Range[T]{cut[T]{aboveValue, v}, cut[T]{kind: aboveAll}}
}

// AtMost returns (-inf, v].
func AtMost[T cmp.Ordered](v T) Range[T] {
	return Range[T]{cut[T]{kind: belowAll}, cut[T]{aboveValue, v}}
}

// LessThan returns (-inf, v).
func LessThan[T cmp.Ordered](v T) Range[T] {
	return Range[T]{cut[T]{kind: belowAll}, cut[T]{belowValue, v}}
}

// AllValues returns (-inf, +inf).
func AllValues[T cmp.Ordered]() Range[T] {
	return Range[T]{cut[T]{kind: belowAll}, cut[T]{kind: aboveAll}}
}

// isZero reports whether r is the zero Range, which carries no bounds at all.
func (r Range[T]) isZero() bool { return r.lo.kind == belowAll && r.hi.kind == belowAll }

// IsEmpty reports whether the range contains no value: [x, x) and (x, x], and
// the zero Range.
func (r Range[T]) IsEmpty() bool { return compareCuts(r.lo, r.hi) == 0 }

// Contains reports whether v lies in the range. O(1).
func (r Range[T]) Contains(v T) bool {
	return compareCuts(r.lo, cut[T]{belowValue, v}) <= 0 &&
		compareCuts(cut[T]{aboveValue, v}, r.hi) <= 0
}

// Encloses reports whether every value of o lies in r. An empty o is enclosed
// by any range.
func (r Range[T]) Encloses(o Range[T]) bool {
	if o.isZero() {
		return true
	}
	return compareCuts(r.lo, o.lo) <= 0 && compareCuts(o.hi, r.hi) <= 0
}

// IsConnected reports whether some (possibly empty) range is enclosed by both
// r and o, that is, whether their union is a single range. [1, 3) and [3, 5)
// are connected; [1, 3) and (3, 5) are not.
func (r Range[T]) IsConnected(o Range[T]) bool {
	if r.isZero() || o.isZero() {
		return false
	}
	return compareCuts(r.lo, o.hi) <= 0 && compareCuts(o.lo, r.hi) <= 0
}

// Intersection returns the largest range enclosed by both. ok is false when
// the ranges are not connected (Guava throws IllegalArgumentException).
func (r Range[T]) Intersection(o Range[T]) (Range[T], bool) {
	if !r.IsConnected(o) {
		return Range[T]{}, false
	}
	return Range[T]{maxCut(r.lo, o.lo), minCut(r.hi, o.hi)}, true
}

// Span returns the smallest range enclosing both.
func (r Range[T]) Span(o Range[T]) Range[T] {
	if r.isZero() {
		return o
	}
	if o.isZero() {
		return r
	}
	return Range[T]{minCut(r.lo, o.lo), maxCut(r.hi, o.hi)}
}

// Lower returns the lower endpoint. ok is false if the range has no lower bound.
func (r Range[T]) Lower() (v T, closed, ok bool) {
	switch r.lo.kind {
	case belowValue:
		return r.lo.v, true, true
	case aboveValue:
		return r.lo.v, false, true
	}
	return v, false, false
}

// Upper returns the upper endpoint. ok is false if the range has no upper bound.
func (r Range[T]) Upper() (v T, closed, ok bool) {
	switch r.hi.kind {
	case aboveValue:
		return r.hi.v, true, true
	case belowValue:
		return r.hi.v, false, true
	}
	return v, false, false
}

// String renders Guava style, e.g. "[1..5)", "(-∞..3]", "(-∞..+∞)".
func (r Range[T]) String() string {
	if r.isZero() {
		return "[empty]"
	}
	s := "(-∞"
	if v, closed, ok := r.Lower(); ok {
		s = fmt.Sprintf("(%v", v)
		if closed {
			s = fmt.Sprintf("[%v", v)
		}
	}
	s += ".."
	if v, closed, ok := r.Upper(); !ok {
		s += "+∞)"
	} else if closed {
		s += fmt.Sprintf("%v]", v)
	} else {
		s += fmt.Sprintf("%v)", v)
	}
	return s
}

type boundJSON[T cmp.Ordered] struct {
	Value  T    `json:"value"`
	Closed bool `json:"closed"`
}

type rangeJSON[T cmp.Ordered] struct {
	Empty bool          `json:"empty,omitempty"`
	Lower *boundJSON[T] `json:"lower,omitempty"`
	Upper *boundJSON[T] `json:"upper,omitempty"`
}

// MarshalJSON encodes {"lower":{"value":v,"closed":b},"upper":{...}}. An
// unbounded side is omitted; the zero (empty) Range encodes as {"empty":true}.
func (r Range[T]) MarshalJSON() ([]byte, error) {
	var j rangeJSON[T]
	if r.isZero() {
		j.Empty = true
	}
	if v, c, ok := r.Lower(); ok {
		j.Lower = &boundJSON[T]{v, c}
	}
	if v, c, ok := r.Upper(); ok {
		j.Upper = &boundJSON[T]{v, c}
	}
	return json.Marshal(j)
}

// UnmarshalJSON decodes the format of MarshalJSON. It returns ErrInvalidRange
// for a reversed range.
func (r *Range[T]) UnmarshalJSON(b []byte) error {
	var j rangeJSON[T]
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	if j.Empty {
		*r = Range[T]{}
		return nil
	}
	lo, hi := cut[T]{kind: belowAll}, cut[T]{kind: aboveAll}
	if j.Lower != nil {
		lo = cut[T]{belowValue, j.Lower.Value}
		if !j.Lower.Closed {
			lo.kind = aboveValue
		}
	}
	if j.Upper != nil {
		hi = cut[T]{belowValue, j.Upper.Value}
		if j.Upper.Closed {
			hi.kind = aboveValue
		}
	}
	nr, err := makeRange(lo, hi)
	if err != nil {
		return err
	}
	*r = nr
	return nil
}

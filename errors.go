package collectx

import "errors"

// ErrValueAlreadyPresent is returned by BiMap.Put when the value is already
// bound to a different key. Use BiMap.ForcePut to evict the other binding.
var ErrValueAlreadyPresent = errors.New("collectx: value already present")

// ErrInvalidRange is returned by range constructors when the lower bound is
// greater than the upper bound, or when both bounds are equal and open.
var ErrInvalidRange = errors.New("collectx: invalid range")

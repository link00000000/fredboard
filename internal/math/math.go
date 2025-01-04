package math

import (
	"cmp"
)

func Clamp[T cmp.Ordered](v, min, max T) T {
	if v > max {
		return max
	}

	if v < min {
		return min
	}

	return v
}

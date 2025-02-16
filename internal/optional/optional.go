package optional

type Optional[T any] struct {
	value   *T
	isValid bool
}

func (o Optional[T]) Set(value T) {
	o.value = &value
	o.isValid = true
}

func (o Optional[T]) Unset() {
	o.value = nil
	o.isValid = false
}

func (o Optional[T]) IsSet() bool {
	return o.isValid
}

func (o Optional[T]) Get() T {
	if !o.isValid {
		panic("attempted to read from an unset Optional")
	}

	return *o.value
}

func Empty[T any]() Optional[T] {
	return Optional[T]{isValid: false}
}

func Make[T any](value T) Optional[T] {
	return Optional[T]{value: &value, isValid: true}
}

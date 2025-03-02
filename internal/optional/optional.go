package optional

import (
	"encoding/json"
	"fmt"
)

type Optional[T any] struct {
	value   *T
	isValid bool
}

func (o *Optional[T]) Set(value T) {
	o.value = &value
	o.isValid = true
}

func (o *Optional[T]) Unset() {
	o.value = nil
	o.isValid = false
}

func (o Optional[T]) IsSet() bool {
	return o.isValid
}

func (o Optional[T]) GetMut() *T {
	if !o.isValid {
		panic("attempted to read from an unset Optional")
	}

	return o.value
}

func (o Optional[T]) Get() T {
	return *o.GetMut()
}

// implements %#v printf
func (o Optional[T]) GoString() string {
	switch {
	case o.IsSet():
		return fmt.Sprintf("Optional{value=%+v}", o.value)
	default:
		return "Optional{not set}"
	}
}

// implements [json.Marshaler]
func (o Optional[T]) MarshalJSON() (data []byte, err error) {
	switch {
	case o.IsSet():
		return json.Marshal(o.Get())
	default:
		return json.Marshal(nil)
	}
}

// implements [json.Unmarshaler]
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 4 && string(data) == "null" {
		o.Unset()
	}

	var v T
	err := json.Unmarshal(data, &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal optional value: %w", err)
	}

	o.Set(v)
	return nil
}

func None[T any]() Optional[T] {
	return Optional[T]{isValid: false}
}

func Make[T any](value T) Optional[T] {
	return Optional[T]{value: &value, isValid: true}
}

func MakePtr[T any](value *T) Optional[T] {
	return Optional[T]{value: value, isValid: value != nil}
}

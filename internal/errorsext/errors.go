package errorsext

import (
	"errors"
	"reflect"
	"sync"
)

type ErrorList struct {
	m    sync.Mutex
	errs []error
}

func NewErrorList(errs ...error) *ErrorList {
	list := &ErrorList{errs: make([]error, 0)}

	for _, err := range errs {
		switch err := err.(type) {
		case interface{ Unwrap() []error }:
			for _, e := range err.Unwrap() {
				list.Add(e)
			}
		default:
			list.Add(err)
		}
	}

	return list
}

func (list *ErrorList) Add(errs ...error) {
	for _, err := range errs {
		if err != nil {
			list.errs = append(list.errs, err)
		}
	}
}

func (list *ErrorList) AddThreadSafe(errs ...error) {
	list.m.Lock()
	defer list.m.Unlock()

	list.Add(errs...)
}

func (list *ErrorList) Any() bool {
	return len(list.errs) != 0
}

func (list *ErrorList) Slice() []error {
	return list.errs
}

func (list *ErrorList) Join() error {
	return errors.Join(list.errs...)
}

// IsNot reports whether any error in err's tree does not match target.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly
// calling its Unwrap() error or Unwrap() []error method. When err wraps multiple
// errors, Is examines err followed by a depth-first traversal of its children.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See [syscall.Errno.Is] for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call [Unwrap] on either.
func IsNot(err, target error) bool {
	if err == nil || target == nil {
		return err != target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	return isNot(err, target, isComparable)
}

func isNot(err, target error, targetComparable bool) bool {
	for {
		if targetComparable && err != target {
			return true
		}

		if x, ok := err.(interface{ Is(error) bool }); ok && !x.Is(target) {
			return true
		}

		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return true
			}
		case interface{ Unwrap() []error }:
			for _, err := range x.Unwrap() {
				if !isNot(err, target, targetComparable) {
					return false
				}
			}
			return true
		default:
			return true
		}
	}
}

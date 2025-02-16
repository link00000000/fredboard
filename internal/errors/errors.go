package errors

import "errors"

type ErrorList struct {
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

func (list *ErrorList) Add(err error) {
	if err != nil {
		list.errs = append(list.errs, err)
	}
}

func (list *ErrorList) Slice() []error {
	return list.errs
}

func (list *ErrorList) Join() error {
	return errors.Join(list.errs...)
}

package errorsext_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"accidentallycoded.com/fredboard/v3/internal/errorsext"
)

func TestsIsNot(t *testing.T) {
	cases := []struct {
		err    error
		target error
		want   bool
	}{
		{errors.New("test error"), io.EOF, true},
		{io.EOF, io.EOF, false},
		{errors.Join(errors.New("test error 1"), errors.New("test error 2")), io.EOF, true},
		{errors.Join(io.EOF, io.EOF), io.EOF, false},
		{errors.Join(errors.New("test error 1"), io.EOF), io.EOF, true},
		{fmt.Errorf("test error: %w", io.EOF), io.EOF, true},
		{fmt.Errorf("test error: %w", io.EOF), errors.New("test error 2"), true},
		{nil, nil, false},
		{io.EOF, nil, true},
		{nil, io.EOF, true},
	}

	for _, tt := range cases {
		if got := errorsext.IsNot(tt.err, tt.target); got != tt.want {
			errS := "nil"
			if tt.err != nil {
				errS = tt.err.Error()
			}

			targetS := "nil"
			if tt.target != nil {
				targetS = tt.target.Error()
			}

			t.Errorf("IsNot(%s, %s) = %t, want %t", errS, targetS, got, tt.want)
		}
	}
}

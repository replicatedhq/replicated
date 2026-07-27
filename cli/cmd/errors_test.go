package cmd

import (
	"errors"
	"testing"
)

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: 0},
		{name: "plain error", err: errors.New("boom"), want: 1},
		{name: "exit 124", err: newExitError(124, errors.New("wait duration exceeded")), want: 124},
		{name: "wrapped exit 124", err: errors.Join(errors.New("outer"), newExitError(124, errors.New("timeout"))), want: 124},
		{name: "zero code falls back to 1", err: &ExitError{Code: 0, Err: errors.New("x")}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ExitError is a command failure that should terminate the process with a
// specific exit code. It must only be converted to os.Exit at the process
// boundary (main), so deferred cleanup and tests can still observe the error.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit status %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Cause supports github.com/pkg/errors.Cause unwrapping.
func (e *ExitError) Cause() error {
	return e.Unwrap()
}

func newExitError(code int, err error) error {
	return &ExitError{Code: code, Err: err}
}

// ExitCode returns the process exit code for err. Non-nil errors without an
// ExitError wrapper map to 1; nil maps to 0.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *ExitError
	if errors.As(err, &exitErr) && exitErr.Code != 0 {
		return exitErr.Code
	}
	return 1
}

func isRBACDeniedError(err error) bool {
	message := strings.TrimSpace(strings.ToLower(err.Error()))
	return strings.Contains(message, "access to ") && strings.HasSuffix(message, " is denied")
}

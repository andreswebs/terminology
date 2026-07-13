// Package terr defines the coded, hinted errors used across the tool, each
// carrying a stable code, exit status, and user-facing hint.
package terr

import "fmt"

// Coded is an error that carries a stable code, a process exit code, and a
// user-facing hint.
type Coded interface {
	error
	Code() string
	ExitCode() int
	Hint() string
}

// E is a coded error with a message, hint, exit code, and optional cause.
type E struct {
	code, msg, hint string
	exit            int
	cause           error
}

var registry []*E

// New creates an E and registers it so it can be enumerated via All. The
// message is formatted from format and args.
func New(code string, exit int, hint, format string, args ...any) *E {
	e := &E{
		code: code,
		exit: exit,
		hint: hint,
		msg:  fmt.Sprintf(format, args...),
	}
	registry = append(registry, e)
	return e
}

// All returns a copy of every error registered via New.
func All() []*E {
	cp := make([]*E, len(registry))
	copy(cp, registry)
	return cp
}

// Newf creates an E without registering it, formatting the message from
// format and args.
func Newf(code string, exit int, hint, format string, args ...any) *E {
	return &E{
		code: code,
		exit: exit,
		hint: hint,
		msg:  fmt.Sprintf(format, args...),
	}
}

func (e *E) Error() string { return e.msg }

// Code returns the stable error code.
func (e *E) Code() string { return e.code }

// Hint returns the user-facing hint for resolving the error.
func (e *E) Hint() string { return e.hint }

// ExitCode returns the process exit code associated with the error.
func (e *E) ExitCode() int { return e.exit }
func (e *E) Unwrap() error { return e.cause }

// Wrap returns a copy of e with cause attached as its underlying error.
func (e *E) Wrap(cause error) *E {
	cp := *e
	cp.cause = cause
	return &cp
}

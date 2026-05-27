package terr

import "fmt"

type Coded interface {
	error
	Code() string
	ExitCode() int
	Hint() string
}

type E struct {
	code, msg, hint string
	exit            int
	cause           error
}

var registry []*E

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

func All() []*E {
	cp := make([]*E, len(registry))
	copy(cp, registry)
	return cp
}

func Newf(code string, exit int, hint, format string, args ...any) *E {
	return &E{
		code: code,
		exit: exit,
		hint: hint,
		msg:  fmt.Sprintf(format, args...),
	}
}

func (e *E) Error() string { return e.msg }
func (e *E) Code() string  { return e.code }
func (e *E) Hint() string  { return e.hint }
func (e *E) ExitCode() int { return e.exit }
func (e *E) Unwrap() error { return e.cause }

func (e *E) Wrap(cause error) *E {
	cp := *e
	cp.cause = cause
	return &cp
}

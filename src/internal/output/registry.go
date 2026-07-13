package output

import "maps"

var envelopes = make(map[string]any)

// RegisterEnvelope associates command with a zero-value envelope used to
// describe its output schema.
func RegisterEnvelope(command string, zero any) {
	envelopes[command] = zero
}

// EnvelopeFor returns the registered zero-value envelope for command.
func EnvelopeFor(command string) (any, bool) {
	v, ok := envelopes[command]
	return v, ok
}

// AllEnvelopes returns a copy of the command-to-envelope registry.
func AllEnvelopes() map[string]any {
	cp := make(map[string]any, len(envelopes))
	maps.Copy(cp, envelopes)
	return cp
}

var exitCodes = make(map[string][]int)

// RegisterExitCodes records the set of exit codes command may return.
func RegisterExitCodes(command string, codes []int) {
	cp := make([]int, len(codes))
	copy(cp, codes)
	exitCodes[command] = cp
}

// ExitCodesFor returns a copy of the exit codes registered for command.
func ExitCodesFor(command string) ([]int, bool) {
	v, ok := exitCodes[command]
	if !ok {
		return nil, false
	}
	cp := make([]int, len(v))
	copy(cp, v)
	return cp, true
}

// AllExitCodes returns a deep copy of the command-to-exit-codes registry.
func AllExitCodes() map[string][]int {
	cp := make(map[string][]int, len(exitCodes))
	for k, v := range exitCodes {
		inner := make([]int, len(v))
		copy(inner, v)
		cp[k] = inner
	}
	return cp
}

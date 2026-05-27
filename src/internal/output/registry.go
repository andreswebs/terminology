package output

import "maps"

var envelopes = make(map[string]any)

func RegisterEnvelope(command string, zero any) {
	envelopes[command] = zero
}

func EnvelopeFor(command string) (any, bool) {
	v, ok := envelopes[command]
	return v, ok
}

func AllEnvelopes() map[string]any {
	cp := make(map[string]any, len(envelopes))
	maps.Copy(cp, envelopes)
	return cp
}

var exitCodes = make(map[string][]int)

func RegisterExitCodes(command string, codes []int) {
	cp := make([]int, len(codes))
	copy(cp, codes)
	exitCodes[command] = cp
}

func ExitCodesFor(command string) ([]int, bool) {
	v, ok := exitCodes[command]
	if !ok {
		return nil, false
	}
	cp := make([]int, len(v))
	copy(cp, v)
	return cp, true
}

func AllExitCodes() map[string][]int {
	cp := make(map[string][]int, len(exitCodes))
	for k, v := range exitCodes {
		inner := make([]int, len(v))
		copy(inner, v)
		cp[k] = inner
	}
	return cp
}

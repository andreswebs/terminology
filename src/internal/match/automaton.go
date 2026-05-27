package match

import (
	"bytes"
	"sort"

	"github.com/cloudflare/ahocorasick"
)

type rawMatch struct {
	PatternID int
	Start     int
	End       int
}

type automaton struct {
	machine  *ahocorasick.Matcher
	patterns [][]byte
}

func buildAutomaton(patterns [][]byte) *automaton {
	cp := make([][]byte, len(patterns))
	for i, p := range patterns {
		cp[i] = make([]byte, len(p))
		copy(cp[i], p)
	}
	return &automaton{
		machine:  ahocorasick.NewMatcher(cp),
		patterns: cp,
	}
}

func (a *automaton) Search(canonical []byte) []rawMatch {
	if len(a.patterns) == 0 {
		return nil
	}

	matched := a.machine.Match(canonical)
	var results []rawMatch
	for _, pid := range matched {
		pat := a.patterns[pid]
		off := 0
		for {
			idx := bytes.Index(canonical[off:], pat)
			if idx < 0 {
				break
			}
			start := off + idx
			results = append(results, rawMatch{
				PatternID: pid,
				Start:     start,
				End:       start + len(pat),
			})
			off = start + 1
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Start != results[j].Start {
			return results[i].Start < results[j].Start
		}
		return (results[i].End - results[i].Start) > (results[j].End - results[j].Start)
	})
	return results
}

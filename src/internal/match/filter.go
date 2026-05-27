package match

func longestMatchPerStart(matches []rawMatch) []rawMatch {
	if len(matches) == 0 {
		return nil
	}
	result := make([]rawMatch, 0, len(matches))
	prevStart := -1
	for _, m := range matches {
		if m.Start == prevStart {
			continue
		}
		result = append(result, m)
		prevStart = m.Start
	}
	return result
}

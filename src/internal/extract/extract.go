package extract

// Candidate is a discovered term along with its occurrence count, the
// heuristic that produced it, and where it was found.
type Candidate struct {
	Term      string
	Frequency int
	Heuristic string
	Locations []Location
}

// Location identifies a single occurrence of a candidate within a file.
type Location struct {
	File   string
	Line   int
	Col    int
	Offset int
}

// Span is a contiguous run of source text together with its starting position.
type Span struct {
	Text   string
	Line   int
	Col    int
	Offset int
}

func (s Span) lineColAt(byteOffset int) (int, int) {
	line := s.Line
	col := s.Col
	for i := 0; i < byteOffset && i < len(s.Text); i++ {
		if s.Text[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

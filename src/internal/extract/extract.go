package extract

type Candidate struct {
	Term      string
	Frequency int
	Heuristic string
	Locations []Location
}

type Location struct {
	File   string
	Line   int
	Col    int
	Offset int
}

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

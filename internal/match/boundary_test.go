package match

import "testing"

func TestValidBoundary(t *testing.T) {
	tests := []struct {
		name  string
		orig  string
		start int
		end   int
		want  bool
	}{
		{
			name:  "space_delimited",
			orig:  "the tzimtzum concept",
			start: 4,
			end:   12,
			want:  true,
		},
		{
			name:  "embedded_in_word",
			orig:  "pretzimtzumx",
			start: 3,
			end:   11,
			want:  false,
		},
		{
			name:  "start_of_text",
			orig:  "tzimtzum concept",
			start: 0,
			end:   8,
			want:  true,
		},
		{
			name:  "end_of_text",
			orig:  "the tzimtzum",
			start: 4,
			end:   12,
			want:  true,
		},
		{
			name:  "punctuation_boundary",
			orig:  "(tzimtzum)",
			start: 1,
			end:   9,
			want:  true,
		},
		{
			name:  "hebrew_adjacent_to_punctuation",
			orig:  "el צמצום,",
			start: 3,
			end:   3 + len("צמצום"),
			want:  true,
		},
		{
			name:  "number_boundary",
			orig:  "3tzimtzum",
			start: 1,
			end:   9,
			want:  false,
		},
		{
			name:  "both_boundaries_invalid",
			orig:  "aצמצוםb",
			start: 1,
			end:   1 + len("צמצום"),
			want:  false,
		},
		{
			name:  "entire_text",
			orig:  "tzimtzum",
			start: 0,
			end:   8,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validBoundary([]byte(tt.orig), tt.start, tt.end)
			if got != tt.want {
				t.Errorf("validBoundary(%q, %d, %d) = %v, want %v",
					tt.orig, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

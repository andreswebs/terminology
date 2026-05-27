//go:build perf

package extract_test

import (
	"strings"
	"testing"
	"time"

	"github.com/andreswebs/terminology/internal/extract"
	"github.com/andreswebs/terminology/internal/markdown"
)

func generateCorpus(targetBytes int) []byte {
	sentences := []string{
		"The Kabbalah tradition discusses many concepts of mystical Judaism. ",
		"Razón Histórica examines the philosophical underpinnings of thought. ",
		"The concept of צמצום appears in Lurianic cosmology as divine contraction. ",
		"Ein Sof represents the infinite divine essence beyond comprehension. ",
		"The Sefirot are the ten emanations through which creation unfolds. ",
		"Modern scholarship addresses the Phenomenology of religious experience. ",
		"The term אור אין סוף refers to the infinite light before creation. ",
		"Scholem and Idel represent divergent approaches to Historical Kabbalah. ",
	}

	var b strings.Builder
	for b.Len() < targetBytes {
		for _, s := range sentences {
			b.WriteString(s)
			if b.Len() >= targetBytes {
				break
			}
		}
		b.WriteString("\n\n")
	}
	return []byte(b.String())
}

func collectSpans(src []byte) []extract.Span {
	var spans []extract.Span
	for s := range markdown.Spans(src) {
		spans = append(spans, extract.Span{
			Text:   s.Text,
			Line:   s.Line,
			Col:    s.Col,
			Offset: s.Offset,
		})
	}
	return spans
}

func TestPerf_Extract_1MB(t *testing.T) {
	const budget = 2 * time.Second

	src := generateCorpus(1024 * 1024)
	spans := collectSpans(src)
	opts := extract.Options{
		MinFreq:  3,
		BaseLang: "es",
	}

	start := time.Now()
	_ = extract.CapitalizedPhrases(spans, "es")
	_ = extract.ForeignScriptTokens(spans, opts)
	_ = extract.HighFrequencyTokens(spans, opts)
	elapsed := time.Since(start)

	if elapsed > budget {
		t.Fatalf("extract 1MB corpus took %s, budget %s", elapsed, budget)
	}
	t.Logf("extract 1MB corpus: %s (budget %s)", elapsed, budget)
}

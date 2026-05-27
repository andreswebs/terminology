package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"unicode"

	"github.com/andreswebs/terminology/internal/extract"
	"github.com/andreswebs/terminology/internal/markdown"
	"github.com/andreswebs/terminology/internal/output"
	"github.com/andreswebs/terminology/internal/tbx"
	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

func Extract() *urfcli.Command {
	return &urfcli.Command{
		Name:      "extract",
		Usage:     "surface candidate terms from a markdown corpus",
		ArgsUsage: "FILE...",
		Arguments: []urfcli.Argument{
			&urfcli.StringArgs{Name: "files", Min: 1, Max: -1},
		},
		Flags: []urfcli.Flag{
			&urfcli.StringFlag{Name: "exclude", Aliases: []string{"x"}, Usage: "exclude terms already in this TBX", TakesFile: true},
			scriptPickFlag(),
			langFlag(false, "language of the corpus"),
			&urfcli.StringFlag{Name: "stopwords", Usage: "path to a newline-separated stopwords file", TakesFile: true},
			&urfcli.IntFlag{Name: "min-freq", Value: 3, Usage: "minimum frequency for high-frequency heuristic"},
			readFieldsFlag(),
		},
		Action: extractAction,
	}
}

func scriptPickFlag() urfcli.Flag {
	allowed := tbx.Script()
	set := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		set[v] = true
	}
	return &urfcli.StringFlag{
		Name:  "script",
		Usage: "filter by script: latin, hebrew, cyrillic, arabic, any",
		Value: "any",
		Validator: func(val string) error {
			if !set[val] {
				return urfcli.Exit("invalid value "+val+"; accepted: latin, hebrew, cyrillic, arabic, any", 2)
			}
			return nil
		},
	}
}

func extractAction(_ context.Context, cmd *urfcli.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	rawFiles := cmd.StringArgs("files")
	type filePair struct{ display, resolved string }
	filePairs := make([]filePair, len(rawFiles))
	for i, f := range rawFiles {
		cleaned, err := sanitizePath(f, cwd)
		if err != nil {
			return err
		}
		filePairs[i] = filePair{display: f, resolved: cleaned}
	}

	excludePath := cmd.String("exclude")
	scriptFlag := cmd.String("script")
	langFlag := cmd.String("lang")
	stopwordsPath := cmd.String("stopwords")
	minFreq := cmd.Int("min-freq")
	fieldsStr := cmd.String("fields")

	opts := extract.Options{
		Script:  scriptFlag,
		MinFreq: int(minFreq),
	}

	if stopwordsPath != "" {
		sw, err := extract.LoadStopwords(stopwordsPath)
		if err != nil {
			return terr.Newf("io_error", 3, "", "loading stopwords: %s", err)
		}
		opts.Stopwords = sw
	}

	var exclusionSet map[string]bool
	if excludePath != "" {
		g, _, err := tbx.Load(excludePath)
		if err != nil {
			if coded, ok := err.(interface{ Code() string }); ok && coded.Code() == "input_too_large" {
				return err
			}
			return terr.Newf("io_error", 3, "", "loading exclusion TBX: %s", err)
		}
		exclusionSet = buildExclusionSet(g)
	}

	var allSpans []fileSpans
	for _, fp := range filePairs {
		data, err := tbx.ReadFileBounded(fp.resolved, tbx.MaxMarkdownSize)
		if err != nil {
			if coded, ok := err.(interface{ Code() string }); ok && coded.Code() == "input_too_large" {
				return err
			}
			return terr.Newf("io_error", 3, "", "reading %s: %s", fp.display, err)
		}
		lang := extract.DetectLang(data, langFlag)
		var spans []extract.Span
		for s := range markdown.Spans(data) {
			spans = append(spans, extract.Span{
				Text:   s.Text,
				Line:   s.Line,
				Col:    s.Col,
				Offset: s.Offset,
			})
		}
		allSpans = append(allSpans, fileSpans{file: fp.display, spans: spans, lang: lang})
	}

	agg := make(map[string]*output.ExtractCandidate)

	for _, fs := range allSpans {
		caps := extract.CapitalizedPhrases(fs.spans, fs.lang)
		fileOpts := opts
		fileOpts.BaseLang = fs.lang
		foreign := extract.ForeignScriptTokens(fs.spans, fileOpts)
		freq := extract.HighFrequencyTokens(fs.spans, opts)

		for _, cList := range [][]extract.Candidate{caps, foreign, freq} {
			for _, c := range cList {
				mergeCandidate(agg, c, fs.file)
			}
		}
	}

	if exclusionSet != nil {
		fold := cases.Fold()
		for key := range agg {
			normalized := fold.String(norm.NFC.String(key))
			if exclusionSet[normalized] {
				delete(agg, key)
			}
		}
	}

	if scriptFlag != "" && scriptFlag != "any" {
		filterCandidatesByScript(agg, scriptFlag)
	}

	candidates := make([]output.ExtractCandidate, 0, len(agg))
	for _, c := range agg {
		candidates = append(candidates, *c)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Frequency != candidates[j].Frequency {
			return candidates[i].Frequency > candidates[j].Frequency
		}
		return candidates[i].Term < candidates[j].Term
	})

	env := output.ExtractEnvelope{
		SchemaVersion: output.SchemaVersion,
		OK:            true,
		Candidates:    candidates,
	}

	if fieldsStr != "" {
		fields, fErr := output.ValidateFields(fieldsStr, env)
		if fErr != nil {
			return fErr
		}

		data, mErr := json.Marshal(env)
		if mErr != nil {
			return fmt.Errorf("marshaling output: %w", mErr)
		}

		projected, pErr := output.ProjectFields(data, fields)
		if pErr != nil {
			return fmt.Errorf("projecting fields: %w", pErr)
		}

		if _, wErr := cmd.Root().Writer.Write(projected); wErr != nil {
			return fmt.Errorf("writing output: %w", wErr)
		}
		if _, wErr := cmd.Root().Writer.Write([]byte("\n")); wErr != nil {
			return fmt.Errorf("writing output: %w", wErr)
		}
	} else {
		if emitErr := output.EmitJSON(cmd.Root().Writer, env); emitErr != nil {
			return fmt.Errorf("writing output: %w", emitErr)
		}
	}

	if len(candidates) == 0 {
		return extractNoCandidates()
	}

	return nil
}

type fileSpans struct {
	file  string
	spans []extract.Span
	lang  string
}

func mergeCandidate(agg map[string]*output.ExtractCandidate, c extract.Candidate, file string) {
	existing, ok := agg[c.Term]
	if !ok {
		locs := make([]output.ExtractLocation, 0, len(c.Locations))
		for _, l := range c.Locations {
			locs = append(locs, output.ExtractLocation{
				File: file,
				Line: l.Line,
				Col:  l.Col,
			})
		}
		agg[c.Term] = &output.ExtractCandidate{
			Term:      c.Term,
			Frequency: c.Frequency,
			Heuristic: c.Heuristic,
			Locations: locs,
		}
		return
	}
	existing.Frequency += c.Frequency
	for _, l := range c.Locations {
		existing.Locations = append(existing.Locations, output.ExtractLocation{
			File: file,
			Line: l.Line,
			Col:  l.Col,
		})
	}
}

func buildExclusionSet(g *tbx.Glossary) map[string]bool {
	fold := cases.Fold()
	set := make(map[string]bool)
	for _, c := range g.Concepts {
		for _, ls := range c.Languages {
			for _, t := range ls.Terms {
				key := fold.String(norm.NFC.String(t.Surface))
				set[key] = true
			}
		}
	}
	return set
}

func filterCandidatesByScript(agg map[string]*output.ExtractCandidate, script string) {
	target := scriptTable(script)
	if target == nil {
		return
	}
	for key, c := range agg {
		if !termMatchesScript(c.Term, target) {
			delete(agg, key)
		}
	}
}

func scriptTable(name string) *unicode.RangeTable {
	switch name {
	case "latin":
		return unicode.Latin
	case "hebrew":
		return unicode.Hebrew
	case "cyrillic":
		return unicode.Cyrillic
	case "arabic":
		return unicode.Arabic
	default:
		return nil
	}
}

func termMatchesScript(term string, target *unicode.RangeTable) bool {
	for _, r := range term {
		if unicode.Is(target, r) {
			return true
		}
	}
	return false
}

func extractNoCandidates() error {
	return &extractNoCandidatesError{}
}

type extractNoCandidatesError struct{}

func (e *extractNoCandidatesError) Error() string { return "no candidates found" }
func (e *extractNoCandidatesError) ExitCode() int { return 1 }
func (e *extractNoCandidatesError) Code() string  { return "no_candidates" }
func (e *extractNoCandidatesError) Hint() string  { return "" }

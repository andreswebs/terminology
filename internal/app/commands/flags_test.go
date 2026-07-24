package commands

import (
	"slices"
	"strings"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
	urfcli "github.com/urfave/cli/v3"
)

func flagByName(flags []urfcli.Flag, name string) urfcli.Flag {
	for _, f := range flags {
		if slices.Contains(f.Names(), name) {
			return f
		}
	}
	return nil
}

func TestWriteFlags(t *testing.T) {
	flags := writeFlags("validate and preview without writing")
	if len(flags) != 3 {
		t.Fatalf("writeFlags: got %d flags, want 3", len(flags))
	}

	dr := flagByName(flags, "dry-run")
	if dr == nil {
		t.Fatal("writeFlags: missing --dry-run")
	}
	if _, ok := dr.(*urfcli.BoolFlag); !ok {
		t.Error("--dry-run should be a BoolFlag")
	}
	if !containsName(dr.Names(), "n") {
		t.Error("--dry-run should have alias -n")
	}

	tx := flagByName(flags, "transaction")
	if tx == nil {
		t.Fatal("writeFlags: missing --transaction")
	}
	if _, ok := tx.(*urfcli.BoolFlag); !ok {
		t.Error("--transaction should be a BoolFlag")
	}

	au := flagByName(flags, "author")
	if au == nil {
		t.Fatal("writeFlags: missing --author")
	}
	sf, ok := au.(*urfcli.StringFlag)
	if !ok {
		t.Fatal("--author should be a StringFlag")
	}
	if !containsName(au.Names(), "a") {
		t.Error("--author should have alias -a")
	}
	if len(sf.Sources.Chain) == 0 {
		t.Error("--author should have env var source")
	}
}

func TestReadFieldsFlag(t *testing.T) {
	f := readFieldsFlag()
	if f == nil {
		t.Fatal("readFieldsFlag returned nil")
	}
	if !containsName(f.Names(), "fields") {
		t.Error("readFieldsFlag should have name 'fields'")
	}
	if !containsName(f.Names(), "F") {
		t.Error("readFieldsFlag should have alias 'F'")
	}
}

func TestLangFlag(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		f := langFlag(false, "restrict to a language section")
		sf, ok := f.(*urfcli.StringFlag)
		if !ok {
			t.Fatal("langFlag should return a StringFlag")
		}
		if sf.Required {
			t.Error("langFlag(false, ...) should not be required")
		}
		if !containsName(f.Names(), "l") {
			t.Error("langFlag should have alias -l")
		}
	})
	t.Run("required", func(t *testing.T) {
		f := langFlag(true, "language tag")
		sf, ok := f.(*urfcli.StringFlag)
		if !ok {
			t.Fatal("langFlag should return a StringFlag")
		}
		if !sf.Required {
			t.Error("langFlag(true, ...) should be required")
		}
	})
}

func TestTermFlag(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		f := termFlag(false, "surface form of the term")
		sf, ok := f.(*urfcli.StringFlag)
		if !ok {
			t.Fatal("termFlag should return a StringFlag")
		}
		if sf.Required {
			t.Error("termFlag(false, ...) should not be required")
		}
		if !containsName(f.Names(), "t") {
			t.Error("termFlag should have alias -t")
		}
	})
	t.Run("required", func(t *testing.T) {
		f := termFlag(true, "surface form")
		sf := f.(*urfcli.StringFlag)
		if !sf.Required {
			t.Error("termFlag(true, ...) should be required")
		}
	})
}

func TestPickFlag(t *testing.T) {
	values := tbx.AdminStatus()
	f := pickFlag("status", "s", "administrative status", func() []string { return values })
	sf, ok := f.(*urfcli.StringFlag)
	if !ok {
		t.Fatal("pickFlag should return a StringFlag")
	}
	if !containsName(f.Names(), "s") {
		t.Error("pickFlag should include the alias")
	}
	if sf.Validator == nil {
		t.Fatal("pickFlag should have a validator")
	}
	if err := sf.Validator(values[0]); err != nil {
		t.Errorf("validator rejected valid value %q: %v", values[0], err)
	}
	err := sf.Validator("bogus")
	if err == nil {
		t.Fatal("validator should reject invalid value")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("error should contain 'invalid value', got: %v", err)
	}
}

func TestPickFlagNoAlias(t *testing.T) {
	f := pickFlag("grammatical-gender", "", "grammatical gender", tbx.GrammaticalGender)
	if containsName(f.Names(), "") {
		t.Error("empty alias should not appear in names")
	}
	if !containsName(f.Names(), "grammatical-gender") {
		t.Error("should have long name")
	}
}

func TestDryRunFlag(t *testing.T) {
	f := dryRunFlag("validate and preview without writing")
	bf, ok := f.(*urfcli.BoolFlag)
	if !ok {
		t.Fatal("dryRunFlag should return a BoolFlag")
	}
	if !containsName(f.Names(), "n") {
		t.Error("dryRunFlag should have alias -n")
	}
	_ = bf
}

func TestAuthorFlag(t *testing.T) {
	f := authorFlag()
	sf, ok := f.(*urfcli.StringFlag)
	if !ok {
		t.Fatal("authorFlag should return a StringFlag")
	}
	if !containsName(f.Names(), "a") {
		t.Error("authorFlag should have alias -a")
	}
	if len(sf.Sources.Chain) == 0 {
		t.Error("authorFlag should have TERMINOLOGY_AUTHOR env source")
	}
}

func TestTransactionFlag(t *testing.T) {
	f := transactionFlag()
	if _, ok := f.(*urfcli.BoolFlag); !ok {
		t.Fatal("transactionFlag should return a BoolFlag")
	}
	if !containsName(f.Names(), "transaction") {
		t.Error("transactionFlag should have name 'transaction'")
	}
}

func containsName(names []string, target string) bool {
	return slices.Contains(names, target)
}

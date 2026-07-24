package commands

import (
	"strings"

	urfcli "github.com/urfave/cli/v3"
)

func writeFlags(dryRunUsage string) []urfcli.Flag {
	return []urfcli.Flag{
		dryRunFlag(dryRunUsage),
		transactionFlag(),
		authorFlag(),
	}
}

func readFieldsFlag() urfcli.Flag {
	return &urfcli.StringFlag{Name: "fields", Aliases: []string{"F"}, Usage: "comma-separated dotted paths to include"}
}

func langFlag(required bool, usage string) urfcli.Flag {
	return &urfcli.StringFlag{Name: "lang", Aliases: []string{"l"}, Usage: usage, Required: required}
}

func termFlag(required bool, usage string) urfcli.Flag {
	return &urfcli.StringFlag{Name: "term", Aliases: []string{"t"}, Usage: usage, Required: required}
}

func pickFlag(name, alias, usage string, valuesFn func() []string) urfcli.Flag {
	allowed := valuesFn()
	set := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		set[v] = true
	}
	f := &urfcli.StringFlag{
		Name:  name,
		Usage: usage,
		Validator: func(val string) error {
			if !set[val] {
				return urfcli.Exit("invalid value "+val+"; accepted: "+strings.Join(allowed, ", "), 2)
			}
			return nil
		},
	}
	if alias != "" {
		f.Aliases = []string{alias}
	}
	return f
}

func dryRunFlag(usage string) urfcli.Flag {
	return &urfcli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: usage}
}

func transactionFlag() urfcli.Flag {
	return &urfcli.BoolFlag{Name: "transaction", Usage: "append a <transacGrp> record"}
}

func authorFlag() urfcli.Flag {
	return &urfcli.StringFlag{
		Name:    "author",
		Aliases: []string{"a"},
		Usage:   "responsibility value for the transaction record",
		Sources: urfcli.EnvVars("TERMINOLOGY_AUTHOR"),
	}
}

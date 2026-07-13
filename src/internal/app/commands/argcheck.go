package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

func argBounds(minArgs, maxArgs int) urfcli.BeforeFunc {
	return func(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
		n := cmd.Args().Len()
		if n < minArgs {
			noun := "argument"
			if minArgs > 1 {
				noun = "arguments"
			}
			return ctx, terr.Newf(
				"missing_argument", 2,
				fmt.Sprintf("see '%s --help'", cmd.FullName()),
				"%s requires %d %s, got %d", cmd.FullName(), minArgs, noun, n,
			)
		}
		if maxArgs >= 0 && n > maxArgs {
			noun := "argument"
			if maxArgs != 1 {
				noun = "arguments"
			}
			return ctx, terr.Newf(
				"excess_arguments", 2,
				fmt.Sprintf("see '%s --help'", cmd.FullName()),
				"%s accepts at most %d %s, got %d", cmd.FullName(), maxArgs, noun, n,
			)
		}
		return ctx, nil
	}
}

func chainBefore(fns ...urfcli.BeforeFunc) urfcli.BeforeFunc {
	return func(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
		for _, fn := range fns {
			var err error
			ctx, err = fn(ctx, cmd)
			if err != nil {
				return ctx, err
			}
		}
		return ctx, nil
	}
}

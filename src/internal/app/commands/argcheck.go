package commands

import (
	"context"
	"fmt"

	"github.com/andreswebs/terminology/internal/terr"
	urfcli "github.com/urfave/cli/v3"
)

func argBounds(min, max int) urfcli.BeforeFunc {
	return func(ctx context.Context, cmd *urfcli.Command) (context.Context, error) {
		n := cmd.Args().Len()
		if n < min {
			noun := "argument"
			if min > 1 {
				noun = "arguments"
			}
			return ctx, terr.Newf(
				"missing_argument", 2,
				fmt.Sprintf("see '%s --help'", cmd.FullName()),
				"%s requires %d %s, got %d", cmd.FullName(), min, noun, n,
			)
		}
		if max >= 0 && n > max {
			noun := "argument"
			if max != 1 {
				noun = "arguments"
			}
			return ctx, terr.Newf(
				"excess_arguments", 2,
				fmt.Sprintf("see '%s --help'", cmd.FullName()),
				"%s accepts at most %d %s, got %d", cmd.FullName(), max, noun, n,
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

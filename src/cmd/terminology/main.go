package main

import (
	"context"
	"os"

	"github.com/andreswebs/terminology/internal/app"
	"github.com/andreswebs/terminology/internal/output"
)

func main() {
	cmd := app.Root()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		output.EmitError(cmd.ErrWriter, cmd.String("format"), err)
		os.Exit(output.ExitCodeFor(err))
	}
}

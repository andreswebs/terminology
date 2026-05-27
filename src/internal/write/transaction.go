package write

import (
	"context"
	"time"

	"github.com/andreswebs/terminology/internal/clock"
	"github.com/andreswebs/terminology/internal/logctx"
	"github.com/andreswebs/terminology/internal/tbx"
)

func NewTransaction(ctx context.Context, author string) tbx.Transaction {
	if author == "" {
		l := logctx.From(ctx)
		l.Warn("missing author for transaction record; omitting responsibility")
	}

	return tbx.Transaction{
		Type:           "modification",
		Date:           clock.Now(ctx).Format(time.RFC3339),
		Responsibility: author,
	}
}

package linguist

import "github.com/andreswebs/terminology/internal/tbx"

func init() {
	tbx.RegisterDialect(
		tbx.DialectLinguist,
		func() tbx.Reader { return NewReader() },
		func() tbx.Writer { return NewWriter() },
	)
}

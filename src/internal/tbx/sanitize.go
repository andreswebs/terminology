package tbx

import (
	"bytes"
	"fmt"
)

func CheckDoctype(data []byte) error {
	upper := bytes.ToUpper(data)
	_, after, ok := bytes.Cut(upper, []byte("<!DOCTYPE"))
	if !ok {
		return nil
	}

	rest := after

	for i := range rest {
		switch rest[i] {
		case '>':
			return nil
		case '[':
			return ErrDangerousDoctype.Wrap(fmt.Errorf("DOCTYPE with internal subset (entity declarations) is not allowed"))
		case 'S':
			if bytes.HasPrefix(rest[i:], []byte("SYSTEM")) {
				return ErrDangerousDoctype.Wrap(fmt.Errorf("DOCTYPE with external ID (SYSTEM) is not allowed"))
			}
		case 'P':
			if bytes.HasPrefix(rest[i:], []byte("PUBLIC")) {
				return ErrDangerousDoctype.Wrap(fmt.Errorf("DOCTYPE with external ID (PUBLIC) is not allowed"))
			}
		}
	}

	return nil
}

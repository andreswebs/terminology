package write

import (
	"fmt"

	"github.com/andreswebs/terminology/internal/tbx"
)

// ConceptIndex returns the position of the concept with the given id in g, or
// ErrNotFound if no concept has that id.
func ConceptIndex(g *tbx.Glossary, id string) (int, error) {
	for i := range g.Concepts {
		if g.Concepts[i].ID == id {
			return i, nil
		}
	}
	return -1, ErrNotFound.Wrap(fmt.Errorf("concept %q not found", id))
}

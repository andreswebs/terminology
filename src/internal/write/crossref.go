package write

import "github.com/andreswebs/terminology/internal/tbx"

// CrossRefsTo returns the ids of concepts (other than id itself) that
// cross-reference id, at either the concept level or any term level.
func CrossRefsTo(g *tbx.Glossary, id string) []string {
	var refs []string
	for _, c := range g.Concepts {
		if c.ID == id {
			continue
		}
		if conceptRefs(c, id) {
			refs = append(refs, c.ID)
		}
	}
	return refs
}

func conceptRefs(c tbx.Concept, targetID string) bool {
	for _, cr := range c.CrossRefs {
		if cr.Target == targetID {
			return true
		}
	}
	for _, ls := range c.Languages {
		for _, t := range ls.Terms {
			for _, cr := range t.CrossRefs {
				if cr.Target == targetID {
					return true
				}
			}
		}
	}
	return false
}

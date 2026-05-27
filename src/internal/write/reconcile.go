package write

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/andreswebs/terminology/internal/tbx"
)

type ReconcileResult struct {
	Added     []string
	Updated   []string
	Removed   []string
	Unchanged []string
}

func Reconcile(g *tbx.Glossary, payload []tbx.Concept, prune bool) (*ReconcileResult, error) {
	return reconcile(g, payload, prune, nil)
}

func ReconcileWithTxn(ctx context.Context, g *tbx.Glossary, payload []tbx.Concept, prune bool, author string) (*ReconcileResult, error) {
	txnInfo := &txnConfig{ctx: ctx, author: author}
	return reconcile(g, payload, prune, txnInfo)
}

type txnConfig struct {
	ctx    context.Context
	author string
}

func reconcile(g *tbx.Glossary, payload []tbx.Concept, prune bool, txn *txnConfig) (*ReconcileResult, error) {
	currentIndex := make(map[string]int, len(g.Concepts))
	for i := range g.Concepts {
		currentIndex[g.Concepts[i].ID] = i
	}

	payloadIDs := make(map[string]bool, len(payload))
	var added, updated, unchanged []string

	for _, pc := range payload {
		payloadIDs[pc.ID] = true

		idx, exists := currentIndex[pc.ID]
		if !exists {
			if txn != nil {
				t := NewTransaction(txn.ctx, txn.author)
				pc.Transactions = append(pc.Transactions, t)
			}
			g.Concepts = append(g.Concepts, pc)
			currentIndex[pc.ID] = len(g.Concepts) - 1
			added = append(added, pc.ID)
			continue
		}

		equal, err := ConceptsEqual(&g.Concepts[idx], &pc)
		if err != nil {
			return nil, fmt.Errorf("comparing concept %q: %w", pc.ID, err)
		}

		if equal {
			unchanged = append(unchanged, pc.ID)
			continue
		}

		existingID := g.Concepts[idx].ID
		g.Concepts[idx] = pc
		g.Concepts[idx].ID = existingID
		if txn != nil {
			t := NewTransaction(txn.ctx, txn.author)
			g.Concepts[idx].Transactions = append(g.Concepts[idx].Transactions, t)
		}
		updated = append(updated, pc.ID)
	}

	var removed []string
	if prune {
		remaining := g.Concepts[:0]
		var toRemove []string

		for i := range g.Concepts {
			if payloadIDs[g.Concepts[i].ID] {
				remaining = append(remaining, g.Concepts[i])
			} else {
				toRemove = append(toRemove, g.Concepts[i].ID)
			}
		}

		for _, removeID := range toRemove {
			if refs := findCrossRefsToInSlice(remaining, removeID); len(refs) > 0 {
				return nil, ErrDanglingCrossref.Wrap(
					fmt.Errorf("concept(s) %s reference %q", strings.Join(refs, ", "), removeID),
				)
			}
		}

		g.Concepts = remaining
		removed = toRemove
	}

	if err := validateForApply(g); err != nil {
		return nil, err
	}

	sort.Strings(added)
	sort.Strings(updated)
	sort.Strings(removed)
	sort.Strings(unchanged)

	return &ReconcileResult{
		Added:     added,
		Updated:   updated,
		Removed:   removed,
		Unchanged: unchanged,
	}, nil
}

func findCrossRefsToInSlice(concepts []tbx.Concept, targetID string) []string {
	var refs []string
	for _, c := range concepts {
		if c.ID == targetID {
			continue
		}
		if conceptRefsTarget(c, targetID) {
			refs = append(refs, c.ID)
		}
	}
	return refs
}

func conceptRefsTarget(c tbx.Concept, targetID string) bool {
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

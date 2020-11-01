//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"github.com/markkurossi/iql/types"
)

// Grouping implements grouping for rows.
type Grouping struct {
	Children map[types.Value]*Grouping
	Rows     []*Row
}

// NewGrouping creates a new grouping object.
func NewGrouping() *Grouping {
	return &Grouping{
		Children: make(map[types.Value]*Grouping),
	}
}

// Add adds a row with the grouping key.
func (g *Grouping) Add(key []types.Value, row *Row) {
	if len(key) == 0 {
		g.Rows = append(g.Rows, row)
		return
	}

	child, ok := g.Children[key[0]]
	if !ok {
		child = NewGrouping()
		g.Children[key[0]] = child
	}
	child.Add(key[1:], row)
}

// Get gets the row groups.
func (g *Grouping) Get() [][]*Row {
	return g.get(nil)
}

func (g *Grouping) get(rows [][]*Row) [][]*Row {
	if len(g.Rows) > 0 {
		rows = append(rows, g.Rows)
	}
	for _, child := range g.Children {
		rows = child.get(rows)
	}
	return rows
}

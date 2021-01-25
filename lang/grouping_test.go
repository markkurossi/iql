//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package lang

import (
	"testing"

	"github.com/markkurossi/iql/types"
)

func TestGrouping(t *testing.T) {
	g := NewGrouping()

	key1 := []types.Value{
		types.BoolValue(true),
		types.IntValue(1),
		types.FloatValue(3.14),
		types.StringValue("Hello, world!"),
	}
	key2 := []types.Value{
		types.BoolValue(true),
		types.IntValue(1),
		types.FloatValue(3.14),
	}
	row1 := &Row{
		Data: []types.Row{
			[]types.Column{
				types.StringColumn("Column 1"),
				types.StringColumn("Column 2"),
			},
		},
	}
	row2 := &Row{
		Data: []types.Row{
			[]types.Column{
				types.StringColumn("C1"),
				types.StringColumn("C2"),
			},
		},
	}

	g.Add(key1, row1)
	g.Add(key1, row1)

	g.Add(key2, row2)

	groups := g.Get()
	if len(groups) != 2 {
		t.Errorf("unexpected groups: got %d, expected 2", len(groups))
	}
	// Shorter keys first
	if len(groups[0]) != 1 {
		t.Errorf("unexpected number of rows in group 0")
	}
	if len(groups[1]) != 2 {
		t.Errorf("unexpected number of rows in group 1")
	}
}

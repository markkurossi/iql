//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"os"
	"testing"

	"github.com/markkurossi/iql/data"
	"github.com/markkurossi/iql/query"
	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

func TestJoin(t *testing.T) {
	ref, err := data.New("data/test.html", "tbody > tr",
		[]types.ColumnSelector{
			{
				Name: types.Reference{
					Column: ".stock",
				},
				As: "Stock",
			},
			{
				Name: types.Reference{
					Column: ".price",
				},
				As: "Price",
			},
			{
				Name: types.Reference{
					Column: ".share",
				},
				As: "Weight",
			},
		})
	if err != nil {
		t.Fatalf("NewHTML failed: %s", err)
	}
	portfolio, err := data.New("data/test.csv", "noheaders",
		[]types.ColumnSelector{
			{
				Name: types.Reference{
					Column: "0",
				},
				As: "Stock",
			},
			{
				Name: types.Reference{
					Column: "1",
				},
				As: "Count",
			},
		})
	if err != nil {
		t.Fatalf("NewHTML failed: %s", err)
	}

	q := query.NewQuery(nil)

	q.Select = []query.ColumnSelector{
		{
			Expr: &query.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			As: "Stock",
		},
		{
			Expr: &query.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Price",
				},
			},
			As: "Price",
		},
		{
			Expr: &query.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Weight",
				},
			},
			As: "Weight",
		},
		{
			Expr: &query.Reference{
				Reference: types.Reference{
					Source: "portfolio",
					Column: "Count",
				},
			},
			As: "Count",
		},
	}
	q.From = []query.SourceSelector{
		{
			Source: ref,
			As:     "ref",
		},
		{
			Source: portfolio,
			As:     "portfolio",
		},
	}
	q.Where = &query.And{
		Left: &query.Binary{
			Type: query.BinNeq,
			Left: &query.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			Right: &query.Constant{
				Value: types.StringValue(""),
			},
		},
		Right: &query.Binary{
			Type: query.BinEq,
			Left: &query.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			Right: &query.Reference{
				Reference: types.Reference{
					Source: "portfolio",
					Column: "Stock",
				},
			},
		},
	}
	tab, err := types.Tabulate(q, tabulate.Unicode)
	if err != nil {
		t.Fatalf("Query failed: %s", err)
	}
	tab.Print(os.Stdout)
}

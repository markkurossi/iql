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
	"github.com/markkurossi/iql/lang"
	"github.com/markkurossi/iql/types"
	"github.com/markkurossi/tabulate"
)

func TestJoin(t *testing.T) {
	ref, err := data.New([]string{"../../data/test.html"}, "tbody > tr",
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
	portfolio, err := data.New([]string{"../../data/test.csv"}, "noheaders",
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

	q := lang.NewQuery(nil)

	q.Select = []lang.ColumnSelector{
		{
			Expr: &lang.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			As: "Stock",
		},
		{
			Expr: &lang.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Price",
				},
			},
			As: "Price",
		},
		{
			Expr: &lang.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Weight",
				},
			},
			As: "Weight",
		},
		{
			Expr: &lang.Reference{
				Reference: types.Reference{
					Source: "portfolio",
					Column: "Count",
				},
			},
			As: "Count",
		},
	}
	q.From = []lang.SourceSelector{
		{
			Source: ref,
			As:     "ref",
		},
		{
			Source: portfolio,
			As:     "portfolio",
		},
	}
	q.Where = &lang.And{
		Left: &lang.Binary{
			Type: lang.BinNeq,
			Left: &lang.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			Right: &lang.Constant{
				Value: types.StringValue(""),
			},
		},
		Right: &lang.Binary{
			Type: lang.BinEq,
			Left: &lang.Reference{
				Reference: types.Reference{
					Source: "ref",
					Column: "Stock",
				},
			},
			Right: &lang.Reference{
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
